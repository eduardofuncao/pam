package explorer

import (
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/eduardofuncao/squix/internal/config"
	"github.com/eduardofuncao/squix/internal/db"
)

type section struct {
	title  string
	items  []string
	folded bool
}

type rowKind int

const (
	rowHeader rowKind = iota
	rowItem
)

type row struct {
	kind    rowKind
	section int
	item    int
}

type PendingAction int

const (
	ActionNone PendingAction = iota
	ActionSelectAll
	ActionColumns
)

type Model struct {
	width  int
	height int

	sections []section
	rows     []row
	cursor   int

	offsetY     int
	visibleRows int

	searchMode   bool
	searchQuery  string
	searchCursor int

	searchMatches []int
	matchIdx      int

	spinnerFrame       int
	relationsMode      bool
	relationsTarget    string
	relationsText      string
	relationsScroll    int
	relationsLoading   bool
	relationsLoadStart time.Time

	helpOverlay bool

	conn         db.DatabaseConnection
	uiVisibility config.UIVisibility
	keyMap       *config.KeyMap

	pendingAction PendingAction
	selectedTable string

	statusMessage string
}

func New(
	conn db.DatabaseConnection,
	tables, views []string,
	restoreCursor int,
	visibility config.UIVisibility,
	keyMap *config.KeyMap,
) Model {
	if keyMap == nil {
		keyMap = config.BuildKeyMap(nil)
	}

	m := Model{
		sections: []section{
			{title: "Tables", items: dedupeSorted(tables)},
			{title: "Views", items: dedupeSorted(views)},
		},
		visibleRows:  10,
		conn:         conn,
		uiVisibility: visibility,
		keyMap:       keyMap,
	}

	m.rebuildRows()
	if restoreCursor > 0 && restoreCursor < len(m.rows) {
		m.cursor = restoreCursor
	}
	m.clampOffset()
	return m
}

func (m Model) Init() tea.Cmd {
	return tea.ClearScreen
}

func (m Model) PendingAction() PendingAction {
	return m.pendingAction
}

func (m Model) SelectedTable() string {
	return m.selectedTable
}

func (m Model) Cursor() int {
	return m.cursor
}

func (m Model) numRows() int {
	return len(m.rows)
}

func (m Model) selectedItemName() (string, bool) {
	if m.cursor < 0 || m.cursor >= len(m.rows) {
		return "", false
	}
	r := m.rows[m.cursor]
	if r.kind != rowItem {
		return "", false
	}
	return m.sections[r.section].items[r.item], true
}

func (m Model) currentSectionIdx() int {
	if m.cursor < 0 || m.cursor >= len(m.rows) {
		return -1
	}
	return m.rows[m.cursor].section
}

func dedupeSorted(items []string) []string {
	seen := make(map[string]bool, len(items))
	out := make([]string, 0, len(items))
	for _, it := range items {
		name := strings.TrimSpace(it)
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true
		out = append(out, name)
	}
	return out
}

func (m *Model) rebuildRows() {
	var rows []row
	for si, sec := range m.sections {
		if len(sec.items) == 0 {
			continue
		}

		rows = append(rows, row{kind: rowHeader, section: si})
		if !sec.folded {
			for ii := range sec.items {
				rows = append(rows, row{kind: rowItem, section: si, item: ii})
			}
		}
	}

	m.rows = rows
	m.clampCursor()
}
