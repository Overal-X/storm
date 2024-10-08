package storm

type WorkflowConfig struct {
	Name string `yaml:"name"`
	On   struct {
		Push        struct{} `yaml:"push"`
		PullRequest struct{} `yaml:"pull-request"`
	} `yaml:"on"`
	Jobs []Job `yaml:"jobs"`

	// Directory to run the workflow from, defaults to the current directory
	Directory string `yaml:"directory"`
}

type Job struct {
	Name   string `yaml:"name"`
	RunsOn string `yaml:"runs-on"`
	Needs  string `yaml:"needs,omitempty"`
	Steps  []Step `yaml:"steps"`
}

type Step struct {
	Name      string `yaml:"name,omitempty"`
	Run       string `yaml:"run,omitempty"`
	Directory string `yaml:"directory"`
}
