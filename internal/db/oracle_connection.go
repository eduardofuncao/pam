package db

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/godror/godror"
)

type OracleConnection struct {
	*BaseConnection
	db *sql.DB
}

func NewOracleConnection(name, connStr string) (*OracleConnection, error) {
	bc := &BaseConnection{
		Name:       name,
		DbType:     "oracle",
		ConnString: connStr,
	}
	return &OracleConnection{BaseConnection: bc}, nil
}

func (oc *OracleConnection) Open() error {
	db, err := sql.Open("godror", oc.ConnString)
	if err != nil {
		return err
	}
	oc.db = db
	return nil
}

func (oc *OracleConnection) Ping() error {
	if oc.db == nil {
		return fmt.Errorf("database is not open")
	}
	return oc.db.Ping()
}

func (oc *OracleConnection) Close() error {
	if oc.db != nil {
		return oc.db.Close()
	}
	return nil
}

func (oc *OracleConnection) Query(queryName string, args ...any) (any, error) {
	query, exists := oc.Queries[queryName]
	if !exists {
		return nil, fmt.Errorf("query not found: %s", queryName)
	}
	return oc.db.Query(query.SQL, args...)
}

func (oc *OracleConnection) ExecQuery(sql string, args ... any) (*sql.Rows, error) {
	return oc.db.Query(sql, args...)
}

func (oc *OracleConnection) Exec(sql string, args ... any) error {
	_, err := oc.db.Exec(sql, args...)
	return err
}

func (oc *OracleConnection) GetTableMetadata(tableName string) (*TableMetadata, error) {
	if oc.db == nil {
		return nil, fmt.Errorf("database is not open")
	}
	
	upperTableName := strings.ToUpper(tableName)
	
	pkQuery := `
		SELECT cols.column_name
		FROM all_constraints cons
		JOIN all_cons_columns cols ON cons.constraint_name = cols.constraint_name
			AND cons.owner = cols.owner
		WHERE cons.constraint_type = 'P'
		AND cons.table_name = :1
		AND ROWNUM = 1
		ORDER BY cols.position
	`
	
	rows, err := oc.db.Query(pkQuery, upperTableName)
	if err != nil {
		return nil, fmt.Errorf("failed to query oracle primary key: %w", err)
	}
	defer rows.Close()
	
	metadata := &TableMetadata{
		TableName: tableName,
	}
	
	if rows.Next() {
		var pkColumn string
		if err := rows.Scan(&pkColumn); err == nil {
			metadata.PrimaryKey = pkColumn
		}
	}
	
	colQuery := `
		SELECT column_name,
		       CASE
		           WHEN data_type IN ('VARCHAR2', 'CHAR', 'NVARCHAR2', 'NCHAR') 
		           THEN data_type || '(' || data_length || ')'
		           WHEN data_type = 'NUMBER' AND data_precision IS NOT NULL
		           THEN data_type || '(' || data_precision || ',' || NVL(data_scale, 0) || ')'
		           ELSE data_type
		       END as full_type
		FROM all_tab_columns
		WHERE table_name = :1
		ORDER BY column_id
	`
	
	colRows, err := oc.db.Query(colQuery, upperTableName)
	if err == nil {
		defer colRows.Close()
		for colRows.Next() {
			var colName, colType string
			if err := colRows.Scan(&colName, &colType); err == nil {
				metadata.Columns = append(metadata.Columns, colName)
				metadata.ColumnTypes = append(metadata.ColumnTypes, colType)  // ADD THIS
			}
		}
	}
	
	return metadata, nil
}

func (oc *OracleConnection) BuildUpdateStatement(tableName, columnName, currentValue, pkColumn, pkValue string) string {
	escapedValue := strings.ReplaceAll(currentValue, "'", "''")
	
	if pkColumn != "" && pkValue != "" {
		escapedPkValue := strings.ReplaceAll(pkValue, "'", "''")
		return fmt.Sprintf(
			"-- Oracle UPDATE statement\nUPDATE %s\nSET %s = '%s'\nWHERE %s = '%s';",
			tableName,
			columnName,
			escapedValue,
			pkColumn,
			escapedPkValue,
		)
	}
	
	return fmt.Sprintf(
		"-- Oracle UPDATE statement\n-- No primary key specified. Edit WHERE clause manually.\nUPDATE %s\nSET %s = '%s'\nWHERE <condition>;\n-- COMMIT;",
		tableName,
		columnName,
		escapedValue,
	)
}

func (oc *OracleConnection) BuildDeleteStatement(tableName, primaryKeyCol, pkValue string) string {
	escapedPkValue := strings. ReplaceAll(pkValue, "'", "''")
	
	return fmt.Sprintf(
		"-- Oracle DELETE statement\n-- WARNING: This will permanently delete data!\n-- Ensure the WHERE clause is correct.\n\nDELETE FROM %s\nWHERE %s = '%s';\n-- COMMIT;",
		tableName,
		primaryKeyCol,
		escapedPkValue,
	)
}
