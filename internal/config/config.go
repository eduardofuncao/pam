package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

const (
	configDirPath     = "$HOME/.config/pam/"
	configFileName    = "config.yaml"
	dirPermissions    = 0755
	filePermissions   = 0644
	msgCreatingConfig = "Creating blank config file at"
)

var CfgPath = os.ExpandEnv(configDirPath)
var CfgFile = filepath.Join(CfgPath, configFileName)

type Config struct {
	CurrentConnection string                    `yaml:"current_connection"`
	Connections       map[string]ConnectionYAML `yaml:"connections"`
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
			fmt.Println(msgCreatingConfig, CfgFile)
			cfg := &Config{
				CurrentConnection: "",
				Connections:       make(map[string]ConnectionYAML),
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
	err := os.MkdirAll(CfgPath, dirPermissions)
	if err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(CfgFile, data, filePermissions)
}
