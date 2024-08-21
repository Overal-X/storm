package storm

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type Ssh struct{}

type AuthenticateArgs struct {
	User          string
	Password      string
	Host          string
	Port          int
	PrivateSshKey string
}

func (s *Ssh) Authenticate(args AuthenticateArgs) (*ssh.Client, error) {
	signers := make([]ssh.AuthMethod, 0)

	if args.PrivateSshKey == "" && args.Password == "" {
		return nil, errors.New("ssh key or password is required")
	}

	if args.PrivateSshKey != "" {
		privateKey, err := ssh.ParsePrivateKey([]byte(args.PrivateSshKey))
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
		signers = append(signers, ssh.PublicKeys(privateKey))
	} else {
		signers = append(signers, ssh.Password(args.Password))
	}

	sshConfig := &ssh.ClientConfig{
		User: args.User,
		Auth: signers,
		// TODO: For production, use a more secure host key callback
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// Connect to the SSH server
	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", args.Host, args.Port), sshConfig)
	if err != nil {
		log.Printf("Failed to dial: %v\n", err)

		return nil, errors.Join(errors.New("ssh authentication failed"), err)
	}
	// TODO: close connection when finished

	return client, nil
}

// Copy file from local server to remote server
//
//	@example
//
//	ssh := NewSsh()
//	sshClient, err := ssh.Authenticate(AuthenticateArgs{
//		Host:     "10.211.55.12",
//		Port:     22,
//		User:     "ubuntu",
//		Password: "1234567890",
//	})
//	fmt.Println(err)
//	ssh.CopyTo(sshClient, "./from/one/place.yaml", "/to/another.yaml")
func (s *Ssh) CopyTo(client *ssh.Client, source string, destination string) error {
	// Create an SFTP client
	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		log.Printf("Failed to create SFTP client: %v\n", err)

		return err
	}

	defer sftpClient.Close()

	// Ensure the destination directory exists
	destDir := filepath.Dir(destination)
	if err := s.CreateDirectory(sftpClient, destDir); err != nil {
		return fmt.Errorf("failed to ensure destination directory exists: %w", err)
	}

	// Open the local file
	localFile, err := os.Open(source)
	if err != nil {
		log.Printf("Failed to open local file: %v\n", err)

		return err
	}
	defer localFile.Close()

	// Create the remote file
	remoteFile, err := sftpClient.Create(destination)
	if err != nil {
		log.Printf("Failed to create remote file: %v\n", err)

		return err
	}
	defer remoteFile.Close()

	// Copy the file from local to remote
	if _, err := localFile.WriteTo(remoteFile); err != nil {
		log.Printf("Failed to write file to remote server: %v\n", err)

		return err
	}

	return nil
}

func (s *Ssh) CreateDirectory(sftpClient *sftp.Client, dirPath string) error {
	// Check if the directory exists
	_, err := sftpClient.Stat(dirPath)
	if err == nil {
		// Directory exists
		return nil
	}
	if os.IsNotExist(err) {
		// Directory does not exist, create it
		if err := s.CreateDirectory(sftpClient, filepath.Dir(dirPath)); err != nil {
			return err
		}
		// Create the directory
		if err := sftpClient.Mkdir(dirPath); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dirPath, err)
		}
	} else {
		return err
	}

	return nil
}

// writerFunc is a helper that turns a callback function into an io.Writer.
func writerFunc(callback func(string)) io.Writer {
	return writerFuncImpl{callback: callback}
}

type writerFuncImpl struct {
	callback func(string)
}

func (w writerFuncImpl) Write(p []byte) (n int, err error) {
	trimmedLine := strings.TrimSpace(string(p))
	w.callback(string(trimmedLine))

	return len(p), nil
}

type ExecuteCommandArgs struct {
	Client         *ssh.Client
	Command        string
	OutputCallback func(string)
	ErrorCallback  func(string)
}

func (s *Ssh) ExecuteCommand(args ExecuteCommandArgs) (string, string, error) {
	// Create a new SSH session
	session, err := args.Client.NewSession()
	if err != nil {
		return "", "", fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer session.Close()

	// Set up pipes for stdout and stderr
	stdoutPipe, err := session.StdoutPipe()
	if err != nil {
		return "", "", fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderrPipe, err := session.StderrPipe()
	if err != nil {
		return "", "", fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	var stdoutBuf, stderrBuf bytes.Buffer

	// Create channels to signal completion of stdout and stderr streaming
	doneOut := make(chan error)
	doneErr := make(chan error)

	// Stream stdout
	go func() {
		multiWriter := io.MultiWriter(&stdoutBuf, writerFunc(args.OutputCallback))
		_, err := io.Copy(multiWriter, stdoutPipe)
		doneOut <- err
	}()

	// Stream stderr
	go func() {
		multiWriter := io.MultiWriter(&stderrBuf, writerFunc(args.ErrorCallback))
		_, err := io.Copy(multiWriter, stderrPipe)
		doneErr <- err
	}()

	// Run the command
	if err := session.Start(args.Command); err != nil {
		return "", "", fmt.Errorf("failed to start command: %w", err)
	}

	// Wait for stdout and stderr to finish streaming
	if err := <-doneOut; err != nil {
		return "", "", fmt.Errorf("error while streaming stdout: %w", err)
	}
	if err := <-doneErr; err != nil {
		return "", "", fmt.Errorf("error while streaming stderr: %w", err)
	}

	// Wait for the session to complete
	if err := session.Wait(); err != nil {
		return stdoutBuf.String(), stderrBuf.String(), fmt.Errorf("failed to execute command: %w", err)
	}

	return stdoutBuf.String(), stderrBuf.String(), nil
}

func NewSsh() *Ssh {
	return &Ssh{}
}
