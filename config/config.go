package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Port              int      `yaml:"port"`
	MaxTasks          int      `yaml:"max_tasks"`
	MaxFilesPerTask   int      `yaml:"max_files_per_task"`
	AllowedExtensions []string `yaml:"allowed_extensions"`
}

func LoadConfig(filename string) (*Config, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("error opening config file: %w", err)
	}
	defer f.Close()

	var cfg Config
	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&cfg)
	if err != nil {
		return nil, fmt.Errorf("error decoding config file: %w", err)
	}

	// Set defaults if not provided
	if cfg.Port == 0 {
		cfg.Port = 8080
	}
	if cfg.MaxTasks == 0 {
		cfg.MaxTasks = 3 // Default value
	}
	if cfg.MaxFilesPerTask == 0 {
		cfg.MaxFilesPerTask = 3
	}

	return &cfg, nil
}
