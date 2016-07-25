package definition

import (
	"fmt"
	"io/ioutil"
	"strings"
	"unicode"
)

// Definition will be able to parse a byte slice and return information that we
// want to extract from it
type Definition interface {
	Parse(content string) []map[string]string
}

// A filterFunc is a function which can be used to modify a string
type filterFunc func(string) string

// DefinitionParser will contain the bytes from a definition file
type DefinitionParser struct {
	filters map[string]filterFunc
	L       *lexer
}

// NewDefinition takes a definition file and will return something that can
// return parsed variables from a byte stream, such as HTML
func NewDefinition(definitionFile string) (*DefinitionParser, error) {
	b, err := ioutil.ReadFile(definitionFile)
	if err != nil {
		return nil, fmt.Errorf("Error opening definition file: %s", err)
	}

	ast := &lexer{}
	err = ast.tokenize(string(b))
	if err != nil {
		return nil, fmt.Errorf("Error tokenizing definition file: %s", err)
	}

	return &DefinitionParser{
		L:       ast,
		filters: filters, // Apply local default filters only
	}, nil
}

// Parse will apply a definition file to some content (probably HTML) and
// return the variables contained. This makes the assumption that the block can
// be repeated, hence the slice
//
// This localises all variables and so is thread safe
func (def *DefinitionParser) Parse(content string) []map[string]string {

	data := make([]map[string]string, 0, 10)

	// Nothing to do
	if len(def.L.ast) == 0 {
		return data
	}

	pos := 0
	tokenIndex := 0
	fields := make(map[string]string, 10)

	variableStart := 0
	variableName := ""
	variableTokenIndex := 0

	for {
		currentToken := def.L.ast[tokenIndex]
		tokenContent := currentToken.content

		switch currentToken.token {

		default:
			// Not a match, reset
			fallthrough
		case tokenEOF:
			// EOF means we want to retry applying our definition to
			// newly seen text to see if we can get multiple
			data = append(data, fields)
			fields = make(map[string]string, 10)
			tokenIndex = 0
			pos -= 1
			continue
		case tokenText:
			// If there is some token text, we care about previous
			// state and can just determine if we're looking into a
			// variable or not and handle it in different ways
			if p, o := HasPrefixIgnoreWhitespace(content[pos:], tokenContent); p {

				// If we're looking to fill in a variable name
				if len(variableName) > 0 {

					// Do nothing
					if variableName == "_" {
						variableStart = 0
						variableName = ""
						variableTokenIndex = 0
						continue
					}

					// While there are no more filters
					filtersToApply := make([]string, 0, 3)
					for j := 1; ; j++ {
						if variableTokenIndex+j >= len(def.L.ast) {
							break
						}
						tok := def.L.ast[variableTokenIndex+j]
						if tok.token == tokenPipe {
							continue
						} else if tok.token == tokenFilter {
							filtersToApply = append(filtersToApply, tok.content)
						} else {
							break
						}
					}
					fields[variableName] = content[variableStart-1 : pos]

					// Apply any filters
					for _, filter := range filtersToApply {
						filterFunc, found := def.filters[filter]
						if found {
							fields[variableName] = filterFunc(fields[variableName])
						}
					}

					// Reset variable
					variableStart = 0
					variableName = ""
					variableTokenIndex = 0
				}
				pos += o
				tokenIndex++
			}
		case tokenFilter,
			tokenPipe,
			tokenRightMeta:
			tokenIndex++
		case tokenLeftMeta:
			// Can optimize and fall through to the proceeding state
			tokenIndex++
			fallthrough
		case tokenVariable:
			// Keep track of our position and the name of the variable
			variableStart = pos
			variableName = def.L.ast[tokenIndex].content
			variableTokenIndex = tokenIndex
			tokenIndex++
		}

		pos++
		if pos >= len(content) {
			break
		}
	}
	return data
}

// Means we can be whitespace agnostic, also means we'll only accept
// values that don't contain spaces (or we're happy for those spaces to
// be stripped)
func stripWhitespace(str string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			// if the character is a space, drop it
			return -1
		}
		// else keep it in the string
		return r
	}, str)
}

// Checks if a string has the prefix, ignoring all spaces. Returns the position
// of the last character of that original prefix
//
// Ignores the space within both the str and the prefix
func HasPrefixIgnoreWhitespace(s, prefix string) (bool, int) {
	j := 0
	i := 0
	for {
		if i >= len(prefix) {
			break
		}
		if j >= len(s) {
			return false, 0
		}
		if s[j] == prefix[i] {
			j++
			i++
		} else if isWhitespace(s[j]) {
			j++
		} else if isWhitespace(prefix[i]) {
			i++
		} else {
			return false, 0
		}
	}
	return true, j
}

// Quick whitespace check on a byte
func isWhitespace(r byte) bool {
	if r == ' ' || r == '\n' || r == '\r' || r == '\t' {
		return true
	}
	return false
}
