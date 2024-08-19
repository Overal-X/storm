package storm

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type Ssh struct{}

type AuthenticateArgs struct {
	User     string
	Password string
	Host     string
	Port     int
}

func (s *Ssh) Authenticate(args AuthenticateArgs) (*ssh.Client, error) {
	sshConfig := &ssh.ClientConfig{
		User: args.User,
		Auth: []ssh.AuthMethod{
			ssh.Password(args.Password), // Use a password
		},
		// TODO: For production, use a more secure host key callback
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	if args.Password != "" {
		sshConfig.Auth = []ssh.AuthMethod{ssh.Password(args.Password)}
	} else {
		sshConfig.Auth = []ssh.AuthMethod{ssh.PublicKeys()}
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

func (s *Ssh) ExecuteCommand(client *ssh.Client, command string) (string, string, error) {
	// Create a new SSH session
	session, err := client.NewSession()
	if err != nil {
		return "", "", fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer session.Close()

	// Set up buffers to capture stdout and stderr
	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	// Run the command
	if err := session.Run(command); err != nil {
		return stdout.String(), stderr.String(), fmt.Errorf("failed to execute command: %w", err)
	}

	return stdout.String(), stderr.String(), nil
}

func NewSsh() *Ssh {
	return &Ssh{}
}
