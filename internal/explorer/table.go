package explorer

import (
	tea "charm.land/bubbletea/v2"
	"github.com/eduardofuncao/squix/internal/config"
	"github.com/eduardofuncao/squix/internal/db"
)

func Render(
	conn db.DatabaseConnection,
	tables, views []string,
	restoreCursor int,
	visibility config.UIVisibility,
	keyMap *config.KeyMap,
) (Model, error) {
	model := New(conn, tables, views, restoreCursor, visibility, keyMap)
	p := tea.NewProgram(model)
	finalModel, err := p.Run()
	if err != nil {
		return model, err
	}
	return finalModel.(Model), nil
}
