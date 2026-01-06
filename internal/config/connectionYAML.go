package config

import (
	"log"

	"github.com/eduardofuncao/pam/internal/db"
)

type ConnectionYAML struct {
	Name       string              `yaml:"name"`
	DBType     string              `yaml:"db_type"`
	ConnString string              `yaml:"conn_string"`
	Schema     string              `yaml:"schema,omitempty"`
	Queries    map[string]db.Query `yaml:"queries"`
	LastQuery  db.Query               `yaml:"last_query"`
}

func ToConnectionYAML(conn db.DatabaseConnection) *ConnectionYAML {
	return &ConnectionYAML{
		Name:       conn.GetName(),
		DBType:     conn.GetDbType(),
		ConnString: conn.GetConnString(),
		Schema:     conn.GetSchema(),
		Queries:    conn.GetQueries(),
		LastQuery:  conn.GetLastQuery(),
	}
}

func FromConnectionYaml(yc *ConnectionYAML) db.DatabaseConnection {
	conn, err := db.CreateConnection(yc.Name, yc.DBType, yc.ConnString)
	if err != nil {
		log.Fatalf("could not create connection from yaml for: %s/%s", yc.DBType, yc.Name)
	}
	conn.SetSchema(yc.Schema)
	conn.SetQueries(yc.Queries)
	conn.SetLastQuery(yc.LastQuery)
	return conn
}
