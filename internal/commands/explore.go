package commands

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/eduardofuncao/pam/internal/config"
	"github.com/eduardofuncao/pam/internal/db"
	"github.com/eduardofuncao/pam/internal/spinner"
	"github.com/eduardofuncao/pam/internal/table"
)

func Explore(cfg *config.Config) {
	ExploreWithArgs(cfg, os.Args, false, nil)
}

func ExploreWithExecutor(cfg *config.Config, cmdExec table.CommandExecutor) {
	ExploreWithArgs(cfg, os.Args, false, cmdExec)
}

func ExploreWithArgs(cfg *config.Config, args []string, fromTUI bool, cmdExec table.CommandExecutor) (*db.TableData, error) {
	if len(args) < 3 {
		if fromTUI {
			return nil, fmt.Errorf("usage: explore <table-name> [--limit|-l <number>]")
		}
		fmt.Println("No table specified. Available tables:")
		ListTables(cfg)
		return nil, nil
	}

	tableName := args[2]
	limit := 1000

	if len(args) > 3 {
		for i := 3; i < len(args); i++ {
			if args[i] == "--limit" || args[i] == "-l" {
				if i+1 < len(args) {
					parsedLimit, err := strconv.Atoi(args[i+1])
					if err != nil {
						if fromTUI {
							return nil, fmt.Errorf("invalid limit value: %s", args[i+1])
						}
						log.Fatalf("Invalid limit value: %s", args[i+1])
					}
					limit = parsedLimit
					break
				}
			}
		}
	}

	currConn := config.FromConnectionYaml(cfg.Connections[cfg.CurrentConnection])

	if err := currConn.Open(); err != nil {
		if fromTUI {
			return nil, fmt.Errorf("could not open connection: %w", err)
		}
		log.Fatalf("Could not open the connection to %s/%s: %s", currConn.GetDbType(), currConn.GetName(), err)
	}

	start := time.Now()
	var done chan struct{}
	if !fromTUI {
		done = make(chan struct{})
		go spinner.Wait(done)
	}

	sqlRows, err := currConn.QueryTableWithLimit(tableName, limit)
	if err != nil {
		if !fromTUI {
			done <- struct{}{}
		}
		if fromTUI {
			return nil, fmt.Errorf("could not query table '%s': %v", tableName, err)
		}
		log.Fatalf("Could not query table '%s': %v", tableName, err)
	}

	querySQL := fmt.Sprintf("SELECT * FROM %s (LIMIT %d)", tableName, limit)
	tableData, err := db.BuildTableData(sqlRows, querySQL, currConn)
	if err != nil {
		if !fromTUI {
			done <- struct{}{}
		}
		if fromTUI {
			return nil, fmt.Errorf("error building table data: %w", err)
		}
		log.Fatalf("Error building table data: %v", err)
	}

	if !fromTUI {
		done <- struct{}{}
		elapsed := time.Since(start)
		if err := table.RenderWithExecutor(tableData, elapsed, cmdExec); err != nil {
			log.Fatalf("Error rendering table: %v", err)
		}
	}

	return tableData, nil
}
