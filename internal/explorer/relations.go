package explorer

import (
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/eduardofuncao/squix/internal/config"
	"github.com/eduardofuncao/squix/internal/db"
	"github.com/eduardofuncao/squix/internal/explain"
	"github.com/eduardofuncao/squix/internal/spinner"
	"github.com/eduardofuncao/squix/internal/styles"
)

const relationsDepth = 2

type relationsTreeMsg struct {
	target string
	tree   string
}

type spinnerTickMsg struct{}

func (m Model) openRelations() (tea.Model, tea.Cmd) {
	name, ok := m.selectedItemName()
	if !ok {
		m.statusMessage = styles.Faint.Render("Select a table or view first")
		return m, nil
	}

	m.relationsMode = true
	m.relationsTarget = name
	m.relationsText = ""
	m.relationsScroll = 0
	m.relationsLoading = true
	m.relationsLoadStart = time.Now()

	return m, tea.Batch(loadRelationsCmd(m.conn, name), m.spinCmd())
}

func loadRelationsCmd(conn db.DatabaseConnection, tableName string) tea.Cmd {
	return func() tea.Msg {
		return relationsTreeMsg{
			target: tableName,
			tree:   explain.BuildTree(conn, tableName, relationsDepth, false),
		}
	}
}

func (m Model) handleRelationsTree(msg relationsTreeMsg) Model {
	if msg.target != m.relationsTarget {
		return m
	}
	m.relationsLoading = false
	m.relationsText = msg.tree
	m.relationsScroll = 0
	return m
}

func (m Model) spinCmd() tea.Cmd {
	if !m.relationsLoading {
		return nil
	}
	return tea.Tick(spinner.TickInterval, func(time.Time) tea.Msg {
		return spinnerTickMsg{}
	})
}

func (m Model) handleSpinnerTick() Model {
	if m.relationsLoading {
		m.spinnerFrame = (m.spinnerFrame + 1) % len(spinner.Stages)
	}
	return m
}

func (m Model) handleRelationsKeyPress(key string) (tea.Model, tea.Cmd) {
	action, matched := m.keyMap.ResolveKey(config.ModeExplorerRelations, key)
	if matched {
		switch action {
		case config.ActionQuit:
			m.relationsMode = false
			m.relationsText = ""
			m.relationsScroll = 0
			m.relationsLoading = false
			return m, nil

		case config.ActionMoveUp:
			return m.scrollRelations(-1), nil
		case config.ActionMoveDown:
			return m.scrollRelations(1), nil
		}
	}
	return m, nil
}

func (m Model) scrollRelations(delta int) Model {
	maxScroll := max(len(m.relationsLines())-m.relationsViewportHeight(), 0)
	next := clampInt(m.relationsScroll+delta, 0, maxScroll)
	m.relationsScroll = next
	return m
}

func clampInt(v, low, high int) int {
	if v < low {
		return low
	}
	if v > high {
		return high
	}
	return v
}
