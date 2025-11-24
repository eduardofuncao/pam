package table

import "github.com/charmbracelet/lipgloss"

const (
	colorSelectedBg   = "62"
	colorSelectedFg   = "230"
	colorHeader       = "205"
	colorCell         = "252"
	colorBorder       = "238"
	colorNull         = "240"
	colorCopiedBlink  = "205"
	colorSuccess      = "40"
	colorError        = "196"
	colorKeyHighlight = "205"
	colorNormal       = "252"
)

var (
	selectedStyle = lipgloss.NewStyle().
			Background(lipgloss.Color(colorSelectedBg)).
			Foreground(lipgloss.Color(colorSelectedFg)).
			Bold(true)

	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorHeader)).
			Bold(true)

	cellStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorCell))

	nullStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorNull))

	borderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorBorder))

	copiedBlinkStyle = lipgloss.NewStyle().
				Background(lipgloss.Color(colorSelectedBg)).
				Foreground(lipgloss.Color(colorCopiedBlink)).
				Bold(true)
)
