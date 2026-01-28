package db

import (
	"database/sql"
	"fmt"
	"log"
)

func FormatTableData(rows *sql.Rows) (columns []string, data [][]string, err error) {
	columns, _, data, err = FormatTableDataWithTypes(rows)
	return columns, data, err
}

func FormatTableDataWithTypes(rows *sql.Rows) (columns []string, columnTypes []string, data [][]string, err error) {
	columns, err = rows.Columns()
	if err != nil {
		log.Fatalf("Error getting columns: %v", err)
	}

	// Get column types from the result set
	columnTypeObjects, err := rows.ColumnTypes()
	if err != nil {
		log.Fatalf("Error getting column types: %v", err)
	}

	columnTypes = make([]string, len(columns))
	for i, ct := range columnTypeObjects {
		columnTypes[i] = ct.DatabaseTypeName()
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
				// Handle byte slices (common with MySQL text/varchar columns)
				if b, ok := val.([]byte); ok {
					rowData[i] = string(b)
				} else {
					rowData[i] = fmt.Sprintf("%v", val)
				}
			}
		}
		data = append(data, rowData)
	}

	if err = rows.Err(); err != nil {
		log.Fatalf("Error during iteration: %v", err)
	}
	return columns, columnTypes, data, nil
}

func GetNextQueryId(queries map[string]Query) (id int) {
	used := make(map[int]bool)
	for _, query := range queries {
		used[query.Id] = true
	}
	for i := 1; ; i++ {
		if !used[i] {
			return i
		}
	}
}
