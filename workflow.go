package storm

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Workflow struct{}

func (w *Workflow) Load(file string) (*WorkflowConfig, error) {
	// dir, err := os.Getwd()
	// if err != nil {
	// 	return nil, errors.Join(errors.New("could not get current directory"), err)
	// }

	workflow := WorkflowConfig{
		// Directory: dir,
	}

	fileContent, _ := os.ReadFile(file)
	yaml.Unmarshal([]byte(fileContent), &workflow)

	return &workflow, nil
}

func (w *Workflow) Dump(content WorkflowConfig) (*string, error) {
	out, err := yaml.Marshal(&content)
	outStr := string(out)

	return &outStr, err
}

func (w *Workflow) RunWithConfig(workflow WorkflowConfig) error {
	for _, job := range workflow.Jobs {
		start := time.Now()

		fmt.Printf("[%s]\n", job.Name)
		for _, step := range job.Steps {
			fmt.Printf("-> %s\n", step.Name)
			fmt.Printf("$ %s \n", step.Run)
			err := w.Execute(ExecuteArgs{
				Directory:      workflow.Directory,
				Command:        step.Run,
				OutputCallback: func(s string) { fmt.Println("> ", s) },
				ErrorCallback:  func(s string) { fmt.Println("> ", s) },
			})

			if err != nil {
				fmt.Printf("> %s \n", err)
			}
		}

		end := time.Now()
		duration := end.Sub(start)

		fmt.Printf("Took %fs to run.\n\n", duration.Seconds())
	}

	return nil
}

func (w *Workflow) RunWithFile(file string) error {
	wc, err := w.Load(file)
	if err != nil {
		return err
	}

	return w.RunWithConfig(*wc)
}

type ExecuteArgs struct {
	Directory      string
	Command        string
	OutputCallback func(string)
	ErrorCallback  func(string)
}

func (w *Workflow) Execute(args ExecuteArgs) error {
	// Trim any leading/trailing whitespace
	command := strings.TrimSpace(args.Command)

	err := os.Chdir(args.Directory)
	if err != nil {
		return fmt.Errorf("cannot change directory %w", err)
	}

	currentCmd := exec.Command("/bin/bash", "-c", command)

	stdoutPipe, err := currentCmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("error creating stdout pipe: %w", err)
	}

	stderrPipe, err := currentCmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("error creating stderr pipe: %w", err)
	}

	// Start the command
	if err := currentCmd.Start(); err != nil {
		return fmt.Errorf("error starting command: %w", err)
	}

	// Stream stdout to the output callback
	go func() {
		scanner := bufio.NewScanner(stdoutPipe)
		for scanner.Scan() {
			args.OutputCallback(scanner.Text())
		}
	}()

	// Stream stderr to the error callback
	go func() {
		scanner := bufio.NewScanner(stderrPipe)
		for scanner.Scan() {
			args.ErrorCallback(scanner.Text())
		}
	}()

	// Wait for the command to finish
	if err := currentCmd.Wait(); err != nil {
		return fmt.Errorf("error waiting for command: %w", err)
	}

	return nil
}

func NewWorkflow() *Workflow {
	return &Workflow{}
}
