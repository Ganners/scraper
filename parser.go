package main

import (
	"encoding/binary"
	"log"

	"github.com/ganners/scraper/definition"
)

// Parser will apply the definition to the html body, to return a series of
// keys to values
func parser(
	in <-chan string,
	errors chan<- error,
	quit chan struct{},
	definitionFile string,
) chan Parsed {

	def, err := definition.NewDefinition(definitionFile)
	if err != nil {
		log.Fatal("failed to read definition: %s", err)
	}

	out := make(chan Parsed)

	for i := 0; i < NumParserWorkers; i++ {
		go func() {
			for {
				select {
				case <-quit:
					return
				case body := <-in:
					p := Parsed{
						Fields: def.Parse(body),
						Size:   binary.Size([]byte(body)),
					}
					out <- p
				}
			}
		}()
	}
	return out
}
