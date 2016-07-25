package main

import "fmt"

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

					if body == "" {
						errors <- fmt.Errorf("Body was empty")
					}

					out <- body
				}
			}
		}()
	}
	return out
}
