package parser

import (
	"fmt"
	"io"
	"io/ioutil"
	"unicode"

	"github.com/tsatke/lua/internal/token"
)

type inMemoryScanner struct {
	input []rune

	state
}

func newInMemoryScanner(source io.Reader) (*inMemoryScanner, error) {
	data, err := ioutil.ReadAll(source)
	if err != nil {
		return nil, fmt.Errorf("read all: %w", err)
	}

	runes := []rune(string(data))
	return &inMemoryScanner{
		input: runes,
		state: state{
			startLine: 1,
			startCol:  1,
			line:      1,
			col:       1,
		},
	}, nil
}

func (s *inMemoryScanner) next() (token.Token, bool) {
	return s.computeNext()
}

func (s *inMemoryScanner) updateStartPositions() {
	s.start = s.pos
	s.startLine = s.line
	s.startCol = s.col
}

func (s *inMemoryScanner) token(typ ...token.Type) token.Token {
	tok := token.New(s.candidate(), token.Position{
		Line:   s.startLine,
		Col:    s.startCol,
		Offset: int64(s.start),
	}, typ...)
	s.updateStartPositions()
	return tok
}

func (s *inMemoryScanner) error(err error) token.Token {
	tok := token.New(err.Error(), token.Position{
		Line:   s.startLine,
		Col:    s.startCol,
		Offset: int64(s.start),
	}, token.Error)
	s.updateStartPositions()
	return tok
}

func (s *inMemoryScanner) candidate() string {
	return string(s.input[s.start:s.pos])
}

func (s *inMemoryScanner) done() bool {
	return s.pos >= len(s.input)
}

func (s *inMemoryScanner) lookahead() (rune, bool) {
	if !s.done() {
		return s.input[s.pos], true
	}
	return 0, false
}

func (s *inMemoryScanner) consume() {
	if s.input[s.pos] == '\n' {
		s.line++
		s.col = 1
	} else {
		s.col++
	}
	s.pos++
}

func (s *inMemoryScanner) consumeN(n int) {
	for i := 0; i < n; i++ {
		s.consume()
	}
}

func (s *inMemoryScanner) check(ahead string) bool {
	runes := []rune(ahead)
	for i, r := range runes {
		if r != s.input[s.pos+i] {
			return false
		}
	}
	s.consumeN(len(runes))
	return true
}

func (s *inMemoryScanner) tkpos() token.Position {
	return token.Position{
		Line:   s.startLine,
		Col:    s.startCol,
		Offset: int64(s.start),
	}
}

func (s *inMemoryScanner) drainWhitespace() {
	for {
		r, ok := s.lookahead()
		if !(ok && unicode.IsSpace(r)) {
			break
		}
		s.consume()
	}
	_ = s.token() // ignore whitespaces
}

func (s *inMemoryScanner) computeNext() (token.Token, bool) {
	s.drainWhitespace()
	r, ok := s.lookahead()
	if !ok {
		return nil, false
	}
	switch r {
	case 'a':
		if s.check("and") {
			return s.token(token.And, token.BinaryOperator), true
		}
	case 'b':
		if s.check("break") {
			return s.token(token.Break), true
		}
	case 'd':
		if s.check("do") {
			return s.token(token.Do), true
		}
	case 'e':
		if s.check("elseif") {
			return s.token(token.Elseif), true
		} else if s.check("else") {
			return s.token(token.Else), true
		} else if s.check("end") {
			return s.token(token.End), true
		}
	case 'f':
		if s.check("false") {
			return s.token(token.False), true
		} else if s.check("for") {
			return s.token(token.For), true
		} else if s.check("function") {
			return s.token(token.Function), true
		}
	case 'i':
		if s.check("if") {
			return s.token(token.If), true
		} else if s.check("in") {
			return s.token(token.In), true
		}
	case 'l':
		if s.check("local") {
			return s.token(token.Local), true
		}
	case 'n':
		if s.check("nil") {
			return s.token(token.Nil), true
		} else if s.check("not") {
			return s.token(token.Not, token.UnaryOperator), true
		}
	case 'o':
		if s.check("or") {
			return s.token(token.Or, token.BinaryOperator), true
		}
	case 'r':
		if s.check("repeat") {
			return s.token(token.Repeat), true
		} else if s.check("return") {
			return s.token(token.Return), true
		}
	case 't':
		if s.check("then") {
			return s.token(token.Then), true
		} else if s.check("true") {
			return s.token(token.True), true
		}
	case 'u':
		if s.check("until") {
			return s.token(token.Until), true
		}
	case 'w':
		if s.check("while") {
			return s.token(token.While), true
		}
	case '(':
		if s.check("(") {
			return s.token(token.ParLeft), true
		}
	case ')':
		if s.check(")") {
			return s.token(token.ParRight), true
		}
	case '[':
		if s.check("[") {
			return s.token(token.BracketLeft), true
		}
	case ']':
		if s.check("]") {
			return s.token(token.BracketRight), true
		}
	case '{':
		if s.check("{") {
			return s.token(token.CurlyLeft), true
		}
	case '}':
		if s.check("}") {
			return s.token(token.CurlyRight), true
		}
	case '.':
		if s.check("...") {
			return s.token(token.Ellipsis), true
		} else if s.check("..") {
			return s.token(token.BinaryOperator), true
		} else if s.check(".") {
			return s.token(token.Dot), true
		}
	case '+':
		if s.check("+") {
			return s.token(token.BinaryOperator), true
		}
	case '-':
		if s.check("-") {
			return s.token(token.UnaryOperator, token.BinaryOperator), true
		}
	case '*':
		if s.check("*") {
			return s.token(token.BinaryOperator), true
		}
	case '/':
		if s.check("/") {
			return s.token(token.BinaryOperator), true
		}
	case '^':
		if s.check("^") {
			return s.token(token.BinaryOperator), true
		}
	case '%':
		if s.check("%") {
			return s.token(token.BinaryOperator), true
		}
	case '<':
		if s.check("<=") {
			return s.token(token.BinaryOperator), true
		} else if s.check("<") {
			return s.token(token.BinaryOperator), true
		}
	case '>':
		if s.check(">=") {
			return s.token(token.BinaryOperator), true
		} else if s.check(">") {
			return s.token(token.BinaryOperator), true
		}
	case '=':
		if s.check("==") {
			return s.token(token.BinaryOperator), true
		} else if s.check("=") {
			return s.token(token.Assign), true
		}
	case '~':
		if s.check("~=") {
			return s.token(token.BinaryOperator), true
		}
	case '#':
		if s.check("#") {
			return s.token(token.UnaryOperator), true
		}
	case '"', '\'':
		return s.string_()
	}
	// if none of these optimized lookaheads match, try this next
	switch {
	case unicode.IsLetter(s.input[s.pos]) || s.input[s.pos] == '_':
		return s.ident()
	}
	return s.error(fmt.Errorf("unexpected rune %s", string(s.input[s.pos]))), false
}

func (s *inMemoryScanner) string_() (token.Token, bool) {
	var delimiter rune
	if s.check("\"") {
		delimiter = '"'
	} else if s.check("'") {
		delimiter = '\''
	} else {
		return s.error(fmt.Errorf("string can not start with '%s'", string(s.input[s.pos]))), false
	}

	var complete bool
	for next, ok := s.lookahead(); ok; next, ok = s.lookahead() {
		s.consume()
		if next == delimiter {
			complete = true
			break
		}
	}
	if !complete {
		return s.error(fmt.Errorf("incomplete string %s<EOF>", s.candidate())), false
	}
	return s.token(token.String), true
}

func (s *inMemoryScanner) ident() (token.Token, bool) {
	first, ok := s.lookahead()
	if !ok {
		return nil, false
	}
	if !(unicode.IsLetter(first) || first == '_') {
		return s.error(fmt.Errorf("expected letter or underscore, but got %s", string(first))), false
	}
	s.consume()
	for {
		next, ok := s.lookahead()
		if !ok {
			break
		}
		if !(unicode.IsLetter(next) || unicode.IsDigit(next) || next == '_') {
			break
		}
		s.consume()
	}
	return s.token(token.Name), true
}
