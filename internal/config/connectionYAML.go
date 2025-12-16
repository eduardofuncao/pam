package config

import (
	"log"

	"github.com/eduardofuncao/pam/internal/db"
	"github.com/eduardofuncao/pam/internal/db/connections"
	"github.com/eduardofuncao/pam/internal/db/types"
)

type ConnectionYAML struct {
	Name       string              `yaml:"name"`
	DBType     string              `yaml:"db_type"`
	ConnString string              `yaml:"conn_string"`
	Queries    map[string]types.Query `yaml:"queries"`
}

func ToConnectionYAML(conn connections.DatabaseConnection) (ConnectionYAML) {
	return ConnectionYAML{
		Name: conn.GetName(),
		DBType: conn.GetDbType(),
		ConnString: conn.GetConnString(),
		Queries: conn.GetQueries(),
	}
}

func FromConnectionYaml(yc ConnectionYAML) (connections.DatabaseConnection) {
	conn, err := db.CreateConnection(yc.Name, yc.DBType, yc.ConnString)
	if err != nil {
		log.Fatalf("could not create connection from yaml for: %s/%s", yc.DBType, yc.Name)
	}
	conn.SetQueries(yc.Queries)
	return conn
}
