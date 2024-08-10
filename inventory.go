package main

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Inventory struct{}

func (c *Inventory) Load(file string) (*InventoryConfig, error) {
	fileContent, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	// Unmarshal the file content into the InventoryConfig struct
	config := &InventoryConfig{}
	err = yaml.Unmarshal(fileContent, &config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func NewInventory() *Inventory {
	return &Inventory{}
}
