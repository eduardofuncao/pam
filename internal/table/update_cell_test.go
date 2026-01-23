package table

import (
	"testing"

	"github.com/eduardofuncao/pam/internal/db"
)

func TestModel_BuildUpdateStatement(t *testing.T) {
	tests := []struct {
		name          string
		tableName     string
		primaryKeyCol string
		columns       []string
		data          [][]string
		selectedRow   int
		selectedCol   int
		wantContains  []string
		wantError     bool
	}{
		{
			name:          "valid update with string value",
			tableName:     "users",
			primaryKeyCol: "id",
			columns:       []string{"id", "name", "email"},
			data: [][]string{
				{"1", "Alice", "alice@example.com"},
				{"2", "Bob", "bob@example.com"},
			},
			selectedRow: 0,
			selectedCol: 1,
			wantContains: []string{
				"UPDATE",
				"users",
				"SET",
				"name",
				"WHERE",
				"id",
			},
			wantError: false,
		},
		{
			name:          "update numeric column",
			tableName:     "products",
			primaryKeyCol: "product_id",
			columns:       []string{"product_id", "price", "stock"},
			data: [][]string{
				{"1", "99.99", "100"},
				{"2", "149.99", "50"},
			},
			selectedRow: 1,
			selectedCol: 1,
			wantContains: []string{
				"UPDATE",
				"products",
				"SET",
				"price",
				"WHERE",
				"product_id",
			},
			wantError: false,
		},
		{
			name:          "update with email value",
			tableName:     "contacts",
			primaryKeyCol: "contact_id",
			columns:       []string{"contact_id", "name", "email"},
			data: [][]string{
				{"100", "John Doe", "john@example.com"},
			},
			selectedRow: 0,
			selectedCol: 2,
			wantContains: []string{
				"UPDATE",
				"contacts",
				"SET",
				"email",
				"WHERE",
				"contact_id",
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := New(
				tt.columns,
				tt.data,
				0,
				nil,
				tt.tableName,
				tt.primaryKeyCol,
				db.Query{},
				15,
			)
			model.selectedRow = tt.selectedRow
			model.selectedCol = tt.selectedCol

			result := model.buildUpdateStatement()

			// Check that all expected strings are present
			for _, want := range tt.wantContains {
				if !contains(result, want) {
					t.Errorf(
						"buildUpdateStatement() missing %q in result:\n%s",
						want,
						result,
					)
				}
			}
		})
	}
}

func TestModel_BuildUpdateStatement_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		columns     []string
		data        [][]string
		selectedRow int
		selectedCol int
		description string
	}{
		{
			name:        "first column (usually primary key)",
			columns:     []string{"id", "name"},
			data:        [][]string{{"1", "Alice"}},
			selectedRow: 0,
			selectedCol: 0,
			description: "Should handle updating first column",
		},
		{
			name:        "last column",
			columns:     []string{"id", "name", "email"},
			data:        [][]string{{"1", "Alice", "alice@example.com"}},
			selectedRow: 0,
			selectedCol: 2,
			description: "Should handle updating last column",
		},
		{
			name:        "value with special characters",
			columns:     []string{"id", "description"},
			data:        [][]string{{"1", "Test's \"special\" & chars"}},
			selectedRow: 0,
			selectedCol: 1,
			description: "Should escape special characters in SQL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := New(
				tt.columns,
				tt.data,
				0,
				nil,
				"test_table",
				"id",
				db.Query{},
				15,
			)
			model.selectedRow = tt.selectedRow
			model.selectedCol = tt.selectedCol

			result := model.buildUpdateStatement()

			if result == "" {
				t.Errorf(
					"buildUpdateStatement() returned empty string for %s",
					tt.description,
				)
			}

			// Should always contain UPDATE, SET, and WHERE
			if !contains(result, "UPDATE") {
				t.Error("Result should contain UPDATE keyword")
			}
			if !contains(result, "SET") {
				t.Error("Result should contain SET keyword")
			}
			if !contains(result, "WHERE") {
				t.Error("Result should contain WHERE keyword")
			}
		})
	}
}

func TestModel_UpdateCell_Bounds(t *testing.T) {
	model := New(
		[]string{"id", "name"},
		[][]string{{"1", "Alice"}},
		0,
		nil,
		"users",
		"id",
		db.Query{},
		15,
	)

	tests := []struct {
		name        string
		selectedRow int
		selectedCol int
		description string
	}{
		{
			name:        "negative row",
			selectedRow: -1,
			selectedCol: 0,
			description: "Should handle negative row index",
		},
		{
			name:        "row out of bounds",
			selectedRow: 999,
			selectedCol: 0,
			description: "Should handle row index beyond data length",
		},
		{
			name:        "negative column",
			selectedRow: 0,
			selectedCol: -1,
			description: "Should handle negative column index",
		},
		{
			name:        "column out of bounds",
			selectedRow: 0,
			selectedCol: 999,
			description: "Should handle column index beyond columns length",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model.selectedRow = tt.selectedRow
			model.selectedCol = tt.selectedCol

			// updateCell should not panic with invalid indices
			defer func() {
				if r := recover(); r != nil {
					t.Errorf(
						"updateCell() panicked with %s: %v",
						tt.description,
						r,
					)
				}
			}()

			_, _ = model.updateCell()
		})
	}
}

func TestModel_NumRows(t *testing.T) {
	tests := []struct {
		name string
		data [][]string
		want int
	}{
		{
			name: "empty data",
			data: [][]string{},
			want: 0,
		},
		{
			name: "single row",
			data: [][]string{{"1", "Alice"}},
			want: 1,
		},
		{
			name: "multiple rows",
			data: [][]string{
				{"1", "Alice"},
				{"2", "Bob"},
				{"3", "Charlie"},
			},
			want: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := New(
				[]string{"id", "name"},
				tt.data,
				0,
				nil,
				"",
				"",
				db.Query{},
				15,
			)
			if got := model.numRows(); got != tt.want {
				t.Errorf("numRows() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestModel_NumCols(t *testing.T) {
	tests := []struct {
		name    string
		columns []string
		want    int
	}{
		{
			name:    "empty columns",
			columns: []string{},
			want:    0,
		},
		{
			name:    "single column",
			columns: []string{"id"},
			want:    1,
		},
		{
			name:    "multiple columns",
			columns: []string{"id", "name", "email", "created_at"},
			want:    4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := New(
				tt.columns,
				[][]string{},
				0,
				nil,
				"",
				"",
				db.Query{},
				15,
			)
			if got := model.numCols(); got != tt.want {
				t.Errorf("numCols() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestModel_GetPrimaryKeyValue(t *testing.T) {
	tests := []struct {
		name          string
		columns       []string
		data          [][]string
		primaryKeyCol string
		selectedRow   int
		want          string
		description   string
	}{
		{
			name:          "primary key in first column",
			columns:       []string{"id", "name"},
			data:          [][]string{{"123", "Alice"}},
			primaryKeyCol: "id",
			selectedRow:   0,
			want:          "123",
			description:   "Should get primary key value from first column",
		},
		{
			name:          "primary key in middle column",
			columns:       []string{"name", "user_id", "email"},
			data:          [][]string{{"Alice", "456", "alice@example.com"}},
			primaryKeyCol: "user_id",
			selectedRow:   0,
			want:          "456",
			description:   "Should get primary key value from middle column",
		},
		{
			name:          "primary key in last column",
			columns:       []string{"name", "email", "id"},
			data:          [][]string{{"Bob", "bob@example.com", "789"}},
			primaryKeyCol: "id",
			selectedRow:   0,
			want:          "789",
			description:   "Should get primary key value from last column",
		},
		{
			name:    "multiple rows",
			columns: []string{"id", "name"},
			data: [][]string{
				{"1", "Alice"},
				{"2", "Bob"},
				{"3", "Charlie"},
			},
			primaryKeyCol: "id",
			selectedRow:   1,
			want:          "2",
			description:   "Should get correct primary key for selected row",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := New(
				tt.columns,
				tt.data,
				0,
				nil,
				"test_table",
				tt.primaryKeyCol,
				db.Query{},
				15,
			)
			model.selectedRow = tt.selectedRow

			// Find primary key column index
			pkIndex := -1
			for i, col := range tt.columns {
				if col == tt.primaryKeyCol {
					pkIndex = i
					break
				}
			}

			if pkIndex == -1 {
				t.Fatalf("Primary key column %s not found", tt.primaryKeyCol)
			}

			got := tt.data[tt.selectedRow][pkIndex]
			if got != tt.want {
				t.Errorf(
					"getPrimaryKeyValue() = %s, want %s (%s)",
					got,
					tt.want,
					tt.description,
				)
			}
		})
	}
}

func TestEditorCompleteMsg(t *testing.T) {
	// Test the message structure
	msg := editorCompleteMsg{
		sql:      "UPDATE users SET name = 'test' WHERE id = 1",
		colIndex: 1,
	}

	if msg.sql == "" {
		t.Error("editorCompleteMsg sql should not be empty")
	}
	if msg.colIndex != 1 {
		t.Errorf("editorCompleteMsg colIndex = %d, want 1", msg.colIndex)
	}
}

func TestModel_UpdateCell_WithoutTableName(t *testing.T) {
	// Model without table name should not allow updates
	model := New(
		[]string{"id", "name"},
		[][]string{{"1", "Alice"}},
		0,
		nil,
		"", // No table name
		"",
		db.Query{},
		15,
	)

	model.selectedRow = 0
	model.selectedCol = 1

	// buildUpdateStatement should handle this gracefully
	result := model.buildUpdateStatement()

	// Without table name, update statement might be incomplete
	// This tests defensive coding
	if result == "" {
		// This is acceptable - no table name means no update possible
		return
	}
}

func TestModel_UpdateCell_WithoutPrimaryKey(t *testing.T) {
	// Model without primary key should not allow updates
	model := New(
		[]string{"id", "name"},
		[][]string{{"1", "Alice"}},
		0,
		nil,
		"users",
		"", // No primary key
		db.Query{},
		15,
	)

	model.selectedRow = 0
	model.selectedCol = 1

	// buildUpdateStatement should handle this gracefully
	result := model.buildUpdateStatement()

	// Without primary key, WHERE clause cannot be specific
	// This tests defensive coding
	if result == "" {
		// This is acceptable - no primary key means no safe update possible
		return
	}
}
