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

	// Defaults provides global fallbacks applied to every job step, such as
	// the shell and directory, unless the step overrides them.
	Defaults Defaults `yaml:"defaults,omitempty"`
}

type Defaults struct {
	// Directory to run steps from, overrides the workflow-level Directory
	// and is itself overridden by a step's own Directory.
	Directory string `yaml:"directory,omitempty"`

	// Shell to run steps with, overridden by a step's own Shell.
	Shell string `yaml:"shell,omitempty"`
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
	Shell     string `yaml:"shell"`
	Directory string `yaml:"directory"`
}
