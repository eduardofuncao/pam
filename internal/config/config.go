package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/eduardofuncao/squix/internal/db"
	"github.com/eduardofuncao/squix/internal/styles"
	"gopkg.in/yaml.v2"
)

var CfgPath = initCfgPath()
var CfgFile = filepath.Join(CfgPath, "config.yaml")

func initCfgPath() string {
	if runtime.GOOS == "windows" {
		dir, err := os.UserConfigDir()
		if err != nil {
			dir = os.ExpandEnv("$HOME/.config")
		}
		return filepath.Join(dir, "squix")
	}
	return os.ExpandEnv("$HOME/.config/squix/")
}

type KeybindingsConfig map[string][]string

type stringOrSlice []string

func (s *stringOrSlice) UnmarshalYAML(unmarshal func(any) error) error {
	var single string
	if err := unmarshal(&single); err == nil {
		*s = []string{single}
		return nil
	}
	var multi []string
	if err := unmarshal(&multi); err != nil {
		return err
	}
	*s = multi
	return nil
}

func (kc *KeybindingsConfig) UnmarshalYAML(unmarshal func(any) error) error {
	var raw map[string]stringOrSlice
	if err := unmarshal(&raw); err != nil {
		return err
	}
	result := make(map[string][]string, len(raw))
	for k, v := range raw {
		result[k] = []string(v)
	}
	*kc = result
	return nil
}

type Config struct {
	CurrentConnection  string                         `yaml:"current_connection"`
	Connections        map[string]*ConnectionYAML     `yaml:"connections"`
	QueryGroups        map[string]map[string]db.Query `yaml:"-"`
	ColorScheme        string                         `yaml:"color_scheme"`
	CustomColorScheme  *styles.ColorScheme            `yaml:"custom_colors,omitempty"`
	History            History                        `yaml:"history"`
	DefaultRowLimit    int                            `yaml:"default_row_limit"`
	DefaultColumnWidth int                            `yaml:"default_column_width"`
	UIVisibility       UIVisibility                   `yaml:"ui_visibility"`
	Keybindings        KeybindingsConfig              `yaml:"keybindings,omitempty"`
	KeyMap             *KeyMap                        `yaml:"-"`
}

type History struct {
	Size int `yaml:"size"`
}

type UIVisibility struct {
	QueryName         bool `yaml:"query_name"`
	QuerySQL          bool `yaml:"query_sql"`
	TypeDisplay       bool `yaml:"type_display"`
	KeyIcons          bool `yaml:"key_icons"`
	FooterCellContent bool `yaml:"footer_cell_content"`
	FooterStats       bool `yaml:"footer_stats"`
	FooterKeymaps     bool `yaml:"footer_keymaps"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("Creating blank config file at", CfgFile)
			cfg := &Config{
				CurrentConnection:  "",
				Connections:        make(map[string]*ConnectionYAML),
				QueryGroups:        make(map[string]map[string]db.Query),
				ColorScheme:        "default",
				History:            History{},
				DefaultRowLimit:    1000,
				DefaultColumnWidth: 15,
				UIVisibility: UIVisibility{
					QueryName:         true,
					QuerySQL:          true,
					TypeDisplay:       true,
					KeyIcons:          true,
					FooterCellContent: true,
					FooterStats:       true,
					FooterKeymaps:     true,
				},
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

	if cfg.QueryGroups == nil {
		cfg.QueryGroups = make(map[string]map[string]db.Query)
	}
	if err := cfg.loadQueryFiles(); err != nil {
		return nil, err
	}
	if cfg.MigrateQueriesToFiles() {
		if err := cfg.Save(); err != nil {
			return nil, err
		}
	}

	if cfg.DefaultColumnWidth == 0 {
		cfg.DefaultColumnWidth = 15
	}

	// DefaultRowLimit == 0 means "no limit" — left as-is so users can disable row limiting.

	// Set UI visibility defaults (all true by default)
	if !cfg.UIVisibility.QueryName && !cfg.UIVisibility.QuerySQL &&
		!cfg.UIVisibility.TypeDisplay && !cfg.UIVisibility.KeyIcons &&
		!cfg.UIVisibility.FooterCellContent && !cfg.UIVisibility.FooterStats &&
		!cfg.UIVisibility.FooterKeymaps {
		// All false means config is unset, use defaults
		cfg.UIVisibility.QueryName = true
		cfg.UIVisibility.QuerySQL = true
		cfg.UIVisibility.TypeDisplay = true
		cfg.UIVisibility.KeyIcons = true
		cfg.UIVisibility.FooterCellContent = true
		cfg.UIVisibility.FooterStats = true
		cfg.UIVisibility.FooterKeymaps = true
	}

	cfg.KeyMap = BuildKeyMap(cfg.Keybindings)

	return &cfg, nil
}

func (c *Config) Save() error {
	if err := os.MkdirAll(CfgPath, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Persist each query group to its .sql file. Disk must match memory, so a
	// now-empty group (e.g. after its last query is removed) drops its file.
	for key, queries := range c.QueryGroups {
		if len(queries) == 0 {
			if err := os.Remove(GroupFile(key)); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("failed to remove empty query group %s: %w", key, err)
			}
			continue
		}
		if err := WriteGroupFile(key, queries); err != nil {
			return fmt.Errorf("failed to write query group %s: %w", key, err)
		}
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(CfgFile, data, 0644)
}

// loadQueryFiles reads queries/<group>.sql into QueryGroups, assigning ids.
// A missing queries/ directory is not an error.
func (c *Config) loadQueryFiles() error {
	entries, err := os.ReadDir(QueriesDir())
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		key := strings.TrimSuffix(e.Name(), ".sql")
		data, err := os.ReadFile(filepath.Join(QueriesDir(), e.Name()))
		if err != nil {
			fmt.Fprintf(os.Stderr, "squix: skipping unreadable query file %s: %v\n", e.Name(), err)
			continue
		}
		queries, err := ParseGroupFile(data)
		if err != nil {
			fmt.Fprintf(os.Stderr, "squix: skipping malformed query file %s: %v\n", e.Name(), err)
			continue
		}
		AssignIDs(queries)
		c.QueryGroups[key] = queries
	}
	return nil
}
