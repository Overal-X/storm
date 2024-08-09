package service

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
	"storm.formatio.org/dto"
)

type WorkflowInterface interface {
	Run()
}

type workflow struct {
	file string
}

func (w *workflow) Run() {
	fileContent, _ := os.ReadFile(w.file)
	workflow := dto.Workflow{}

	yaml.Unmarshal([]byte(fileContent), &workflow)
	for name, job := range workflow.Jobs {
		start := time.Now()

		fmt.Printf("[%s]\n", name)
		for _, step := range job.Steps {
			fmt.Printf("-> %s\n", step.Name)
			fmt.Printf("$ %s \n", step.Run)
			output, err := w.Execute(step.Run)

			fmt.Printf("> %s \n", *output)
			if err != nil {
				os.Exit(1)

				break
			}
		}

		end := time.Now()
		duration := end.Sub(start)

		fmt.Printf("Took %fs to run.\n\n", duration.Seconds())
	}
}

func (*workflow) Execute(command string) (*string, error) {
	result := make([]string, 0)
	command = strings.Trim(command, "")

	chainnedCommand := strings.Split(command, " && ")

	for _, command := range chainnedCommand {
		splittedCommand := strings.Split(command, " ")
		name := splittedCommand[0]
		args := make([]string, 0)
		if len(splittedCommand) > 1 {
			args = splittedCommand[1:]
		}

		out, err := exec.Command(name, args...).Output()
		if err != nil {
			return nil, err
		}

		result = append(result, string(out))
	}

	output := strings.Join(result, "\n")

	return &output, nil
}

func WorkflowService(file string) WorkflowInterface {
	return &workflow{file: file}
}