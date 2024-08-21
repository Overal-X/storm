package storm

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

type Agent struct {
	// ablyClient *AblyClient

	inventory *Inventory
	workflow  *Workflow
	ssh       *Ssh
}

type RunArgs struct {
	Wf string
	If string

	Wc WorkflowConfig
	Ic InventoryConfig
}

type RunOption func(*RunArgs)

func (a *Agent) WithConfigs(w WorkflowConfig, i InventoryConfig) RunOption {
	return func(ra *RunArgs) {
		ra.Wc = w
		ra.Ic = i
	}
}

func (a *Agent) WithFiles(w string, i string) RunOption {
	return func(ra *RunArgs) {
		ra.Wf = w
		ra.If = i
	}
}

func (a *Agent) Run(opts ...RunOption) error {
	var wc *WorkflowConfig
	var ic *InventoryConfig
	var args RunArgs

	for _, opt := range opts {
		opt(&args)
	}

	if args.Wf != "" && args.If != "" {
		_wc, err := a.workflow.Load(args.Wf)
		if err != nil {
			return err
		}
		wc = _wc

		_ic, err := a.inventory.Load(args.If)
		if err != nil {
			return err
		}
		ic = _ic
	} else {
		wc = &args.Wc
		ic = &args.Ic
	}

	if wc == nil && ic == nil {
		return errors.New("invalid inventory and workflow configurations")
	}

	for _, server := range ic.Servers {
		fmt.Printf("Server: [%s]\n", server.Name)

		sshClient, err := a.ssh.Authenticate(AuthenticateArgs{
			Host:          server.Host,
			Port:          server.Port,
			User:          server.User,
			Password:      server.SshPassword,
			PrivateSshKey: server.PrivateSshKey,
		})
		if err != nil {
			return errors.Join(err, errors.New("authentication failed"))
		}

		destinationFilePath := fmt.Sprintf("/home/%s/workflow.yaml", server.User)
		content, err := a.workflow.Dump(*wc)
		if err != nil {
			log.Println(err)

			return errors.Join(errors.New("could dump workflow config"), err)
		}

		_, _, err = a.ssh.ExecuteCommand(ExecuteCommandArgs{
			Client:         sshClient,
			Command:        fmt.Sprintf("echo '%s' > %s", *content, destinationFilePath),
			OutputCallback: func(s string) {},
			ErrorCallback:  func(s string) { fmt.Println("> ", s) },
		})
		if err != nil {
			return errors.Join(errors.New("could generate workflow file"), err)
		}

		_, _, err = a.ssh.ExecuteCommand(ExecuteCommandArgs{
			Client:         sshClient,
			Command:        fmt.Sprintf("~/.storm/bin/storm run %s", destinationFilePath),
			OutputCallback: func(s string) { fmt.Println(s) },
			ErrorCallback:  func(s string) { fmt.Println(s) },
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// This is meant for testing locally or in CI
func (a *Agent) InstallDev(ic InventoryConfig) error {
	os.Setenv("GOOS", "linux")
	os.Setenv("GOARCH", "arm64")

	_, err := exec.Command("go", "build", "-o", "storm", "./cmd").Output()
	if err != nil {
		return errors.Join(errors.New("build failed; could not build storm"), err)
	}

	defer os.Remove("./storm")

	for _, server := range ic.Servers {
		fmt.Printf("Server: [%s]\n", server.Name)

		sshClient, err := a.ssh.Authenticate(AuthenticateArgs{
			Host:          server.Host,
			Port:          server.Port,
			User:          server.User,
			Password:      server.SshPassword,
			PrivateSshKey: server.PrivateSshKey,
		})
		if err != nil {
			return err
		}

		fmt.Print("Installing storm on server ... ")

		err = a.ssh.CopyTo(sshClient, "./storm", fmt.Sprintf("/home/%s/.storm/bin/storm", server.User))
		if err != nil {
			fmt.Print(errors.Join(errors.New("ssh can't copy file"), err))

			return err
		}

		_, _, err = a.ssh.ExecuteCommand(ExecuteCommandArgs{
			Client:         sshClient,
			Command:        "chmod +x ~/.storm/bin/storm",
			OutputCallback: func(s string) {},
			ErrorCallback:  func(s string) { fmt.Println("> ", s) },
		})
		if err != nil {
			return err
		}

		fmt.Println("Storm is Ready!")
	}

	return nil
}

func (a *Agent) InstallProd(ic InventoryConfig) error {
	for _, server := range ic.Servers {
		fmt.Printf("Server: [%s]\n", server.Name)

		sshClient, err := a.ssh.Authenticate(AuthenticateArgs{
			Host:          server.Host,
			Port:          server.Port,
			User:          server.User,
			Password:      server.SshPassword,
			PrivateSshKey: server.PrivateSshKey,
		})
		if err != nil {
			return err
		}

		platform := strings.Split(runtime.GOOS, "/")[0]
		fmt.Printf("Installing storm on %s server ... \n", platform)

		switch platform {
		case "windows":
			_, _, err := a.ssh.ExecuteCommand(ExecuteCommandArgs{
				Client:         sshClient,
				Command:        "powershell -c irm https://raw.githubusercontent.com/Overal-X/formatio.storm/main/scripts/install.sh | iex",
				OutputCallback: func(s string) {},
				ErrorCallback:  func(s string) {},
			})
			if err != nil {
				return errors.Join(err, errors.New("build failed; could not install storm"))
			}
		case "linux":
		case "darwin":
			_, _, err := a.ssh.ExecuteCommand(ExecuteCommandArgs{
				Client:         sshClient,
				Command:        "curl -fsSL https://raw.githubusercontent.com/Overal-X/formatio.storm/main/scripts/install.sh | bash",
				OutputCallback: func(s string) { fmt.Println("> ", s) },
				ErrorCallback:  func(s string) { fmt.Println("> ", s) },
			})
			if err != nil {
				return errors.Join(err, errors.New("build failed; could not install storm"))
			}
		default:
			return errors.New("platform not supported")
		}

		fmt.Println("Storm is Ready!")
	}

	return nil
}

type InstallArgs struct {
	If string
	Ic InventoryConfig

	// Installation mode; options are `dev` or `prod`
	Mode string
}

func (a *Agent) Install(args InstallArgs) error {
	var ic *InventoryConfig = &args.Ic

	if args.If != "" {
		_ic, err := a.inventory.Load(args.If)
		if err != nil {
			return err
		}
		ic = _ic
	}

	switch args.Mode {
	case "dev":
		return a.InstallDev(*ic)
	case "prod":
		return a.InstallProd(*ic)
	default:
		return errors.New("installation mode not supported")
	}
}

type UninstallArgs struct {
	If string
	Ic InventoryConfig
}

func (a *Agent) Uninstall(args UninstallArgs) error {
	var ic *InventoryConfig = &args.Ic

	if args.If != "" {
		_ic, err := a.inventory.Load(args.If)
		if err != nil {
			return err
		}

		ic = _ic
	}

	for _, server := range ic.Servers {
		fmt.Printf("Server: [%s]\n", server.Name)

		sshClient, err := a.ssh.Authenticate(AuthenticateArgs{
			Host:          server.Host,
			Port:          server.Port,
			User:          server.User,
			Password:      server.SshPassword,
			PrivateSshKey: server.PrivateSshKey,
		})
		if err != nil {
			return err
		}

		_, _, err = a.ssh.ExecuteCommand(ExecuteCommandArgs{
			Client:         sshClient,
			Command:        "which ~/.storm/bin/storm",
			OutputCallback: func(s string) {},
			ErrorCallback:  func(s string) {},
		})
		if err != nil {
			fmt.Println("Storm is not installed.")

			continue
		}

		fmt.Println("Removing storm from server ... ")

		_, _, err = a.ssh.ExecuteCommand(ExecuteCommandArgs{
			Client:         sshClient,
			Command:        "rm -rf ~/.storm/",
			OutputCallback: func(s string) {},
			ErrorCallback:  func(s string) {},
		})
		if err != nil {
			return errors.Join(errors.New("cannot remove ~/.storm/"), err)
		}

		fmt.Println("Storm has been removed (:")
	}

	return nil
}

func NewAgent() *Agent {
	return &Agent{
		// ablyClient: NewAblyClient(),
		workflow:  NewWorkflow(),
		inventory: NewInventory(),
		ssh:       NewSsh(),
	}
}
