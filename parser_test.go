package main

import (
	"io/ioutil"
	"reflect"
	"testing"
)

func TestParser(t *testing.T) {
	// Create a definition file
	content := []byte("<path>{{path}}</path>")
	tmpfile, err := ioutil.TempFile("", "foo.definition")
	if err != nil {
		t.Fatalf("failed to create tmp file: %s", err)
	}
	if _, err := tmpfile.Write(content); err != nil {
		t.Fatalf("failed to write content: %s", err)
	}
	defer tmpfile.Close()

	input := make(chan string)
	errors := make(chan error)
	quit := make(chan struct{})
	out := parser(input, errors, quit, tmpfile.Name())

	for _, test := range []struct {
		input  string
		output Parsed
	}{
		{
			input: "<!-- --><path>foo.jpg</path><!-- -->",
			output: Parsed{
				Fields: []map[string]interface{}{
					{
						"path": "foo.jpg",
					},
				},
				Size: 36,
			},
		},
		{
			input: "<path>foo.jpg</path><!-- --><path>bar.jpg</path> EOF",
			output: Parsed{
				Fields: []map[string]interface{}{
					{
						"path": "foo.jpg",
					},
					{
						"path": "bar.jpg",
					},
				},
				Size: 52,
			},
		},
	} {
		input <- test.input
		output := <-out

		if !reflect.DeepEqual(output, test.output) {
			t.Errorf("Output %+v does not match expected %+v", output, test.output)
		}
	}
}
