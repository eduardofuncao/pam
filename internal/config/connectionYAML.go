package config

import (
	"log"
	"os"

	"github.com/eduardofuncao/squix/internal/db"
)

type ConnectionYAML struct {
	Name       string `yaml:"name"`
	DBType     string `yaml:"db_type"`
	ConnString string `yaml:"conn_string"`
	Schema     string `yaml:"schema,omitempty"`
	// Deprecated: legacy migration input only. Saved queries now live in
	// queries/<group>.sql files. Kept so the first load after upgrade can lift
	// legacy per-connection queries; cleared by MigrateQueriesToFiles and never
	// repopulated.
	Queries map[string]db.Query `yaml:"queries,omitempty"`
}

func ToConnectionYAML(conn db.DatabaseConnection) *ConnectionYAML {
	return &ConnectionYAML{
		Name:       conn.GetName(),
		DBType:     conn.GetDbType(),
		ConnString: conn.GetConnString(),
		Schema:     conn.GetSchema(),
	}
}

func FromConnectionYaml(yc *ConnectionYAML) db.DatabaseConnection {
	conn, err := db.CreateConnection(yc.Name, yc.DBType, os.ExpandEnv(yc.ConnString))
	if err != nil {
		log.Fatalf("could not create connection from yaml for: %s/%s", yc.DBType, yc.Name)
	}
	conn.SetSchema(yc.Schema)
	return conn
}
