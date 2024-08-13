package storm

import (
	"errors"
	"fmt"
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

func (a *Agent) RunWithConfigs(inventory InventoryConfig, workflow WorkflowConfig) error {
	return nil
}

func (a *Agent) RunWithFiles(inventory string, workflow string) error {
	ic, err := a.inventory.Load(inventory)
	if err != nil {
		return err
	}

	for name, server := range ic.Servers {
		fmt.Printf("Server: [%s]\n", name)

		sshClient, err := a.ssh.Authenticate(AuthenticateArgs{
			Host:     server.Host,
			Port:     server.Port,
			User:     server.User,
			Password: server.SshPassword,
		})
		if err != nil {
			return errors.Join(err, errors.New("authentication failed"))
		}

		splittedWorkflowPath := strings.Split(workflow, "/")
		filename := splittedWorkflowPath[len(splittedWorkflowPath)-1]
		destinationFilePath := fmt.Sprintf("/home/%s/%s", server.User, filename)

		a.ssh.CopyTo(sshClient, workflow, destinationFilePath)

		output, outputErr, err := a.ssh.ExecuteCommand(sshClient, fmt.Sprintf("~/.storm/bin/storm run %s", destinationFilePath))
		if err != nil {
			return errors.Join(err, errors.New("could not run workflow"))
		}

		fmt.Println(output, outputErr)
	}

	return nil
}

// This is meant for testing locally or in CI
func (a *Agent) InstallDev(inventory string) error {
	fmt.Println("dev installation ...")

	// TODO: use host arch to build and install agent
	// TODO: curl command to install on any host

	os.Setenv("GOOS", "linux")
	os.Setenv("GOARCH", "arm64")

	_, err := exec.Command("go", "build", "-o", "./storm").Output()
	if err != nil {
		return errors.New("build failed; could not build storm")
	}
	defer os.Remove("./storm")

	ic, err := a.inventory.Load(inventory)
	if err != nil {
		return err
	}

	for name, server := range ic.Servers {
		fmt.Printf("Server: [%s]\n", name)

		sshClient, err := a.ssh.Authenticate(AuthenticateArgs{
			Host:     server.Host,
			Port:     server.Port,
			User:     server.User,
			Password: server.SshPassword,
		})
		if err != nil {
			return err
		}

		a.ssh.CopyTo(sshClient, "./storm", fmt.Sprintf("/home/%s/.storm/bin/storm", server.User))

		_, _, err = a.ssh.ExecuteCommand(sshClient, "which storm")
		if err != nil {
			fmt.Println("Installing storm on server ...")

			_, _, err = a.ssh.ExecuteCommand(
				sshClient,
				fmt.Sprintf("echo 'export PATH=/home/%s/.storm/bin:$PATH' >> ~/.bashrc && source ~/.bashrc", server.User),
			)
			if err != nil {
				return err
			}
			_, _, err = a.ssh.ExecuteCommand(sshClient, "chmod +x ~/.storm/bin/storm")
			if err != nil {
				return err
			}

			fmt.Println("Storm is Ready!")
		}
	}

	return nil
}

func (a *Agent) InstallProd(inventory string) error {
	fmt.Println("production installation ...")

	ic, err := a.inventory.Load(inventory)
	if err != nil {
		return err
	}

	for name, server := range ic.Servers {
		fmt.Printf("Server: [%s]\n", name)

		sshClient, err := a.ssh.Authenticate(AuthenticateArgs{
			Host:     server.Host,
			Port:     server.Port,
			User:     server.User,
			Password: server.SshPassword,
		})
		if err != nil {
			return err
		}

		platform := strings.Split(runtime.GOOS, "/")[0]
		fmt.Printf("Installing storm on %s server ... ", platform)

		switch platform {
		case "windows":
			_, _, err := a.ssh.ExecuteCommand(
				sshClient,
				"powershell -c irm https://raw.githubusercontent.com/Overal-X/formatio.storm/main/scripts/install.sh | iex",
			)
			if err != nil {
				return errors.Join(err, errors.New("build failed; could not install storm"))
			}
		case "linux":
		case "darwin":
			_, _, err = a.ssh.ExecuteCommand(
				sshClient,
				"curl -fsSL https://raw.githubusercontent.com/Overal-X/formatio.storm/main/scripts/install.sh | bash",
			)
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

func (a *Agent) Install(inventory string, mode string) error {
	switch mode {
	case "dev":
		return a.InstallDev(inventory)
	case "prod":
		return a.InstallProd(inventory)
	default:
		return errors.New("installation mode not supported")
	}
}

func (a *Agent) Uninstall(inventory string) error {
	ic, err := a.inventory.Load(inventory)
	if err != nil {
		return err
	}

	for name, server := range ic.Servers {
		fmt.Printf("Server: [%s]\n", name)

		sshClient, err := a.ssh.Authenticate(AuthenticateArgs{
			Host:     server.Host,
			Port:     server.Port,
			User:     server.User,
			Password: server.SshPassword,
		})
		if err != nil {
			return err
		}

		_, _, err = a.ssh.ExecuteCommand(sshClient, "which storm")
		if err == nil {
			fmt.Println("Storm is not installed.")

			return nil
		}

		fmt.Println("Removing storm from server ...")

		_, _, err = a.ssh.ExecuteCommand(
			sshClient,
			fmt.Sprintf(`
					sed -i '/export PATH=\/home\/%s\/.storm\/bin:$PATH/d' ~/.bashrc && \
					source ~/.bashrc
			`, server.User),
		)
		if err != nil {
			return err
		}

		_, _, err = a.ssh.ExecuteCommand(sshClient, "rm -rf ~/.storm/")
		if err != nil {
			return errors.New("cannot remove ~/.storm/")
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
