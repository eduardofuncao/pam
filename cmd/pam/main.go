package main

import (
	"log"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"

	"github.com/eduardofuncao/pam/internal/commands/handler"
	"github.com/eduardofuncao/pam/internal/config"
)

func main() {
	cfg, err := config.LoadConfig(config.CfgFile)
	if err != nil {
		log.Fatal("Could not load config file", err)
	}

	handler.Parse(cfg)
}
