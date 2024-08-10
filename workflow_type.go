package main

type WorkflowConfig struct {
	Name string `yaml:"name"`
	On   struct {
		Push        struct{} `yaml:"push"`
		PullRequest struct{} `yaml:"pull-request"`
	} `yaml:"on"`
	Jobs map[string]Job `yaml:"jobs"`
}

type Job struct {
	RunsOn string `yaml:"runs-on"`
	Needs  string `yaml:"needs,omitempty"`
	Steps  []Step `yaml:"steps"`
}

type Step struct {
	Name string `yaml:"name,omitempty"`
	Run  string `yaml:"run,omitempty"`
}
