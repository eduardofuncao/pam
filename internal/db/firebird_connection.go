package db

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/nakagami/firebirdsql"
)

type FirebirdConnection struct {
	*BaseConnection
	db *sql.DB
}

func NewFirebirdConnection(name, connStr string) (*FirebirdConnection, error) {
	return &FirebirdConnection{
		BaseConnection: &BaseConnection{
			Name:       name,
			DbType:     "firebird",
			ConnString: connStr,
			Queries:    make(map[string]Query),
		},
	}, nil
}

func (f *FirebirdConnection) Open() error {
	var err error
	f.db, err = sql.Open("firebirdsql", f.ConnString)
	if err != nil {
		return fmt.Errorf("failed to open firebird database: %w", err)
	}

	if err := f.db.Ping(); err != nil {
		return fmt.Errorf("failed to ping firebird database: %w", err)
	}

	return nil
}

func (f *FirebirdConnection) Ping() error {
	if f.db == nil {
		return fmt.Errorf("database not initialized")
	}
	return f.db.Ping()
}

func (f *FirebirdConnection) Close() error {
	if f.db != nil {
		return f.db.Close()
	}
	return nil
}

func (f *FirebirdConnection) Query(queryName string, args ...any) (any, error) {
	query, ok := f.Queries[queryName]
	if !ok {
		return nil, fmt.Errorf("query %s not found", queryName)
	}

	f.SetLastQuery(query)

	return f.db.Query(query.SQL, args...)
}

func (f *FirebirdConnection) ExecQuery(sqlStr string, args ...any) (*sql.Rows, error) {
	return f.db.Query(sqlStr, args...)
}

func (f *FirebirdConnection) Exec(sqlStr string, args ...any) error {
	_, err := f.db.Exec(sqlStr, args...)
	return err
}

func (f *FirebirdConnection) SetSchema(schema string) {
	// Firebird doesn't use schemas like PostgreSQL
	// No-op implementation
}

func (f *FirebirdConnection) GetInfoSQL(infoType string) string {
	switch infoType {
	case "tables":
		return `
			SELECT TRIM(RDB$RELATION_NAME) as name
			FROM RDB$RELATIONS
			WHERE RDB$VIEW_BLR IS NULL
			AND RDB$SYSTEM_FLAG = 0
			ORDER BY RDB$RELATION_NAME
		`
	case "views":
		return `
			SELECT TRIM(RDB$RELATION_NAME) as name
			FROM RDB$RELATIONS
			WHERE RDB$VIEW_BLR IS NOT NULL
			AND RDB$SYSTEM_FLAG = 0
			ORDER BY RDB$RELATION_NAME
		`
	case "columns":
		return `
			SELECT TRIM(RF.RDB$FIELD_NAME) as column_name,
				   CASE F.RDB$FIELD_TYPE
					   WHEN 7 THEN 'SMALLINT'
					   WHEN 8 THEN 'INTEGER'
					   WHEN 9 THEN 'QUAD'
					   WHEN 10 THEN 'FLOAT'
					   WHEN 12 THEN 'DATE'
					   WHEN 13 THEN 'TIME'
					   WHEN 14 THEN
					       CASE
					           WHEN F.RDB$CHARACTER_LENGTH > 0
					           THEN 'CHAR(' || CAST(F.RDB$CHARACTER_LENGTH AS INTEGER) || ')'
					           ELSE 'CHAR'
					       END
					   WHEN 16 THEN 'BIGINT'
					   WHEN 27 THEN 'DOUBLE PRECISION'
					   WHEN 35 THEN 'TIMESTAMP'
					   WHEN 37 THEN
					       CASE
					           WHEN F.RDB$CHARACTER_LENGTH > 0
					           THEN 'VARCHAR(' || CAST(F.RDB$CHARACTER_LENGTH AS INTEGER) || ')'
					           ELSE 'VARCHAR'
					       END
					   WHEN 261 THEN 'BLOB'
					   ELSE 'UNKNOWN'
				   END as data_type,
				   RF.RDB$NULL_FLAG as nullable,
				   COALESCE(F.RDB$CHARACTER_LENGTH, 0) as character_maximum_length
			FROM RDB$RELATION_FIELDS RF
			LEFT JOIN RDB$FIELDS F ON RF.RDB$FIELD_SOURCE = F.RDB$FIELD_NAME
			WHERE TRIM(RF.RDB$RELATION_NAME) = ?
			ORDER BY RF.RDB$FIELD_POSITION
		`
	default:
		return ""
	}
}

func (f *FirebirdConnection) GetTables() ([]string, error) {
	if f.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	query := `
		SELECT TRIM(RDB$RELATION_NAME) as name
		FROM RDB$RELATIONS
		WHERE RDB$VIEW_BLR IS NULL
		  AND RDB$SYSTEM_FLAG = 0
		ORDER BY RDB$RELATION_NAME
	`

	rows, err := f.db.Query(query)
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

func (f *FirebirdConnection) GetViews() ([]string, error) {
	if f.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	query := `
		SELECT TRIM(RDB$RELATION_NAME) as name
		FROM RDB$RELATIONS
		WHERE RDB$VIEW_BLR IS NOT NULL
		  AND RDB$SYSTEM_FLAG = 0
		ORDER BY RDB$RELATION_NAME
	`

	rows, err := f.db.Query(query)
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

func (f *FirebirdConnection) GetForeignKeys(tableName string) ([]ForeignKey, error) {
	if f.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	// Firebird foreign key query
	query := `
		SELECT
			TRIM(RCONSEG.RDB$FIELD_NAME) as column_name,
			TRIM(RREL.RDB$RELATION_NAME) as foreign_table,
			TRIM(RPKEYSEG.RDB$FIELD_NAME) as foreign_column
		FROM RDB$RELATION_CONSTRAINTS RCON
		JOIN RDB$INDICES RIND ON RCON.RDB$INDEX_NAME = RIND.RDB$INDEX_NAME
		JOIN RDB$INDEX_SEGMENTS RCONSEG ON RIND.RDB$INDEX_NAME = RCONSEG.RDB$INDEX_NAME
		JOIN RDB$RELATION_CONSTRAINTS RREF ON RIND.RDB$FOREIGN_KEY = RREF.RDB$INDEX_NAME
		JOIN RDB$INDICES RINDREF ON RREF.RDB$INDEX_NAME = RINDREF.RDB$INDEX_NAME
		JOIN RDB$INDEX_SEGMENTS RPKEYSEG ON RINDREF.RDB$INDEX_NAME = RPKEYSEG.RDB$INDEX_NAME
		JOIN RDB$RELATIONS RREL ON RREF.RDB$RELATION_NAME = RREL.RDB$RELATION_NAME
		WHERE TRIM(RCON.RDB$RELATION_NAME) = ?
		AND RCON.RDB$CONSTRAINT_TYPE = 'FOREIGN KEY'
		ORDER BY RCONSEG.RDB$FIELD_POSITION
	`

	rows, err := f.db.Query(query, strings.ToUpper(tableName))
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

func (f *FirebirdConnection) GetForeignKeysReferencingTable(tableName string) ([]ForeignKey, error) {
	if f.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	// Find all foreign keys that reference this table (reverse direction)
	query := `
		SELECT
			TRIM(RCONSEG.RDB$FIELD_NAME) as column_name,
			TRIM(RREL.RDB$RELATION_NAME) as foreign_table,
			TRIM(RPKEYSEG.RDB$FIELD_NAME) as foreign_column
		FROM RDB$RELATION_CONSTRAINTS RCON
		JOIN RDB$INDICES RIND ON RCON.RDB$INDEX_NAME = RIND.RDB$INDEX_NAME
		JOIN RDB$INDEX_SEGMENTS RCONSEG ON RIND.RDB$INDEX_NAME = RCONSEG.RDB$INDEX_NAME
		JOIN RDB$RELATION_CONSTRAINTS RREF ON RIND.RDB$FOREIGN_KEY = RREF.RDB$INDEX_NAME
		JOIN RDB$INDICES RINDREF ON RREF.RDB$INDEX_NAME = RINDREF.RDB$INDEX_NAME
		JOIN RDB$INDEX_SEGMENTS RPKEYSEG ON RINDREF.RDB$INDEX_NAME = RPKEYSEG.RDB$INDEX_NAME
		JOIN RDB$RELATIONS RREL ON RCON.RDB$RELATION_NAME = RREL.RDB$RELATION_NAME
		WHERE TRIM(RREF.RDB$RELATION_NAME) = ?
		AND RCON.RDB$CONSTRAINT_TYPE = 'FOREIGN KEY'
		ORDER BY RREL.RDB$RELATION_NAME, RCONSEG.RDB$FIELD_POSITION
	`

	rows, err := f.db.Query(query, strings.ToUpper(tableName))
	if err != nil {
		return nil, fmt.Errorf("failed to query referencing foreign keys: %w", err)
	}
	defer rows.Close()

	var foreignKeys []ForeignKey
	for rows.Next() {
		var fk ForeignKey
		// In this reverse query, the FK is in another table pointing to this table
		if err := rows.Scan(&fk.Column, &fk.ReferencedTable, &fk.ReferencedColumn); err == nil {
			foreignKeys = append(foreignKeys, fk)
		}
	}

	return foreignKeys, nil
}

func (f *FirebirdConnection) GetTableMetadata(tableName string) (*TableMetadata, error) {
	metadata := &TableMetadata{
		TableName: tableName,
	}

	pkQuery := `
		SELECT RC.RDB$CONSTRAINT_NAME
		FROM RDB$RELATION_CONSTRAINTS RC
		WHERE TRIM(RC.RDB$RELATION_NAME) = ?
		AND RC.RDB$CONSTRAINT_TYPE = 'PRIMARY KEY'
	`
	var pkName sql.NullString
	err := f.db.QueryRow(pkQuery, strings.ToUpper(tableName)).Scan(&pkName)
	if err == nil && pkName.Valid {
		pkColQuery := `
			SELECT TRIM(ICS.RDB$FIELD_NAME)
			FROM RDB$RELATION_CONSTRAINTS RC
			JOIN RDB$INDEX_SEGMENTS ICS ON RC.RDB$INDEX_NAME = ICS.RDB$INDEX_NAME
			WHERE TRIM(RC.RDB$CONSTRAINT_NAME) = ?
		`
		var pkColumn string
		err := f.db.QueryRow(pkColQuery, strings.TrimSpace(pkName.String)).Scan(&pkColumn)
		if err == nil {
			metadata.PrimaryKeys = append(metadata.PrimaryKeys, pkColumn)
		}
	}

	colQuery := f.GetInfoSQL("columns")
	rows, err := f.db.Query(colQuery, strings.ToUpper(tableName))
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var colName, dataType string
		var nullable sql.NullInt64
		var charLen int

		if err := rows.Scan(&colName, &dataType, &nullable, &charLen); err != nil {
			return nil, fmt.Errorf("failed to scan column: %w", err)
		}

		metadata.Columns = append(metadata.Columns, colName)
		metadata.ColumnTypes = append(metadata.ColumnTypes, dataType)
	}

	// Fetch foreign keys
	fks, err := f.GetForeignKeys(tableName)
	if err == nil {
		metadata.ForeignKeys = fks
	}

	return metadata, nil
}

func (f *FirebirdConnection) BuildDeleteStatement(tableName, primaryKeyCol, pkValue string) string {
	return fmt.Sprintf("DELETE FROM %s WHERE %s = '%s'", tableName, primaryKeyCol, pkValue)
}

func (f *FirebirdConnection) GetUniqueConstraints(tableName string) ([]string, error) {
	if f.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	query := `
		SELECT TRIM(ICS.RDB$FIELD_NAME) as column_name
		FROM RDB$RELATION_CONSTRAINTS RC
		JOIN RDB$INDEX_SEGMENTS ICS ON RC.RDB$INDEX_NAME = ICS.RDB$INDEX_NAME
		WHERE TRIM(RC.RDB$RELATION_NAME) = ?
		AND RC.RDB$CONSTRAINT_TYPE = 'UNIQUE'
		ORDER BY ICS.RDB$FIELD_POSITION
	`

	rows, err := f.db.Query(query, strings.ToUpper(tableName))
	if err != nil {
		return nil, fmt.Errorf("failed to query unique constraints: %w", err)
	}
	defer rows.Close()

	var uniqueColumns []string
	for rows.Next() {
		var column string
		if err := rows.Scan(&column); err == nil {
			uniqueColumns = append(uniqueColumns, column)
		}
	}

	return uniqueColumns, nil
}

func (f *FirebirdConnection) GetPlaceholder(paramIndex int) string {
	return "?"
}

func (f *FirebirdConnection) ApplyRowLimit(sqlStr string, limit int) string {
	// Firebird uses FIRST/SKIP syntax
	// Convert SELECT ... FROM to SELECT FIRST n ... FROM
	if !strings.Contains(strings.ToUpper(sqlStr), "SELECT") {
		return sqlStr
	}

	sqlStr = strings.TrimSpace(sqlStr)
	upperSQL := strings.ToUpper(sqlStr)

	// Check if already has FIRST
	if strings.Contains(upperSQL, " FIRST ") {
		return sqlStr
	}

	// Find SELECT position
	selectPos := strings.Index(upperSQL, "SELECT")
	if selectPos == -1 {
		return sqlStr
	}

	// Insert FIRST n after SELECT
	before := sqlStr[:selectPos+6] // "SELECT" length
	after := sqlStr[selectPos+6:]

	return fmt.Sprintf("%s FIRST %d %s", before, limit, after)
}
