package table

import (
	"strings"
	"testing"

	"github.com/eduardofuncao/squix/internal/config"
)

func TestCellMatchProgress(t *testing.T) {
	// Progress comes from the stored index, not the live cursor.
	m := Model{searchMatches: []CellPosition{{0, 0}, {1, 2}, {3, 1}}}

	m.cellMatchIdx = 1
	if cur, tot := m.cellMatchProgress(); cur != 2 || tot != 3 {
		t.Errorf("idx=1: got (%d,%d), want (2,3)", cur, tot)
	}

	m.cellMatchIdx = 0
	if cur, tot := m.cellMatchProgress(); cur != 1 || tot != 3 {
		t.Errorf("idx=0: got (%d,%d), want (1,3)", cur, tot)
	}

	m.searchMatches = nil
	if cur, tot := m.cellMatchProgress(); cur != 0 || tot != 0 {
		t.Errorf("empty: got (%d,%d), want (0,0)", cur, tot)
	}
}

func TestColMatchProgress(t *testing.T) {
	m := Model{searchColMatches: []int{0, 3, 5}}

	m.colMatchIdx = 2
	if cur, tot := m.colMatchProgress(); cur != 3 || tot != 3 {
		t.Errorf("idx=2: got (%d,%d), want (3,3)", cur, tot)
	}

	m.colMatchIdx = 0
	if cur, tot := m.colMatchProgress(); cur != 1 || tot != 3 {
		t.Errorf("idx=0: got (%d,%d), want (1,3)", cur, tot)
	}
}

func TestMatchLabel(t *testing.T) {
	if got := matchLabel(3, 12); got != "[3/12]" {
		t.Errorf("got %q, want %q", got, "[3/12]")
	}
	if got := matchLabel(1, 1); got != "[1/1]" {
		t.Errorf("got %q, want %q", got, "[1/1]")
	}
}

func TestCyclerSetsCellMatchIdx(t *testing.T) {
	m := Model{searchMatches: []CellPosition{{0, 0}, {0, 2}, {2, 0}}}

	// From match 0 → next lands on match 1 ({0,2}).
	m.selectedRow, m.selectedCol = 0, 0
	m = m.nextSearchMatch()
	if m.cellMatchIdx != 1 || m.selectedRow != 0 || m.selectedCol != 2 {
		t.Errorf("next from 0: idx=%d pos=(%d,%d), want idx=1 pos=(0,2)", m.cellMatchIdx, m.selectedRow, m.selectedCol)
	}

	// → next lands on match 2 ({2,0}).
	m = m.nextSearchMatch()
	if m.cellMatchIdx != 2 || m.selectedRow != 2 || m.selectedCol != 0 {
		t.Errorf("next from 1: idx=%d pos=(%d,%d), want idx=2 pos=(2,0)", m.cellMatchIdx, m.selectedRow, m.selectedCol)
	}

	// → wrap to match 0.
	m = m.nextSearchMatch()
	if m.cellMatchIdx != 0 {
		t.Errorf("wrap next: idx=%d, want 0", m.cellMatchIdx)
	}

	// prev from match 0 → wrap to last (match 2).
	m = m.prevSearchMatch()
	if m.cellMatchIdx != 2 {
		t.Errorf("wrap prev: idx=%d, want 2", m.cellMatchIdx)
	}
}

func TestCyclerSetsColMatchIdx(t *testing.T) {
	m := Model{searchColMatches: []int{1, 3, 5}}

	m.selectedCol = 1
	m = m.nextColumnMatch()
	if m.colMatchIdx != 1 || m.selectedCol != 3 {
		t.Errorf("next col: idx=%d col=%d, want idx=1 col=3", m.colMatchIdx, m.selectedCol)
	}

	// prev from col 3 → match 0 (col 1).
	m = m.prevColumnMatch()
	if m.colMatchIdx != 0 || m.selectedCol != 1 {
		t.Errorf("prev col: idx=%d col=%d, want idx=0 col=1", m.colMatchIdx, m.selectedCol)
	}

	// prev from col 1 → wrap to last (col 5, idx 2).
	m = m.prevColumnMatch()
	if m.colMatchIdx != 2 || m.selectedCol != 5 {
		t.Errorf("wrap prev col: idx=%d col=%d, want idx=2 col=5", m.colMatchIdx, m.selectedCol)
	}
}

func TestSelectionBoundsCounts(t *testing.T) {
	m := Model{
		columns:        []string{"a", "b", "c", "d"},
		visualMode:     true,
		visualStartRow: 1,
		visualStartCol: 0,
		selectedRow:    3,
		selectedCol:    2,
	}

	// Characterwise: rows 1..3, cols 0..2 → 3x3.
	minR, maxR, minC, maxC := m.getSelectionBounds()
	if rows, cols := maxR-minR+1, maxC-minC+1; rows != 3 || cols != 3 {
		t.Errorf("charwise: got %dx%d, want 3x3", rows, cols)
	}

	// Linewise: columns span all of them (4) regardless of start col.
	m.visualLineMode = true
	_, _, minC, maxC = m.getSelectionBounds()
	if cols := maxC - minC + 1; cols != 4 {
		t.Errorf("linewise cols: got %d, want 4 (all columns)", cols)
	}
}

func TestRenderFooter_SearchCounter(t *testing.T) {
	m := Model{
		data:          [][]string{{"a", "b"}, {"c", "d"}},
		columns:       []string{"c1", "c2"},
		selectedRow:   0,
		selectedCol:   0,
		searchMatches: []CellPosition{{0, 0}, {1, 1}},
		cellMatchIdx:  0,
		uiVisibility:  config.UIVisibility{FooterCellContent: true, FooterStats: true},
		width:         80,
	}
	out := m.renderFooter()
	if !strings.Contains(out, "1,1") {
		t.Errorf("cursor restyle: missing '1,1' in:\n%s", out)
	}
	if !strings.Contains(out, "[1/2]") {
		t.Errorf("search counter: missing '[1/2]' in:\n%s", out)
	}

	// Frozen: cursor moved off any match, but the stored index is unchanged →
	// counter still shows [2/2], never a plain total.
	m.selectedRow, m.selectedCol = 5, 5
	m.cellMatchIdx = 1
	if out := m.renderFooter(); !strings.Contains(out, "[2/2]") {
		t.Errorf("frozen counter: missing '[2/2]' in:\n%s", out)
	}
}

func TestRenderFooter_VisualCounter(t *testing.T) {
	m := Model{
		data:           [][]string{{"a", "b"}, {"c", "d"}, {"e", "f"}, {"g", "h"}, {"i", "j"}},
		columns:        []string{"c1", "c2", "c3", "c4"},
		selectedRow:    1,
		selectedCol:    1,
		visualMode:     true,
		visualStartRow: 0,
		visualStartCol: 0,
		uiVisibility:   config.UIVisibility{FooterCellContent: true},
		width:          80,
	}
	// Characterwise: rows 0..1 (2) x cols 0..1 (2).
	if out := m.renderFooter(); !strings.Contains(out, "-- VISUAL 2x2 --") {
		t.Errorf("charwise counter: missing '-- VISUAL 2x2 --' in:\n%s", out)
	}
	// Linewise: rows 0..1 (2) x all 4 columns.
	m.visualLineMode = true
	if out := m.renderFooter(); !strings.Contains(out, "-- V-LINE 2x4 --") {
		t.Errorf("linewise counter: missing '-- V-LINE 2x4 --' in:\n%s", out)
	}
}
