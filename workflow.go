package storm

import (
	"bytes"
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

func (w *Workflow) RunWithConfig(workflow WorkflowConfig) error {
	for _, job := range workflow.Jobs {
		start := time.Now()

		fmt.Printf("[%s]\n", job.Name)
		for _, step := range job.Steps {
			fmt.Printf("-> %s\n", step.Name)
			fmt.Printf("$ %s \n", step.Run)
			output, err := w.Execute(step.Run)

			if err != nil {
				fmt.Printf("> %s \n", err)
			} else {
				fmt.Printf("> %s \n", output)
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

func (w *Workflow) Execute(command string) (string, error) {
	// Trim any leading/trailing whitespace
	command = strings.TrimSpace(command)

	// Split commands by "&&" to execute them in sequence
	commands := strings.Split(command, " && ")
	var finalOutput strings.Builder

	for i, cmd := range commands {
		cmd = strings.TrimSpace(cmd)
		if cmd == "" {
			continue
		}

		// Handle piping within the command
		pipeCommands := strings.Split(cmd, "|")

		var err error
		var lastOutput *bytes.Buffer

		for j, pipeCmd := range pipeCommands {
			pipeCmd = strings.TrimSpace(pipeCmd)
			if pipeCmd == "" {
				continue
			}

			parts := strings.Fields(pipeCmd)
			name := parts[0]
			args := parts[1:]

			currentCmd := exec.Command(name, args...)
			if j > 0 {
				// Set the previous command's output as the input of the current command
				currentCmd.Stdin = lastOutput
			}
			if j < len(pipeCommands)-1 {
				// For all but the last command in the pipeline, capture the output
				lastOutput = &bytes.Buffer{}
				currentCmd.Stdout = lastOutput
			} else {
				// For the last command, write output to finalOutput
				currentCmd.Stdout = &finalOutput
			}

			if err = currentCmd.Run(); err != nil {
				return "", fmt.Errorf("error executing command: %w", err)
			}
		}

		// Add a newline if it's not the last command
		if i < len(commands)-1 {
			finalOutput.WriteString("\n")
		}
	}

	return finalOutput.String(), nil
}

func NewWorkflow() *Workflow {
	return &Workflow{}
}
