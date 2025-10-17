package db

import "fmt"

type PostgresAdapter struct {
	BaseAdapter
}

func NewPostgresAdapter(config DatabaseConfig) *PostgresAdapter {
	return &PostgresAdapter{
		BaseAdapter: BaseAdapter{
			Config: config,
			Driver: "postgres",
		},
	}
}

func (p *PostgresAdapter) GetConnectionString() string {
	if p.Config.ConnString != "" {
		return p.Config.ConnString
	}
	
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s",
		p.Config.Host,
		p.Config.Port,
		p.Config.Username,
		p.Config.Password,
		p.Config.Database,
	)
	
	if sslMode, ok := p.Config.Options["sslmode"]; ok {
		connStr += fmt.Sprintf(" sslmode=%s", sslMode)
	} else {
		connStr += " sslmode=disable"
	}
	
	return connStr
}
