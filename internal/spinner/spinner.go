package spinner

import (
	"fmt"
	"time"

	"github.com/eduardofuncao/pam/internal/styles"
)

func Wait(done chan struct{}) {
	spinnerStages := []string{"▉", "▊", "▋", "▌", "▍", "▎", "▏", "▎", "▍", "▌", "▋", "▊", "▉"}
	var passed time.Duration = 0
	for {
		for _, s := range spinnerStages {
			select {
			case <-done:
				fmt.Print("\r")
				return
			default:
				fmt.Printf("\r%s %.2fs", s, passed.Seconds())
				passed += 100 * time.Millisecond
				time.Sleep(100 * time.Millisecond)
			}
		}
	}
}

func CircleWait(done chan struct{}) {
	// Custom pulsing animation
	stages := []string{" ", ".", "o", "O", "@", "*"}
	for {
		for _, s := range stages {
			select {
			case <-done:
				fmt.Print("\r")
				return
			default:
				fmt.Printf("\r%s Checking...", styles.Success.Render(s))
				time.Sleep(100 * time.Millisecond)
			}
		}
	}
}

func CircleWaitWithTimer(done chan struct{}) {
	// Custom pulsing animation with timer
	stages := []string{" ", ".", "o", "O", "@", "*"}
	var passed time.Duration = 0
	for {
		for _, s := range stages {
			select {
			case <-done:
				fmt.Print("\r")
				return
			default:
				fmt.Printf("\r%s %.2fs", styles.Success.Render(s), passed.Seconds())
				passed += 100 * time.Millisecond
				time.Sleep(100 * time.Millisecond)
			}
		}
	}
}
