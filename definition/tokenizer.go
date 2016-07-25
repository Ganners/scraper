package definition

import (
	"log"
	"strings"
)

const (
	// EOF is rune 0, it is used to determine termination
	eof       rune = 0
	leftMeta       = "{{"
	rightMeta      = "}}"
	pipe           = '|'
)

// A token represents a lexical type
type token int8

const (
	tokenText token = iota
	tokenSkip
	tokenLeftMeta
	tokenRightMeta
	tokenVariable
	tokenPipe
	tokenFilter
	tokenEOF
	tokenError
)

// An element consists of a token (lexical type) and the content contained
// within it
type element struct {
	token   token
	content string
}

// The lexer contains a flat AST
type lexer struct {
	content string
	ast     []element

	start int
	pos   int
}

// A stateFunc represents the state of a scanner, and returns a
// function of the next state
type stateFunc func(*lexer) stateFunc

// Errorf will print out an error and terminate the state
func errorf(err string) stateFunc {
	return func(l *lexer) stateFunc {
		log.Printf("[Error] Error: %s", err)
		l.emit(tokenError)
		return nil
	}
}

// The filter state is a lot like the variable state, it will emit a filter and
// allow another filter to be applied, if not the end of the variable
func filterState(l *lexer) stateFunc {
	for {
		if strings.HasPrefix(l.content[l.pos:], rightMeta) {
			if l.pos > l.start {
				l.emit(tokenFilter)
			}
			return rightMetaState
		}
		r := l.next()
		if r == eof || r == '\n' || r == ' ' {
			return errorf("undisclosed action")
		}
		if r == pipe {
			if l.pos > l.start {
				l.emit(tokenFilter)
			}
			return pipeState
		}
	}
}

// The pipeState preceeds a filter
func pipeState(l *lexer) stateFunc {
	l.pos += 1
	l.emit(tokenPipe)
	return filterState
}

// VariableState will look for a right meta or a pipe, most other things are
// not allowed
func variableState(l *lexer) stateFunc {
	for {
		if strings.HasPrefix(l.content[l.pos:], rightMeta) {
			if l.pos > l.start {
				l.emit(tokenVariable)
			}
			return rightMetaState
		}

		r := l.next()
		if r == eof || r == '\n' || r == ' ' {
			return errorf("undisclosed action")
		}
		if r == pipe {
			if l.pos > l.start {
				l.emit(tokenVariable)
			}
			return pipeState
		}
	}
}

func rightMetaState(l *lexer) stateFunc {
	l.pos += len(rightMeta)
	l.emit(tokenRightMeta)
	return textState
}

func leftMetaState(l *lexer) stateFunc {
	l.pos += len(leftMeta)
	l.emit(tokenLeftMeta)
	return variableState
}

// textState represents the initial state, we can assume that it will be text
// but this will flip us into a variable if we need
func textState(l *lexer) stateFunc {
	for {
		if strings.HasPrefix(l.content[l.pos:], leftMeta) {
			if l.pos > l.start {
				l.emit(tokenText)
			}
			return leftMetaState
		}
		if l.next() == eof {
			break
		}
	}
	if l.pos > l.start {
		l.emit(tokenText)
	}

	l.emit(tokenEOF)
	return nil
}

// Converts to a very simple syntax tree
func (l *lexer) tokenize(content string) error {
	l.content = content
	for state := textState; state != nil; {
		state = state(l)
	}
	return nil
}

// Next will incremenet the position and return the rune at that position if it
// can
func (l *lexer) next() rune {
	l.pos++
	if l.pos >= len(l.content) {
		return eof
	}
	return rune(l.content[l.pos])
}

// Emit will add the element to the AST and set the new start position
func (l *lexer) emit(t token) {
	l.ast = append(l.ast, element{
		token:   t,
		content: l.content[l.start:l.pos],
	})
	l.start = l.pos
}
