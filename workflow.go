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
	fileContent, _ := os.ReadFile(file)
	workflow := WorkflowConfig{}

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
			err := w.Execute(step.Run, func(s string) {
				fmt.Printf("> %s \n", s)
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

func (w *Workflow) Execute(command string, outputCallback func(string)) error {
	// Trim any leading/trailing whitespace
	command = strings.TrimSpace(command)

	// Use `/bin/bash -c` to execute the command with pipes
	currentCmd := exec.Command("/bin/bash", "-c", command)

	stdoutPipe, err := currentCmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("error creating stdout pipe: %w", err)
	}

	// Start the command
	if err := currentCmd.Start(); err != nil {
		return fmt.Errorf("error starting command: %w", err)
	}

	// Stream the output to the callback
	go func() {
		scanner := bufio.NewScanner(stdoutPipe)
		for scanner.Scan() {
			outputCallback(scanner.Text())
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
