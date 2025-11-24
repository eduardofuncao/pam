package db

import (
	"database/sql"
	"fmt"
	"log"
	"regexp"
	"strings"
)

type Cell struct {
	Value       string // Display value
	RawValue    any    // Original database value for queries
	ColumnName  string
	ColumnType  string
	RowIndex    int
	ColumnIndex int
}

type Row []Cell

type TableData struct {
	Columns    []string
	Rows       []Row
	TableName  string
	SQL        string
	Connection DatabaseConnection
}

func FormatTableData(rows *sql.Rows) (columns []string, data [][]string, err error) {
	columns, err = rows.Columns()
	if err != nil {
		log.Fatalf("Error getting columns: %v", err)
	}

	values := make([]any, len(columns))
	valuePtrs := make([]any, len(columns))
	for i := range columns {
		valuePtrs[i] = &values[i]
	}
	for rows.Next() {
		err = rows.Scan(valuePtrs...)
		if err != nil {
			log.Fatalf("Error scanning row: %v", err)
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
		log.Fatalf("Error during iteration: %v", err)
	}
	return columns, data, nil
}

func BuildTableData(rows *sql.Rows, sqlQuery string, conn DatabaseConnection) (*TableData, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("getting columns: %w", err)
	}

	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, fmt.Errorf("getting column types: %w", err)
	}

	tableName := extractTableName(sqlQuery)

	values := make([]any, len(columns))
	valuePtrs := make([]any, len(columns))
	for i := range columns {
		valuePtrs[i] = &values[i]
	}

	var tableRows []Row
	rowIndex := 0

	for rows.Next() {
		err = rows.Scan(valuePtrs...)
		if err != nil {
			return nil, fmt.Errorf("scanning row: %w", err)
		}

		row := make(Row, len(columns))
		for colIndex, val := range values {
			cellValue := "NULL"
			if val != nil {
				switch v := val.(type) {
				case []byte:
					cellValue = string(v)
				default:
					cellValue = fmt.Sprintf("%v", v)
				}
			}
			row[colIndex] = Cell{
				Value:       cellValue,
				RawValue:    val,
				ColumnName:  columns[colIndex],
				ColumnType:  columnTypes[colIndex].DatabaseTypeName(),
				RowIndex:    rowIndex,
				ColumnIndex: colIndex,
			}
		}
		tableRows = append(tableRows, row)
		rowIndex++
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating rows: %w", err)
	}

	return &TableData{
		Columns:    columns,
		Rows:       tableRows,
		TableName:  tableName,
		SQL:        sqlQuery,
		Connection: conn,
	}, nil
}

func extractTableName(sqlQuery string) string {
	re := regexp.MustCompile(`(?i)FROM\s+([a-zA-Z_][a-zA-Z0-9_]*)`)
	matches := re.FindStringSubmatch(sqlQuery)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return ""
}

func GetNextQueryId(queries map[string]Query) (id int) {
	used := make(map[int]bool)
	for _, query := range queries{
		used[query.Id] = true
	}
	for i := 1; ;i++ {
		if !used[i]{
			return i
		}
	}
}
