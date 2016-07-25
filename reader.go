package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Reader will spawn a single goroutine which will start an interactive
// terminal, asking for user input
func reader(
	inputReady chan struct{},
	errors chan<- error,
	quit chan struct{},
) chan string {

	out := make(chan string)
	reader := bufio.NewReader(os.Stdin)

	go func() {
		for {
			select {
			case <-quit:
				return
			case <-inputReady:

			Retry:
				fmt.Print("Enter a URL (type q to quit): ")
				text, err := reader.ReadString('\n')
				if err != nil {
					errors <- fmt.Errorf("error reading from stdin: %s", err)
					continue
				}

				text = strings.TrimSpace(text)

				// Skip if it was a miss-press
				if len(text) == 0 {
					// Don't wait for another input ready, just start again
					goto Retry
				}

				// If 'q' is typed, quit
				if text == "q" {
					close(quit)
				}

				out <- text
			}
		}
	}()
	return out
}
