package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

var CfgPath = os.ExpandEnv("$HOME/.config/pam/")
var CfgFile = filepath.Join(CfgPath, "config.yaml")

type Config struct {
	CurrentConnection string                    `yaml:"current_connection"`
	Connections       map[string]*ConnectionYAML `yaml:"connections"`
	Style             Style                     `yaml:"style"`
	History           History                   `yaml:"history"`
}

type Style struct {
	Accent string `yaml:"accent_color"`
}

type History struct {
	Size int `yaml:"size"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("Creating blank config file at", CfgFile)
			cfg := &Config{
				CurrentConnection: "",
				Connections:       make(map[string]*ConnectionYAML),
				Style:             Style{},
				History:           History{},
			}
			err := cfg.Save()
			if err != nil {
				return nil, err
			}
			return cfg, nil
		}
		return nil, err
	}
	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *Config) Save() error {
	err := os.MkdirAll(CfgPath, 0755)
	if err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(CfgFile, data, 0644)
}
