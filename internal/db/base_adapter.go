package db

import (
	"database/sql"
	"fmt"
)

type BaseAdapter struct {
	Config DatabaseConfig
	DB     *sql.DB
	Driver string
}

func (b *BaseAdapter) Connect() error {
	connString := b.GetConnectionString()
	
	db, err := sql.Open(b.Driver, connString)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	
	if err := db.Ping(); err != nil {
		db.Close()
		return fmt.Errorf("failed to ping database: %w", err)
	}
	
	b.DB = db
	return nil
}

func (b *BaseAdapter) Close() error {
	if b.DB != nil {
		return b.DB.Close()
	}
	return nil
}

func (b *BaseAdapter) Ping() error {
	if b.DB == nil {
		return fmt.Errorf("database connection not established")
	}
	return b.DB.Ping()
}

func (b *BaseAdapter) Query(query string, args ...any) ([]string, [][]string, error) {
	if b.DB == nil {
		return nil, nil, fmt.Errorf("database connection not established")
	}
	
	rows, err := b.DB.Query(query, args...)
	if err != nil {
		return nil, nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()
	
	columns, err := rows.Columns()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get columns: %w", err)
	}
	
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range columns {
		valuePtrs[i] = &values[i]
	}
	
	var data [][]string
	for rows.Next() {
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, nil, fmt.Errorf("failed to scan row: %w", err)
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
		return nil, nil, fmt.Errorf("row iteration error: %w", err)
	}
	
	return columns, data, nil
}

func (b *BaseAdapter) Exec(query string, args ...any) (sql.Result, error) {
	if b.DB == nil {
		return nil, fmt.Errorf("database connection not established")
	}
	return b.DB.Exec(query, args...)
}

func (b *BaseAdapter) GetDB() *sql.DB {
	return b.DB
}

func (b *BaseAdapter) GetConnectionString() string {
	if b.Config.ConnString != "" {
		return b.Config.ConnString
	}
	return ""
}
