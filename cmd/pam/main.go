package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"

	"gopkg.in/yaml.v2"

	"github.com/eduardofuncao/pam/internal/table"
)

var cfgPath = os.ExpandEnv("$HOME/.config/pam/config.yaml")

func main() {
	cfg, err := loadConfig(cfgPath)
	if err != nil {
		log.Fatal("Could not load config file", err)
	}

	if len(os.Args) < 2 {
		fmt.Println("Usage:")
		fmt.Println("pam create <name> <db-type> <connection-string>")
		fmt.Println("pam switch <db-name>")
		fmt.Println("pam add <query-name> <query>")
		fmt.Println("pam query <query-name>")
		fmt.Println("pam get <db-type> <connection-string> <sql-query>")
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {

	case "create":
		if len(os.Args) < 4 {
			log.Fatal("Usage: pam create <name> <db-type> <connection-string>")
		}
		cfg.Current.Name = os.Args[2]
		cfg.Current.DBType = os.Args[3]
		cfg.Current.DBConnectionString = os.Args[4]

		cfg.Connections[cfg.Current.Name] = Connection{
			DBType:             os.Args[3],
			DBConnectionString: os.Args[4],
			Queries:            make(map[string]string),
		}

		SaveConfig(cfgPath, cfg)

	case "switch":
		if len(os.Args) < 3 {
			log.Fatal("Usage: pam switch <db-name>")
		}
		cfg.Current.Name = os.Args[2]
		cfg.Current.DBType = cfg.Connections[cfg.Current.Name].DBType
		cfg.Current.DBConnectionString = cfg.Connections[cfg.Current.Name].DBConnectionString
		SaveConfig(cfgPath, cfg)

	case "add":
		if len(os.Args) < 3 {
			log.Fatal("Usage: pam add <query-name> <query>")
		}
		//add connection to cfg if it does not exist
		_, ok := cfg.Connections[cfg.Current.Name]
		if !ok {
			cfg.Connections[cfg.Current.Name] = Connection{}
		}
		cfg.Connections[cfg.Current.Name].Queries[os.Args[2]] = os.Args[3]
		SaveConfig(cfgPath, cfg)

	case "query":
		if len(os.Args) < 3 {
			log.Fatal("Usage:pam query <query-name>")
		}
		dbType := cfg.Current.DBType
		connStr := cfg.Current.DBConnectionString
		query := cfg.Connections[cfg.Current.Name].Queries[os.Args[2]]
		queryDB(dbType, connStr, query)

	case "list":
		var objectType string
		if len(os.Args) < 3 {
			objectType = ""
		} else {
			objectType = os.Args[2]
		}

		switch objectType {
		case "connections":
			for name, connection := range cfg.Connections {
				fmt.Printf("- %s (%s)\n", name, connection.DBConnectionString)
			}

		case "", "queries":
			for name, query := range cfg.Connections[cfg.Current.Name].Queries {
				fmt.Printf("- %s (%s)\n", name, query)
			}
		}

	case "edit":
		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "vim"
		}

		cmd := exec.Command(editor, cfgPath)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			log.Fatalf("Failed to open editor: %v", err)
		}

	case "history":
		return

	default:
		log.Fatalf("Unknown command: %s", command)
	}

	fmt.Printf("\nconnected to: %s/%s\n", cfg.Current.DBType, cfg.Current.Name)
}

func queryDB(dbType, connStr, query string) {
	db, err := sql.Open(dbType, connStr)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatalf("Cannot connect to database: %v", err)
	}

	rows, err := db.Query(query)
	if err != nil {
		log.Fatalf("Error executing query: %v", err)
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		log.Fatalf("Error getting columns: %v", err)
	}

	// Prepare scan arguments
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range columns {
		valuePtrs[i] = &values[i]
	}

	// Collect all rows data
	var data [][]string

	for rows.Next() {
		err = rows.Scan(valuePtrs...)
		if err != nil {
			log.Fatalf("Error scanning row: %v", err)
		}

		rowData := make([]string, len(columns))
		for i, val := range values {
			if val == nil {
				rowData[i] = "NULL"
			} else {
				rowData[i] = fmt.Sprintf("%v", val)
			}
		}
		data = append(data, rowData)
	}

	if err = rows.Err(); err != nil {
		log.Fatalf("Error during iteration: %v", err)
	}

	if err := table.RenderTable(columns, data); err != nil {
		log.Fatalf("Error rendering table: %v", err)
	}
}

type Connection struct {
	DBType             string            `yaml:"db_type"`
	DBConnectionString string            `yaml:"db_connection_string"`
	Queries            map[string]string `yaml:"queries"`
}

type Style struct {
	Accent string `yaml:"accent_color"`
}

type History struct {
	Size int `yaml:"size"`
}

type Config struct {
	Current struct {
		Name               string `yaml:"name"`
		DBType             string `yaml:"db_type"`
		DBConnectionString string `yaml:"db_connection_string"`
	} `yaml:"current"`
	Connections map[string]Connection `yaml:"connections"`
	Style Style `yaml:"style"`
	History History `yaml:"history"`
}

func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("Creating blank config file at", cfgPath)
			cfg := &Config{
				Connections: make(map[string]Connection),
			}
			err := SaveConfig(cfgPath, cfg)
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

func SaveConfig(path string, cfg *Config) error {
	dir := filepath.Dir(path)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
