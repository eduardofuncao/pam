package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"gopkg.in/yaml.v2"
)

var cfgPath = os.ExpandEnv("$HOME/.config/pam/config.yaml")

func main() {
	cfg, err := loadConfig(cfgPath)
	if err != nil {
		log.Fatal("Could not load context from config file", err)
	}

	if len(os.Args) < 2 {
		fmt.Println("Usage:")
		fmt.Println("pam connect <db-type> <connection-string>")
		fmt.Println("pam get <db-type> <connection-string> <sql-query>")
		fmt.Println("\ndb-type: postgres or mysql")
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "connect":
		if len(os.Args) < 4 {
			log.Fatal("Usage: connect <db-type> <connection-string>")
		}
		dbType := os.Args[2]
		connStr := os.Args[3]
		connectDB(dbType, connStr)

	case "context":
		if len(os.Args) < 4 {
			log.Fatal("Usage: pam context <name> <db-type> <connection-string>")
		}
		cfg.Current.Name = os.Args[2]
		cfg.Current.DBType = os.Args[3]
		cfg.Current.DBConnectionString = os.Args[4]
		SaveConfig(cfgPath, cfg)

	case "add":
		if len(os.Args) < 3 {
			log.Fatal("Usage: pam add <query-name> <query>")
		}
		queryToSave := SavedQuery{
			Query: os.Args[3],
		}
		//add connection to cfg if it does not exist
		_, ok := cfg.Connections[cfg.Current.Name] 
		if !ok {
			cfg.Connections[cfg.Current.Name] = make(map[string]SavedQuery)
		}
		cfg.Connections[cfg.Current.Name][os.Args[2]] = queryToSave
		SaveConfig(cfgPath, cfg)

	case "query":
		if len(os.Args) < 3 {
			log.Fatal("Usage:pam query <query-name>")
		}
		dbType := cfg.Current.DBType
		connStr := cfg.Current.DBConnectionString
		connections := cfg.Connections[cfg.Current.Name]
		query := connections[os.Args[2]]
		log.Println(dbType, connStr, os.Args[2], query)
		queryDB(dbType, connStr, query.Query)

	default:
		log.Fatalf("Unknown command: %s", command)
	}
}

func connectDB(dbType, connStr string) {
	db, err := sql.Open(dbType, connStr)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatalf("Cannot connect to database: %v", err)
	}

	fmt.Println("✓ Successfully connected to database!")
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

	// Print header
	for i, col := range columns {
		if i > 0 {
			fmt.Print("\t")
		}
		fmt.Print(col)
	}
	fmt.Println()

	// Print separator
	for range columns {
		fmt.Print("--------\t")
	}
	fmt.Println()

	// Prepare scan arguments
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range columns {
		valuePtrs[i] = &values[i]
	}

	// Print rows
	for rows.Next() {
		err = rows.Scan(valuePtrs...)
		if err != nil {
			log.Fatalf("Error scanning row: %v", err)
		}

		for i, val := range values {
			if i > 0 {
				fmt.Print("\t")
			}
			if val == nil {
				fmt.Print("NULL")
			} else {
				fmt.Print(val)
			}
		}
		fmt.Println()
	}

	if err = rows.Err(); err != nil {
		log.Fatalf("Error during iteration: %v", err)
	}
}

type SavedQuery struct {
	Query string `yaml:"query"`
}

type Config struct {
	Current struct {
		Name               string `yaml:"name"`
		DBType             string `yaml:"db_type"`
		DBConnectionString string `yaml:"db_connection_string"`
	} `yaml:"current"`
	Connections map[string]map[string]SavedQuery `yaml:"connections"`
}

func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			cfg := &Config{
				Connections: make(map[string]map[string]SavedQuery),
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
