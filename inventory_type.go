package storm

import (
	"fmt"
	"os"
)

type InventoryConfig struct {
	Servers []Server `yaml:"servers"`
}

type Server struct {
	Name string `yaml:"name"`

	// IP Address or Domain
	Host string `yaml:"host"`

	// SSH Port, defaults to 22
	Port         int    `yaml:"port,omitempty"`
	User         string `yaml:"user"`
	SudoPassword string `yaml:"sudo-pass"`
	SshPassword  string `yaml:"ssh-pass"`

	// File path to the SSH private key
	PrivateSshKey string `yaml:"private-ssh-key"`
}

// Custom UnmarshalYAML to read the private SSH key file
func (s *Server) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type rawServer Server // Create a new type to avoid recursion
	raw := rawServer{
		Port: 22, // Set default SSH port
	}

	if err := unmarshal(&raw); err != nil {
		return err
	}

	if raw.PrivateSshKey != "" {
		// Check if the file exists
		if _, err := os.Stat(raw.PrivateSshKey); os.IsNotExist(err) {
			err = fmt.Errorf("SSH private key file %s does not exist", raw.PrivateSshKey)

			fmt.Println(err)

			return err
		}

		// Now read the private SSH key file content
		keyContent, err := os.ReadFile(raw.PrivateSshKey)
		if err != nil {
			err = fmt.Errorf("failed to read SSH private key file %s: %w", raw.PrivateSshKey, err)

			fmt.Println(err)

			return err
		}

		raw.PrivateSshKey = string(keyContent)
	}

	// Assign unmarshaled and processed values back to the original struct
	*s = Server(raw)

	return nil
}
