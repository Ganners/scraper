package definition

import (
	"reflect"
	"strings"
	"testing"
)

func TestParseDefinition(t *testing.T) {

	for _, test := range []struct {
		definition string
		content    string
		expected   []map[string]string
	}{
		{
			definition: strings.Join([]string{
				`<a href="{{someLink}}">{{someLinkText}}</a>`,
			}, "\n"),
			content: strings.Join([]string{
				`This should be skipped`,
				`<a href="SomeLink1">SomeLinkText1</a>`,
				`<a href="SomeLink2">SomeLinkText2</a>`,
				`This should be skipped`,
			}, "\n"),
			expected: []map[string]string{
				{
					"someLink":     "SomeLink1",
					"someLinkText": "SomeLinkText1",
				},
				{
					"someLink":     "SomeLink2",
					"someLinkText": "SomeLinkText2",
				},
			},
		},
		{
			definition: strings.Join([]string{
				`<a href="{{someLink|lowercase}}">{{someLinkText|lowercase|uppercase}}</a>`,
			}, "\n"),
			content: strings.Join([]string{
				`This should be skipped`,
				`<a href="SomeLink1">SomeLinkText1</a>`,
				`<a href="SomeLink2">SomeLinkText2</a>`,
				`This should be skipped`,
			}, "\n"),
			expected: []map[string]string{
				{
					"someLink":     "somelink1",
					"someLinkText": "SOMELINKTEXT1",
				},
				{
					"someLink":     "somelink2",
					"someLinkText": "SOMELINKTEXT2",
				},
			},
		},
	} {
		parser := &DefinitionParser{
			filters: filters,
		}
		ast := &lexer{}
		ast.tokenize(test.definition)
		parser.L = ast

		vars := parser.Parse(test.content)

		if !reflect.DeepEqual(vars, test.expected) {
			t.Errorf("Expected vars to be %+v, got %+v", test.expected, vars)
		}
	}
}

func TestHasPrefixIgnoreWhitespace(t *testing.T) {
	for _, test := range []struct {
		str       string
		prefix    string
		hasPrefix bool
		offset    int
	}{
		{
			str:       "This is a test",
			prefix:    "T h i s",
			hasPrefix: true,
			offset:    4,
		},
		{
			str:       "<a\nhref = 'something'\n></a>",
			prefix:    "<a  href=' something '",
			hasPrefix: true,
			offset:    21,
		},
	} {
		hasPrefix, offset := HasPrefixIgnoreWhitespace(test.str, test.prefix)
		if hasPrefix != test.hasPrefix {
			t.Errorf("Expected hasPrefix to be %t, got %t", test.hasPrefix, hasPrefix)
		}
		if offset != test.offset {
			t.Errorf("Expected offset to be %d, got %d", test.offset, offset)
		}
	}
}
