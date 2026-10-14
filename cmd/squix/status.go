package main

import (
	"fmt"
	"time"

	"github.com/eduardofuncao/squix/internal/db"
	"github.com/eduardofuncao/squix/internal/spinner"
	"github.com/eduardofuncao/squix/internal/styles"
)

func (a *App) printConnStatus(conn db.DatabaseConnection) {
	currConn := a.config.Connections[a.config.CurrentConnection]
	connInfo := fmt.Sprintf("%s/%s", currConn.DBType, currConn.Name)
	if currConn.Schema != "" {
		connInfo += fmt.Sprintf(" (schema: %s)", currConn.Schema)
	}

	queryCount := len(a.config.QueriesFor(a.config.CurrentConnection))

	interactive := spinner.Interactive()
	if interactive {
		// Provisional header, erased and reprinted with status below.
		fmt.Printf("Using %s\n", styles.Title.Render(connInfo))
	}

	done := make(chan struct{})
	reachable := make(chan bool)

	go func() {
		err := conn.Ping()
		reachable <- (err == nil)
	}()

	go spinner.CircleWait(done)

	var isReachable bool
	select {
	case isReachable = <-reachable:
		spinner.Stop(done)
	case <-time.After(5 * time.Second):
		spinner.Stop(done)
		isReachable = false
	}

	if interactive {
		// Clear the header line above the spinner (Stop already cleared the
		// spinner line itself).
		fmt.Print("\033[1A")
		fmt.Print("\r\033[2K")
	}

	circleIcon := "●"
	statusText := "reachable"
	if !isReachable {
		circleIcon = "○"
		statusText = "unreachable"
	}

	fmt.Printf("%s Using %s\n", styles.Success.Render(circleIcon), styles.Title.Render(connInfo))
	fmt.Printf("  %d saved queries, %s\n", queryCount, styles.Faint.Render(statusText))
}

func (a *App) handleStatus() {
	if a.config.CurrentConnection == "" {
		fmt.Println(styles.Faint.Render("No active connection"))
		return
	}

	conn := a.config.LiveConnection(a.config.CurrentConnection)
	conn.Open()
	defer conn.Close()

	a.printConnStatus(conn)
}
