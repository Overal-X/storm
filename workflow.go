package storm

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Workflow struct{}

func (w *Workflow) Load(file string) (*WorkflowConfig, error) {
	workflow := WorkflowConfig{}

	fileContent, _ := os.ReadFile(file)
	yaml.Unmarshal([]byte(fileContent), &workflow)

	return &workflow, nil
}

func (w *Workflow) Dump(content WorkflowConfig) (*string, error) {
	out, err := yaml.Marshal(&content)
	outStr := string(out)

	return &outStr, err
}

type State struct {
	IsSuccessful bool
	IsCompleted  bool
}

type JobState map[string]State

const (
	StepOutputTypePlain = iota + 1
	StepOutputTypeStruct
	StepOutputTypeJson
)

type StepOutputPlain string

type WorkflowStepOutputStruct struct {
	// workflow step path
	// 	example; `build.installing curl`, means 👇
	// 	- name: build
	// 		steps:
	// 		- name: installing curl
	// 			run: sudo apt install -y curl
	Path    string
	Command string
	Message string
}

type WorkflowRunArgs struct {
	File           *string
	Config         *WorkflowConfig
	Callback       func(interface{})
	StepOutputType int
}

type WorkflowRunOptions func(*WorkflowRunArgs)

func (w *Workflow) WorkflowWithFile(file string) WorkflowRunOptions {
	return func(wra *WorkflowRunArgs) {
		wra.File = &file
	}
}

func (w *Workflow) WorkflowWithConfig(config WorkflowConfig) WorkflowRunOptions {
	return func(wra *WorkflowRunArgs) {
		wra.Config = &config
	}
}

func (w *Workflow) WorkflowWithCallback(callback func(interface{}), sot int) WorkflowRunOptions {
	return func(wra *WorkflowRunArgs) {
		wra.Callback = callback
		wra.StepOutputType = sot
	}
}

func (w *Workflow) Run(opts ...WorkflowRunOptions) error {
	args := WorkflowRunArgs{
		StepOutputType: StepOutputTypePlain,
		Callback:       func(sos interface{}) {},
	}

	for _, opt := range opts {
		opt(&args)
	}

	if args.File == nil && args.Config == nil {
		return errors.New("either file or config must be specified to run a workflow")
	}

	if args.File != nil && args.Config == nil {
		_config, err := w.Load(*args.File)
		if err != nil {
			return err
		}

		args.Config = _config
	}

	jobState := make(JobState, 0)

	for _, job := range args.Config.Jobs {
		jobState[job.Name] = State{IsSuccessful: true, IsCompleted: true}

		// TODO: handle error for when `job.Needs` is not found in `jobState`; aka, don't exist
		if job.Needs != "" && (!jobState[job.Needs].IsCompleted || !jobState[job.Needs].IsSuccessful) {
			err := fmt.Errorf("> dependencies error, %s job failed", job.Needs)
			fmt.Println(err)

			return err
		}

		start := time.Now()

		if args.StepOutputType == StepOutputTypePlain {
			fmt.Printf("[%s]\n", job.Name)
		}

		err := func() error {
			for _, step := range job.Steps {
				if args.StepOutputType == StepOutputTypePlain {
					fmt.Printf("-> %s\n", step.Name)
					fmt.Printf("$ %s \n", step.Run)
				}

				callback := func(s string) {
					switch args.StepOutputType {
					case StepOutputTypePlain:
						fmt.Println("> ", s)
					case StepOutputTypeStruct:
						args.Callback(WorkflowStepOutputStruct{
							Path:    fmt.Sprintf("%s.%s", job.Name, step.Name),
							Command: step.Run,
							Message: s,
						})
					case StepOutputTypeJson:
						payload := WorkflowStepOutputStruct{
							Path:    fmt.Sprintf("%s.%s", job.Name, step.Name),
							Command: step.Run,
							Message: s,
						}
						payloadString, err := json.Marshal(&payload)
						if err != nil {
							fmt.Println("could not marshel workflow payload to json. reason: ", err)
							break
						}

						args.Callback(string(payloadString))
					}
				}

				err := w.Execute(ExecuteArgs{
					Directory:      args.Config.Directory,
					Command:        step.Run,
					OutputCallback: callback,
					ErrorCallback:  callback,
				})
				if err != nil {
					return err
				}
			}

			return nil
		}()

		end := time.Now()
		duration := end.Sub(start)

		switch args.StepOutputType {
		case StepOutputTypePlain:
			fmt.Printf("Took %fs to run.\n\n", duration.Seconds())
		case StepOutputTypeStruct:
			args.Callback(WorkflowStepOutputStruct{
				Path:    "__builtin__.TimeTaken",
				Command: "TimeTaken",
				Message: fmt.Sprintf("%fs", duration.Seconds()),
			})
		}

		if err != nil {
			state := jobState[job.Name]
			state.IsSuccessful = false
			state.IsCompleted = false

			jobState[job.Name] = state
		}
	}

	return nil
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

	currentDirectory, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("cannot get current directory %w", err)
	}

	err = os.Chdir(args.Directory)
	if err != nil {
		return fmt.Errorf("cannot change directory %w", err)
	}

	defer os.Chdir(currentDirectory)

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
