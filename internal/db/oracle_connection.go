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
	
	if oc.Schema != "" {
		alterSessionSQL := fmt.Sprintf("ALTER SESSION SET CURRENT_SCHEMA = %s", oc.Schema)
		_, err = oc.db. Exec(alterSessionSQL)
		if err != nil {
			oc.db.Close()
			return fmt.Errorf("failed to set schema to '%s': %w", oc.Schema, err)
		}
	}
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
	
	// Get the current schema (respects ALTER SESSION SET CURRENT_SCHEMA)
	var currentOwner string
	ownerQuery := `SELECT SYS_CONTEXT('USERENV', 'CURRENT_SCHEMA') FROM DUAL`
	row := oc.db.QueryRow(ownerQuery)
	if err := row.Scan(&currentOwner); err != nil {
		// If we can't get the owner, fall back to the BaseConnection. Schema if set
		if oc.Schema != "" {
			currentOwner = strings.ToUpper(oc. Schema)
		} else {
			currentOwner = ""
		}
	}
	
	// Primary key query
	pkQuery := `
		SELECT cols.column_name
		FROM all_constraints cons
		JOIN all_cons_columns cols ON cons.constraint_name = cols.constraint_name
			AND cons.owner = cols.owner
		WHERE cons.constraint_type = 'P'
		AND cons.table_name = : 1
		AND ROWNUM = 1
		ORDER BY cols.position
	`
	
	// Add owner filter if we have it
	if currentOwner != "" {
		pkQuery = `
			SELECT cols.column_name
			FROM all_constraints cons
			JOIN all_cons_columns cols ON cons.constraint_name = cols.constraint_name
				AND cons.owner = cols.owner
			WHERE cons.constraint_type = 'P'
			AND cons. table_name = :1
			AND cons.owner = :2
			AND ROWNUM = 1
			ORDER BY cols.position
		`
	}
	
	metadata := &TableMetadata{
		TableName:  tableName,
	}
	
	// Execute PK query
	var rows *sql.Rows
	var err error
	if currentOwner != "" {
		rows, err = oc.db. Query(pkQuery, upperTableName, currentOwner)
	} else {
		rows, err = oc.db.Query(pkQuery, upperTableName)
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed to query oracle primary key: %w", err)
	}
	defer rows.Close()
	
	if rows.Next() {
		var pkColumn string
		if err := rows. Scan(&pkColumn); err == nil {
			metadata.PrimaryKey = pkColumn
		}
	}
	
	// Column metadata query - FILTER BY OWNER TO AVOID DUPLICATES
	colQuery := `
		SELECT column_name, data_type, data_length, data_precision, data_scale
		FROM all_tab_columns
		WHERE table_name = : 1
		ORDER BY column_id
	`
	
	// Add owner filter if we have it
	if currentOwner != "" {
		colQuery = `
			SELECT column_name, data_type, data_length, data_precision, data_scale
			FROM all_tab_columns
			WHERE table_name = :1
			  AND owner = :2
			ORDER BY column_id
		`
	}
	
	// Execute column query
	var colRows *sql.Rows
	if currentOwner != "" {
		colRows, err = oc. db.Query(colQuery, upperTableName, currentOwner)
	} else {
		colRows, err = oc.db. Query(colQuery, upperTableName)
	}
	
	if err != nil {
		return metadata, nil // Return partial metadata
	}
	defer colRows. Close()
	
	for colRows.Next() {
		var colName, dataType string
		var dataLength, dataPrecision, dataScale sql.NullInt64
		
		if err := colRows.Scan(&colName, &dataType, &dataLength, &dataPrecision, &dataScale); err != nil {
			continue
		}
		
		// Build type string
		var fullType string
		if dataType == "CHAR" || dataType == "VARCHAR2" || dataType == "NVARCHAR2" || dataType == "NCHAR" {
			if dataLength.Valid {
				fullType = fmt. Sprintf("%s(%d)", dataType, dataLength.Int64)
			} else {
				fullType = dataType
			}
		} else if dataType == "NUMBER" {
			if dataPrecision.Valid && dataScale.Valid {
				fullType = fmt. Sprintf("%s(%d,%d)", dataType, dataPrecision.Int64, dataScale.Int64)
			} else if dataPrecision.Valid {
				fullType = fmt.Sprintf("%s(%d)", dataType, dataPrecision.Int64)
			} else {
				fullType = dataType
			}
		} else if dataType == "BLOB" || dataType == "CLOB" {
			// Don't show length for LOBs
			fullType = dataType
		} else {
			fullType = dataType
		}
		
		metadata.Columns = append(metadata.Columns, colName)
		metadata.ColumnTypes = append(metadata.ColumnTypes, fullType)
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
