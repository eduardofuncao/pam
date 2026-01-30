package table

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/eduardofuncao/pam/internal/db"
)

func Render(
	columns []string,
	columnTypes []string,
	data [][]string,
	elapsed time.Duration,
	conn db.DatabaseConnection,
	tableName, primaryKeyCol string,
	query db.Query,
	columnWidth int,
	saveCallback func(query db.Query) (db.Query, error),
	initialStatus ...string,
) (Model, error) {
	model := New(
		columns,
		columnTypes,
		data,
		elapsed,
		conn,
		tableName,
		primaryKeyCol,
		query,
		columnWidth,
	)
	model.saveQueryCallback = saveCallback
	if len(initialStatus) > 0 && initialStatus[0] != "" {
		model.statusMessage = initialStatus[0]
	}
	p := tea.NewProgram(model)
	finalModel, err := p.Run()
	if err != nil {
		return model, err
	}
	return finalModel.(Model), nil
}

func RenderTablesList(
	columns []string,
	data [][]string,
	elapsed time.Duration,
	conn db.DatabaseConnection,
	query db.Query,
	columnWidth int,
) (Model, error) {
	model := New(columns, nil, data, elapsed, conn, "", "", query, columnWidth)
	model.isTablesList = true
	p := tea.NewProgram(model)
	finalModel, err := p.Run()
	if err != nil {
		return model, err
	}
	return finalModel.(Model), nil
}
