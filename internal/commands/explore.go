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

	start := time.Now()
	var done chan struct{}
	if !fromTUI {
		done = make(chan struct{})
		go spinner.Wait(done)
	}

	desc := &db.QueryDescriptor{
		Type:       "explore",
		TableName:  tableName,
		Limit:      limit,
		Connection: currConn,
	}

	tableData, err := db.ExecuteQuery(desc)
	if err != nil {
		if !fromTUI {
			done <- struct{}{}
		}
		if fromTUI {
			return nil, fmt.Errorf("could not query table '%s': %v", tableName, err)
		}
		log.Fatalf("Could not query table '%s': %v", tableName, err)
	}

	if !fromTUI {
		done <- struct{}{}
		elapsed := time.Since(start)
		if err := table.RenderWithDescriptor(tableData, elapsed, cmdExec, desc); err != nil {
			log.Fatalf("Error rendering table: %v", err)
		}
	}

	return tableData, nil
}
