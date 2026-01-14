package main

import (
	"fmt"

	"github.com/eduardofuncao/pam/internal/config"
	"github.com/eduardofuncao/pam/internal/spinner"
	"github.com/eduardofuncao/pam/internal/styles"
)

func (a *App) handleStatus() {
	if a.config.CurrentConnection == "" {
		fmt.Println(styles.Faint.Render("No active connection"))
		return
	}

	currConn := a.config.Connections[a.config.CurrentConnection]

	connInfo := fmt.Sprintf("%s/%s", currConn.DBType, currConn.Name)
	if currConn.Schema != "" {
		connInfo += fmt.Sprintf(" (schema: %s)", currConn.Schema)
	}

	queryCount := 0
	if currConn.Queries != nil {
		queryCount = len(currConn.Queries)
	}

	fmt.Printf("Using %s\n", styles.Title.Render(connInfo))

	done := make(chan struct{})
	reachable := make(chan bool)

	go func() {
		conn := config.FromConnectionYaml(currConn)
		conn.Open()
		defer conn.Close()

		err := conn.Ping()
		reachable <- (err == nil)
	}()

	go spinner.CircleWait(done)

	isReachable := <-reachable
	close(done)

	fmt.Print("\r\033[2K") // Clear current line
	fmt.Print("\033[1A")   // Move up one line
	fmt.Print("\r\033[2K") // Clear that line too

	circleIcon := "●"
	if !isReachable {
		circleIcon = "○"
	}

	statusText := "reachable"
	if !isReachable {
		statusText = "unreachable"
	}

	// Print final output
	fmt.Printf("%s Using %s\n", styles.Success.Render(circleIcon), styles.Title.Render(connInfo))
	fmt.Printf("  %d saved queries, %s\n", queryCount, styles.Faint.Render(statusText))
}
