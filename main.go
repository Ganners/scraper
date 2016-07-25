// This handles pretty much everything, it is pipeline and worker based meaning
// that data gets sent between workers via channels in a daisy-chain of
// goroutines.
package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/ganners/scraper/definition"
)

const (
	// For some processes it makes sense to spawn a number of works who can
	// handle those requests. 4 is just an arbitrary power of 2
	NumGetterWorkers    = 4
	NumParserWorkers    = 4
	NumPresenterWorkers = 4

	// The definition file for which to apply. This particular one works for
	// the following URL:
	// > http://www.sainsburys.co.uk/shop/gb/groceries/fruit-veg/ripe---ready
	DefinitionFile = "sainsburys-cache.definition"
)

var (
	// The re-usable thread-safe reader, default to GoogleCacheReader, it'll
	// give us an SEO friendly version
	DefaultWebReader WebReader = NewGoogleCacheReader()
)

func main() {

	errors := make(chan error)
	quit := make(chan struct{})

	// We can signal when we're ready to take new input
	inputReady := make(chan struct{})

	// Orchestrate the pipeline
	input := reader(inputReady, errors, quit)
	webContent := getter(input, errors, quit)
	parsedContent := parser(webContent, errors, quit)
	printable := presenter(parsedContent, errors, quit)

	go func() {
		// Listen to errors and kill the application if one comes in Possibly
		// not the desired behaviour in the real world but works here
		err := <-errors
		close(quit)
		log.Fatalf("Error: %s", err)
	}()

	go func() {
		// Print out anything which comes back from the printable str
		for str := range printable {
			fmt.Println(str)
			inputReady <- struct{}{} // Ask for another URL
		}
	}()

	// Ask for the initial input
	inputReady <- struct{}{}

	// Terminate when quit is fired
	<-quit
}

// Reader will spawn a single goroutine which will start an interactive
// terminal, asking for user input
func reader(
	inputReady chan struct{},
	errors chan<- error,
	quit chan struct{},
) chan string {
	out := make(chan string)
	go func() {
		for {
			reader := bufio.NewReader(os.Stdin)
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

// Getter will take a URL input and perform some action to grab the contents of
// that web page. There are a number of strategies available for this.
func getter(
	in <-chan string,
	errors chan<- error,
	quit chan struct{},
) chan string {
	out := make(chan string)
	for i := 0; i < NumGetterWorkers; i++ {
		go func() {
			for {
				select {
				case <-quit:
					return
				case url := <-in:
					body, err := DefaultWebReader.GetBody(url)
					if err != nil {
						errors <- fmt.Errorf("could not read url: %s", err)
					}
					out <- body
				}
			}
		}()
	}
	return out
}

// Parser will apply the definition to the html body, to return a series of
// keys to values
func parser(
	in <-chan string,
	errors chan<- error,
	quit chan struct{},
) chan []map[string]string {
	out := make(chan []map[string]string)

	for i := 0; i < NumParserWorkers; i++ {
		go func() {
			// My definer isn't thread safe at the moment, it could be however
			// with a functional parser
			def, err := definition.NewDefinition(DefinitionFile)
			if err != nil {
				log.Fatal("failed to read definition: %s", err)
			}
			for {
				select {
				case <-quit:
					return
				case body := <-in:
					out <- def.Parse(body)
				}
			}
		}()
	}
	return out
}

// The presenter will convert the map into a string which can be printed, it
// will not print itself as it might be better as a syncronous process on the
// main goroutine
func presenter(
	in <-chan []map[string]string,
	errors chan<- error,
	quit chan struct{},
) chan string {
	out := make(chan string)
	for i := 0; i < NumPresenterWorkers; i++ {
		go func() {
			for {
				select {
				case <-quit:
					return
				case fields := <-in:
					// Present it! Using a string is convenient so would lead
					// to lots of allocations due to the immutibility of
					// strings. Better to use a byte slice
					str := ""
					for _, row := range fields {
						for fieldName, value := range row {
							str += fieldName + ": " + value + "\n"
						}
						str += "\n"
					}

					// Just print it out
					out <- str
				}
			}
		}()
	}
	return out
}
