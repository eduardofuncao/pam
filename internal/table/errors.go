package table

import (
	"fmt"
	"os"
	"github.com/eduardofuncao/pam/internal/styles"
)

func printError(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintln(os.Stderr, styles.Error.Render("✗ Error:"), msg)
	os.Exit(1)
}
