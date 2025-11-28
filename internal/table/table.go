package table

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/eduardofuncao/pam/internal/db"
)

func Render(columns []string, data [][]string, elapsed time.Duration, conn db. DatabaseConnection, tableName, primaryKeyCol string) error {
	model := New(columns, data, elapsed, conn, tableName, primaryKeyCol)
	p := tea.NewProgram(model)
	_, err := p.Run()
	return err
}
