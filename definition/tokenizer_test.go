package definition

import (
	"reflect"
	"strings"
	"testing"
)

func TestTokenize(t *testing.T) {

	for _, test := range []struct {
		definition string
		expected   *lexer
	}{
		{
			definition: strings.Join([]string{
				`text only`,
			}, "\n"),
			expected: &lexer{
				ast: []element{
					{
						token:   tokenText,
						content: `text only`,
					},
					{
						token: tokenEOF,
					},
				},
			},
		},
		{
			definition: strings.Join([]string{
				`<a href="{{someLink}}">{{someLinkText}}</a>`,
			}, "\n"),
			expected: &lexer{
				ast: []element{
					{
						token:   tokenText,
						content: `<a href="`,
					},
					{
						token:   tokenLeftMeta,
						content: `{{`,
					},
					{
						token:   tokenVariable,
						content: `someLink`,
					},
					{
						token:   tokenRightMeta,
						content: `}}`,
					},
					{
						token:   tokenText,
						content: `">`,
					},
					{
						token:   tokenLeftMeta,
						content: `{{`,
					},
					{
						token:   tokenVariable,
						content: `someLinkText`,
					},
					{
						token:   tokenRightMeta,
						content: `}}`,
					},
					{
						token:   tokenText,
						content: `</a>`,
					},
					{
						token: tokenEOF,
					},
				},
			},
		},
		{
			definition: strings.Join([]string{
				`<a href="{{someLink|filter1|filter2}}">{{someLinkText|filter3}}</a>`,
			}, "\n"),
			expected: &lexer{
				ast: []element{
					{
						token:   tokenText,
						content: `<a href="`,
					},
					{
						token:   tokenLeftMeta,
						content: `{{`,
					},
					{
						token:   tokenVariable,
						content: `someLink`,
					},
					{
						token:   tokenPipe,
						content: `|`,
					},
					{
						token:   tokenFilter,
						content: `filter1`,
					},
					{
						token:   tokenPipe,
						content: `|`,
					},
					{
						token:   tokenFilter,
						content: `filter2`,
					},
					{
						token:   tokenRightMeta,
						content: `}}`,
					},
					{
						token:   tokenText,
						content: `">`,
					},
					{
						token:   tokenLeftMeta,
						content: `{{`,
					},
					{
						token:   tokenVariable,
						content: `someLinkText`,
					},
					{
						token:   tokenPipe,
						content: `|`,
					},
					{
						token:   tokenFilter,
						content: `filter3`,
					},
					{
						token:   tokenRightMeta,
						content: `}}`,
					},
					{
						token:   tokenText,
						content: `</a>`,
					},
					{
						token: tokenEOF,
					},
				},
			},
		},
	} {
		ast := &lexer{}
		err := ast.tokenize(test.definition)
		if err != nil {
			t.Errorf("Did not expect to receive an error, got %s", err.Error())
		}

		if !reflect.DeepEqual(ast.ast, test.expected.ast) {
			t.Errorf("Expected AST to be\n%+v, got\n%+v", test.expected.ast, ast.ast)
		}
	}
}
