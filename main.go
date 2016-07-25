// This handles pretty much everything, it is pipeline and worker based meaning
// that data gets sent between workers via channels in a daisy-chain of
// goroutines.
package main

import (
	"encoding/json"
	"fmt"
	"log"
)

const (
	// For some processes it makes sense to spawn a number of works who can
	// handle those requests. 4 is just an arbitrary power of 2
	NumGetterWorkers    = 4
	NumParserWorkers    = 4
	NumPresenterWorkers = 4

	// The definition file for which to apply. This particular one works for
	// the following URL:
	ListDefinition    = "definitions/sainsburys-list.definition"
	ProductDefinition = "definitions/sainsburys-product.definition"
)

var (
	// The re-usable thread-safe reader
	// The HttpReader is the most simple
	// The GoogleCacheReader is useful for grabbing the live site's source
	DefaultWebReader WebReader = NewHttpReader()
)

// Parsed represents the fields and the body size of the page that has
// been returned
type Parsed struct {
	Fields []map[string]interface{}
	Size   int
}

func main() {

	// errors will exit the program if an error is received
	errors := make(chan error)

	// quit will exit all running goroutine (should be called with
	// close())
	quit := make(chan struct{})

	// We can signal when we're ready to take new input
	inputReady := make(chan struct{})

	// Orchestrate the pipeline
	input := reader(inputReady, errors, quit)
	webContent := getter(input, errors, quit)
	parsedContent := parser(webContent, errors, quit, ListDefinition)
	printable := sainsburysFormatter(parsedContent, errors, quit)

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

// The presenter will convert the map into a string which can be printed, it
// will not print itself as it might be better as a syncronous process on the
// main goroutine
//
// This particular worker is tied specifically to the Sainsburys
// definition file for the price calculation. It will gracefully handle
// missing fields.
//
// It also handles the fetching of child page descriptions
func sainsburysFormatter(
	in <-chan Parsed,
	errors chan<- error,
	quit chan struct{},
) chan string {

	out := make(chan string)

	for i := 0; i < NumPresenterWorkers; i++ {
		go func() {

			// Set up getter and presenter, these will be a child pipeline of
			// the presenter
			//
			// Each worker gets it's own pipeline so it can be used synchrously
			// and in order.
			subIn := make(chan string)
			descriptionPageGetter := getter(subIn, errors, quit)
			descriptionParser := parser(descriptionPageGetter, errors, quit, ProductDefinition)

			for {
				select {
				case <-quit:
					return
				case parsed := <-in:

					// Presentation layer. This will look at all of the
					// fields and run some processing which will
					// calculate the total per unit and measure, and
					// also we can use that information to define the
					// size which we add on.
					presentation := &struct {
						Products     []map[string]interface{} `json:"products"`
						TotalUnit    int                      `json:"totalUnitPrice"`
						TotalMeasure int                      `json:"totalMeasurePrice"`
						NumProducts  int                      `json:"numProducts"`
					}{
						Products:     parsed.Fields,
						TotalUnit:    0,
						TotalMeasure: 0,
						NumProducts:  len(parsed.Fields),
					}

					// Sum units and measures
					for i, product := range parsed.Fields {

						// Grab the description (wait for one element, making
						// use synchronously)
						subIn <- product["productPath"].(string)
						description := <-descriptionParser

						if len(description.Fields) == 1 {
							parsed.Fields[i]["description"] = description.Fields[0]["description"]
							parsed.Fields[i]["size"] = float64(description.Size) / 1024
						}

						pricePerMeasure, ok := product["pricePerMeasure"].(int)
						if !ok {
							continue
						}
						pricePerUnit, ok := product["pricePerUnit"].(int)
						if !ok {
							continue
						}
						presentation.TotalUnit += pricePerUnit
						presentation.TotalMeasure += pricePerMeasure

						// Set the number of units (the quantity)
						parsed.Fields[i]["quantity"] = pricePerUnit / pricePerMeasure
					}

					b, err := json.Marshal(presentation)
					if err != nil {
						errors <- fmt.Errorf("unable to marshal presentation into json: %s", err)
					}

					// Just print it out
					out <- string(b)
				}
			}
		}()
	}
	return out
}
