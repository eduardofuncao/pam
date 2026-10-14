package table

import (
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/eduardofuncao/squix/internal/config"
	"github.com/eduardofuncao/squix/internal/db"
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
	visibility config.UIVisibility,
	keyMap *config.KeyMap,
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
		visibility,
		keyMap,
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
