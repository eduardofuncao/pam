package styles

import "github.com/charmbracelet/lipgloss"

// Color constants
const (
	ColorAccent     = "205" // Magenta - used for titles, headers, emphasis
	ColorSuccess    = "171" // Purple - used for success messages
	ColorKeyword    = "86"  // Cyan - used for SQL keywords
	ColorString     = "220" // Yellow - used for SQL strings
	ColorFaint      = "238" // Gray - used for borders, separators, help text
	ColorHighlight  = "62"  // Dark Cyan - used for selected/highlighted backgrounds
	ColorSelected   = "230" // Light Yellow - used for selected text foreground
	ColorCellNormal = "252" // Light Gray - used for normal cell text
)

// Common reusable styles
var (
	// Title style - used for query names, headings
	Title = lipgloss.NewStyle().
		Bold(true). 
		Foreground(lipgloss.Color(ColorAccent))

	// Success style - used for confirmation messages
	Success = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorSuccess)). 
		Bold(true)

	// Error style - used for error messages
	Error = lipgloss.NewStyle(). 
		Foreground(lipgloss.Color("196")). // Red
		Bold(true)

	// Faint style - used for help text, footers, secondary info
	Faint = lipgloss.NewStyle().
		Faint(true)

	// Separator style - used for visual dividers
	Separator = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorFaint))
)

// SQL syntax highlighting styles
var (
	SQLKeyword = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorKeyword)).
		Bold(true)

	SQLString = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorString))

	// SearchMatch style - used for highlighting search term matches
	SearchMatch = lipgloss.NewStyle().
		Foreground(lipgloss.Color("226")). // Bright Yellow
		Bold(true).
		Background(lipgloss.Color("57"))   // Purple background for contrast
)

// Table component styles
var (
	TableSelected = lipgloss.NewStyle(). 
			Background(lipgloss.Color(ColorHighlight)).
			Foreground(lipgloss.Color(ColorSelected)). 
			Bold(true)

	TableHeader = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorAccent)).
			Bold(true)

	TableCell = lipgloss.NewStyle(). 
			Foreground(lipgloss.Color(ColorCellNormal))

	TableBorder = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorFaint))

	TableCopiedBlink = lipgloss.NewStyle(). 
				Background(lipgloss. Color(ColorHighlight)).
				Foreground(lipgloss. Color(ColorAccent)).
				Bold(true)

	TableUpdated = lipgloss.NewStyle().
				Background(lipgloss. Color(ColorHighlight)).
				Foreground(lipgloss. Color(ColorAccent)).
				Bold(true)

	TableDeleted = lipgloss.NewStyle().
				Background(lipgloss. Color(ColorHighlight)).
				Foreground(lipgloss. Color(ColorAccent)).
				Bold(true)
)

// Explain command styles
var (
	TableName = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorAccent)).
		Bold(true)

	PrimaryKeyLabel = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorSuccess)).
		Bold(true)

	BelongsToStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("218")). // Vivid Orange-1
		Bold(true)

	HasManyStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("86")). // Cyan
		Bold(true)

	CardinalityStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("228")). // Orange/Yellow
		Bold(true)

	TreeConnector = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorFaint))
)
