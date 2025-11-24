package table

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/eduardofuncao/pam/internal/db"
)

func Render(tableData *db.TableData, elapsed time.Duration) error {
	return RenderWithExecutor(tableData, elapsed, nil)
}

func RenderWithExecutor(tableData *db.TableData, elapsed time.Duration, cmdExec CommandExecutor) error {
	model := New(tableData, elapsed, cmdExec)
	p := tea.NewProgram(model)
	_, err := p.Run()
	return err
}
