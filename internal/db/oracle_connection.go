//go:build cgo

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
		_, err = oc.db.Exec(alterSessionSQL)
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

func (oc *OracleConnection) ExecQuery(sql string, args ...any) (*sql.Rows, error) {
	return oc.db.Query(sql, args...)
}

func (oc *OracleConnection) Exec(sql string, args ...any) error {
	_, err := oc.db.Exec(sql, args...)
	return err
}

func (oc *OracleConnection) GetTableMetadata(tableName string) (*TableMetadata, error) {
	if oc.db == nil {
		return nil, fmt.Errorf("database is not open")
	}

	upperTableName := strings.ToUpper(tableName)

	var currentOwner string
	ownerQuery := `SELECT SYS_CONTEXT('USERENV', 'CURRENT_SCHEMA') FROM DUAL`
	row := oc.db.QueryRow(ownerQuery)
	if err := row.Scan(&currentOwner); err != nil {
		if oc.Schema != "" {
			currentOwner = strings.ToUpper(oc.Schema)
		} else {
			currentOwner = ""
		}
	}

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
		TableName: tableName,
	}

	var rows *sql.Rows
	var err error
	if currentOwner != "" {
		rows, err = oc.db.Query(pkQuery, upperTableName, currentOwner)
	} else {
		rows, err = oc.db.Query(pkQuery, upperTableName)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to query oracle primary key: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		var pkColumn string
		if err := rows.Scan(&pkColumn); err == nil {
			metadata.PrimaryKeys = append(metadata.PrimaryKeys, pkColumn)
		}
	}

	colQuery := `
		SELECT column_name, data_type, data_length, data_precision, data_scale
		FROM all_tab_columns
		WHERE table_name = : 1
		ORDER BY column_id
	`

	if currentOwner != "" {
		colQuery = `
			SELECT column_name, data_type, data_length, data_precision, data_scale
			FROM all_tab_columns
			WHERE table_name = :1
			  AND owner = :2
			ORDER BY column_id
		`
	}

	var colRows *sql.Rows
	if currentOwner != "" {
		colRows, err = oc.db.Query(colQuery, upperTableName, currentOwner)
	} else {
		colRows, err = oc.db.Query(colQuery, upperTableName)
	}

	if err != nil {
		return metadata, nil // Return partial metadata
	}
	defer colRows.Close()

	for colRows.Next() {
		var colName, dataType string
		var dataLength, dataPrecision, dataScale sql.NullInt64

		if err := colRows.Scan(&colName, &dataType, &dataLength, &dataPrecision, &dataScale); err != nil {
			continue
		}

		// Build type string
		var fullType string
		switch dataType {
		case "CHAR", "VARCHAR2", "NVARCHAR2", "NCHAR":
			if dataLength.Valid {
				fullType = fmt.Sprintf("%s(%d)", dataType, dataLength.Int64)
			} else {
				fullType = dataType
			}
		case "NUMBER":
			if dataPrecision.Valid && dataScale.Valid {
				fullType = fmt.Sprintf("%s(%d,%d)", dataType, dataPrecision.Int64, dataScale.Int64)
			} else if dataPrecision.Valid {
				fullType = fmt.Sprintf("%s(%d)", dataType, dataPrecision.Int64)
			} else {
				fullType = dataType
			}
		case "BLOB", "CLOB":
			fullType = dataType
		default:
			fullType = dataType
		}

		metadata.Columns = append(metadata.Columns, colName)
		metadata.ColumnTypes = append(metadata.ColumnTypes, fullType)
	}

	// Fetch foreign keys
	fks, err := oc.GetForeignKeys(tableName)
	if err == nil {
		metadata.ForeignKeys = fks
	}

	return metadata, nil
}

func (oc *OracleConnection) GetInfoSQL(infoType string) string {
	schema := strings.ToUpper(oc.Schema)
	if schema == "" {
		schema = "SYS_CONTEXT('USERENV', 'CURRENT_SCHEMA')"
	} else {
		schema = "'" + schema + "'"
	}

	switch infoType {
	case "tables":
		return fmt.Sprintf(`SELECT OWNER as schema,
		       TABLE_NAME as name,
		       OWNER as owner
		FROM ALL_TABLES
		WHERE OWNER = %s`, schema)
	case "views":
		return fmt.Sprintf(`SELECT OWNER as schema,
		       VIEW_NAME as name,
		       OWNER as owner
		FROM ALL_VIEWS
		WHERE OWNER = %s`, schema)
	default:
		return ""
	}
}

func (oc *OracleConnection) GetTables() ([]string, error) {
	if oc.db == nil {
		return nil, fmt.Errorf("database is not open")
	}

	// Get the current schema
	var currentSchema string
	schemaQuery := `SELECT SYS_CONTEXT('USERENV', 'CURRENT_SCHEMA') FROM DUAL`
	row := oc.db.QueryRow(schemaQuery)
	if err := row.Scan(&currentSchema); err != nil {
		if oc.Schema != "" {
			currentSchema = oc.Schema
		} else {
			return nil, fmt.Errorf("failed to determine current schema: %w", err)
		}
	}

	query := `
		SELECT TABLE_NAME
		FROM ALL_TABLES
		WHERE OWNER = :1
		ORDER BY TABLE_NAME
	`

	rows, err := oc.db.Query(query, currentSchema)
	if err != nil {
		return nil, fmt.Errorf("failed to query tables: %w", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err == nil {
			tables = append(tables, tableName)
		}
	}

	return tables, nil
}

func (oc *OracleConnection) GetViews() ([]string, error) {
	if oc.db == nil {
		return nil, fmt.Errorf("database is not open")
	}

	// Get the current schema
	var currentSchema string
	schemaQuery := `SELECT SYS_CONTEXT('USERENV', 'CURRENT_SCHEMA') FROM DUAL`
	row := oc.db.QueryRow(schemaQuery)
	if err := row.Scan(&currentSchema); err != nil {
		if oc.Schema != "" {
			currentSchema = oc.Schema
		} else {
			return nil, fmt.Errorf("failed to determine current schema: %w", err)
		}
	}

	query := `
		SELECT VIEW_NAME
		FROM ALL_VIEWS
		WHERE OWNER = :1
		ORDER BY VIEW_NAME
	`

	rows, err := oc.db.Query(query, currentSchema)
	if err != nil {
		return nil, fmt.Errorf("failed to query views: %w", err)
	}
	defer rows.Close()

	var views []string
	for rows.Next() {
		var viewName string
		if err := rows.Scan(&viewName); err == nil {
			views = append(views, viewName)
		}
	}

	return views, nil
}

func (oc *OracleConnection) GetForeignKeys(tableName string) ([]ForeignKey, error) {
	if oc.db == nil {
		return nil, fmt.Errorf("database is not open")
	}

	upperTableName := strings.ToUpper(tableName)

	// Get the current schema
	var currentOwner string
	ownerQuery := `SELECT SYS_CONTEXT('USERENV', 'CURRENT_SCHEMA') FROM DUAL`
	row := oc.db.QueryRow(ownerQuery)
	if err := row.Scan(&currentOwner); err != nil {
		if oc.Schema != "" {
			currentOwner = strings.ToUpper(oc.Schema)
		} else {
			currentOwner = ""
		}
	}

	query := `
		SELECT
			acc.column_name,
			ac2.table_name AS foreign_table_name,
			acc2.column_name AS foreign_column_name
		FROM all_constraints ac
		JOIN all_constraints ac2 ON ac.r_constraint_name = ac2.constraint_name
		JOIN all_cons_columns acc ON ac.constraint_name = acc.constraint_name
		JOIN all_cons_columns acc2 ON ac2.constraint_name = acc2.constraint_name
		WHERE ac.constraint_type = 'R'
		AND ac.table_name = :1
		AND acc.position = 1
		AND acc2.position = 1
	`

	var rows *sql.Rows
	var err error
	if currentOwner != "" {
		query += " AND ac.owner = :2 AND ac2.owner = :3"
		rows, err = oc.db.Query(query, upperTableName, currentOwner, currentOwner)
	} else {
		rows, err = oc.db.Query(query, upperTableName)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to query foreign keys: %w", err)
	}
	defer rows.Close()

	var foreignKeys []ForeignKey
	for rows.Next() {
		var fk ForeignKey
		if err := rows.Scan(&fk.Column, &fk.ReferencedTable, &fk.ReferencedColumn); err == nil {
			foreignKeys = append(foreignKeys, fk)
		}
	}

	return foreignKeys, nil
}

func (oc *OracleConnection) GetForeignKeysReferencingTable(tableName string) ([]ForeignKey, error) {
	if oc.db == nil {
		return nil, fmt.Errorf("database is not open")
	}

	upperTableName := strings.ToUpper(tableName)

	// Get the current schema
	var currentOwner string
	ownerQuery := `SELECT SYS_CONTEXT('USERENV', 'CURRENT_SCHEMA') FROM DUAL`
	row := oc.db.QueryRow(ownerQuery)
	if err := row.Scan(&currentOwner); err != nil {
		if oc.Schema != "" {
			currentOwner = strings.ToUpper(oc.Schema)
		} else {
			currentOwner = ""
		}
	}

	query := `
		SELECT
			acc.column_name,
			ac.table_name AS foreign_table_name,
			acc2.column_name AS foreign_column_name
		FROM all_constraints ac
		JOIN all_constraints ac2 ON ac.r_constraint_name = ac2.constraint_name
		JOIN all_cons_columns acc ON ac.constraint_name = acc.constraint_name
		JOIN all_cons_columns acc2 ON ac2.constraint_name = acc2.constraint_name
		WHERE ac.constraint_type = 'R'
		AND ac2.table_name = :1
		AND acc.position = 1
		AND acc2.position = 1
	`

	var rows *sql.Rows
	var err error
	if currentOwner != "" {
		query += " AND ac.owner = :2 AND ac2.owner = :3"
		rows, err = oc.db.Query(query, upperTableName, currentOwner, currentOwner)
	} else {
		rows, err = oc.db.Query(query, upperTableName)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to query referencing foreign keys: %w", err)
	}
	defer rows.Close()

	var foreignKeys []ForeignKey
	for rows.Next() {
		var fk ForeignKey
		if err := rows.Scan(&fk.Column, &fk.ReferencedTable, &fk.ReferencedColumn); err == nil {
			foreignKeys = append(foreignKeys, fk)
		}
	}

	return foreignKeys, nil
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

func (oc *OracleConnection) ApplyRowLimit(sql string, limit int) string {
	trimmedSQL := strings.ToUpper(strings.TrimSpace(sql))
	if !strings.HasPrefix(trimmedSQL, "SELECT") && !strings.HasPrefix(trimmedSQL, "WITH") {
		return sql
	}

	upperSQL := strings.ToUpper(sql)
	if strings.Contains(upperSQL, "FETCH FIRST") || strings.Contains(upperSQL, "ROWNUM") {
		return sql
	}

	return fmt.Sprintf("%s\nFETCH FIRST %d ROWS ONLY", strings.TrimRight(sql, ";"), limit)
}

func (oc *OracleConnection) BuildDeleteStatement(tableName, primaryKeyCol, pkValue string) string {
	escapedPkValue := strings.ReplaceAll(pkValue, "'", "''")

	return fmt.Sprintf(
		"-- Oracle DELETE statement\n-- WARNING: This will permanently delete data!\n-- Ensure the WHERE clause is correct.\n\nDELETE FROM %s\nWHERE %s = '%s';\n-- COMMIT;",
		tableName,
		primaryKeyCol,
		escapedPkValue,
	)
}

func (oc *OracleConnection) GetPlaceholder(paramIndex int) string {
	return fmt.Sprintf(":%d", paramIndex)
}
