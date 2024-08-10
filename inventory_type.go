package main

type InventoryConfig struct {
	Servers map[string]Server `yaml:"servers"`
}

type Server struct {
	// IP Address or Domain
	Host string `yaml:"host"`

	// SSH Port, defaults to 22
	Port         int    `yaml:"port,omitempty"`
	User         string `yaml:"user"`
	SudoPassword string `yaml:"sudo-pass"`
	SshPassword  string `yaml:"ssh-pass"`
}
