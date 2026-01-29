package styles

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
)

var ActiveScheme ColorScheme

// Style variables used throughout the application
var (
	Title, Success, Error, Faint, Separator lipgloss.Style
	SQLKeyword, SQLString, SearchMatch lipgloss.Style
	TableSelected, TableHeader, TableCell, TableBorder lipgloss.Style
	TableCopiedBlink, TableUpdated, TableDeleted lipgloss.Style
	TableName, PrimaryKeyLabel lipgloss.Style
	BelongsToStyle, HasManyStyle, HasOneStyle, HasManyToManyStyle, CardinalityStyle, TreeConnector lipgloss.Style
)

// InitScheme initializes the color scheme from config
func InitScheme(schemeName string, custom *ColorScheme) {
	var scheme ColorScheme

	if custom != nil {
		scheme = *custom
		if scheme.Accent == "" {
			scheme.Accent = DefaultScheme.Accent
			fmt.Fprintf(os.Stderr, "Warning: Incomplete custom color scheme, using defaults for missing values\n")
		}
	} else {
		scheme = GetScheme(schemeName)
	}

	ActiveScheme = scheme
	reloadAllStyles()
}

// reloadAllStyles updates all style variables based on ActiveScheme
func reloadAllStyles() {
	Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(ActiveScheme.Primary))

	Success = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ActiveScheme.Success)).
		Bold(true)

	Error = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ActiveScheme.Error)).
		Bold(true)

	Faint = lipgloss.NewStyle().
		Faint(true)

	Separator = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ActiveScheme.Muted))

	SQLKeyword = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ActiveScheme.Accent)).
		Bold(true)

	SQLString = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ActiveScheme.Success))

	SearchMatch = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ActiveScheme.Accent)).
		Bold(true).
		Background(lipgloss.Color(ActiveScheme.Highlight))

	TableSelected = lipgloss.NewStyle().
		Background(lipgloss.Color(ActiveScheme.Highlight)).
		Foreground(lipgloss.Color(ActiveScheme.Normal)).
		Bold(true)

	TableHeader = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ActiveScheme.Primary)).
		Bold(true)

	TableCell = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ActiveScheme.Normal))

	TableBorder = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ActiveScheme.Muted))

	TableCopiedBlink = lipgloss.NewStyle().
		Background(lipgloss.Color(ActiveScheme.Highlight)).
		Foreground(lipgloss.Color(ActiveScheme.Primary)).
		Bold(true)

	TableUpdated = lipgloss.NewStyle().
		Background(lipgloss.Color(ActiveScheme.Highlight)).
		Foreground(lipgloss.Color(ActiveScheme.Success)).
		Bold(true)

	TableDeleted = lipgloss.NewStyle().
		Background(lipgloss.Color("52")).
		Foreground(lipgloss.Color(ActiveScheme.Error)).
		Bold(true)

	// Explain command & metadata styles
	TableName = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ActiveScheme.Primary)).
		Bold(true)

	PrimaryKeyLabel = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ActiveScheme.Success)).
		Bold(true)

	BelongsToStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ActiveScheme.Accent)).
		Bold(true)

	HasManyStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ActiveScheme.Accent)).
		Bold(true)

	HasOneStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ActiveScheme.Success)).
		Bold(true)

	HasManyToManyStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("13")). // Purple for N:N
		Bold(true)

	CardinalityStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ActiveScheme.Accent)).
		Bold(true)

	TreeConnector = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ActiveScheme.Muted))
}
