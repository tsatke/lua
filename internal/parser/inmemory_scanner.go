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

func (s *inMemoryScanner) checkWord(ahead string) bool {
	runes := []rune(ahead)
	for i, r := range runes {
		if r != s.input[s.pos+i] {
			return false
		}
	}

	if len(s.input) > s.pos+len(runes) {
		r := s.input[s.pos+len(runes)]
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
			/*
				Assuming that ahead is e.g. 'and', we can't match a variable name like
				'and_this_is_my_var', or 'andThis', which is, why we check if the word
				is followed by a rune that would be valid for a Lua name.
			*/
			return false
		}
	}
	s.consumeN(len(runes))
	return true
}

func (s *inMemoryScanner) checkNumber() bool {
	i := 0
	hasMore := func() bool {
		return len(s.input) > s.pos+i
	}
	get := func() rune {
		return s.input[s.pos+i]
	}
	consume := func() {
		i++
	}

	var signed bool
	// optional sign
	if hasMore() && get() == '-' {
		signed = true
		consume() // consume the sign
	}

	// optional integral digits
	for hasMore() && unicode.IsDigit(get()) {
		consume()
	}

	// optional fractional part
	if hasMore() && get() == '.' {
		consume()

		if !(hasMore() && unicode.IsDigit(get())) {
			// no digit, require at least one digit after decimal point
			return false
		}

		// optional fractional digits
		for hasMore() && unicode.IsDigit(get()) {
			consume()
		}
	}

	// optional exponent part
	if hasMore() && (get() == 'e' || get() == 'E') {
		consume()

		if !(hasMore() && unicode.IsDigit(get())) {
			// no digit, require at least one digit after exponent indicator
			return false
		}

		// optional exponent digits
		for hasMore() && unicode.IsDigit(get()) {
			consume()
		}
	}
	if i == 0 || (i == 1 && signed) {
		// if we didn't consume any runes or just one rune, but it was the sign,
		// then this is not a number
		return false
	}
	s.consumeN(i)
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

func (s *inMemoryScanner) skipRemainingLine() {
	var done bool
	for !done {
		next, ok := s.lookahead()
		if !ok {
			return
		}
		if next == '\n' {
			done = true
		}
		s.consume()
	}
	_ = s.token() // ignore this line
}

func (s *inMemoryScanner) computeNext() (token.Token, bool) {
start:
	if s.pos == 0 {
		// skip shebang
		if s.check("#!") {
			s.skipRemainingLine()
		}
	}
	s.drainWhitespace()
	r, ok := s.lookahead()
	if !ok {
		return nil, false
	}
	switch r {
	case 'a':
		if s.checkWord("and") {
			return s.token(token.And, token.BinaryOperator), true
		}
	case 'b':
		if s.checkWord("break") {
			return s.token(token.Break), true
		}
	case 'd':
		if s.checkWord("do") {
			return s.token(token.Do), true
		}
	case 'e':
		if s.checkWord("elseif") {
			return s.token(token.Elseif), true
		} else if s.checkWord("else") {
			return s.token(token.Else), true
		} else if s.checkWord("end") {
			return s.token(token.End), true
		}
	case 'f':
		if s.checkWord("false") {
			return s.token(token.False), true
		} else if s.checkWord("for") {
			return s.token(token.For), true
		} else if s.checkWord("function") {
			return s.token(token.Function), true
		}
	case 'i':
		if s.checkWord("if") {
			return s.token(token.If), true
		} else if s.checkWord("in") {
			return s.token(token.In), true
		}
	case 'l':
		if s.checkWord("local") {
			return s.token(token.Local), true
		}
	case 'n':
		if s.checkWord("nil") {
			return s.token(token.Nil), true
		} else if s.checkWord("not") {
			return s.token(token.Not, token.UnaryOperator), true
		}
	case 'o':
		if s.checkWord("or") {
			return s.token(token.Or, token.BinaryOperator), true
		}
	case 'r':
		if s.checkWord("repeat") {
			return s.token(token.Repeat), true
		} else if s.checkWord("return") {
			return s.token(token.Return), true
		}
	case 't':
		if s.checkWord("then") {
			return s.token(token.Then), true
		} else if s.checkWord("true") {
			return s.token(token.True), true
		}
	case 'u':
		if s.checkWord("until") {
			return s.token(token.Until), true
		}
	case 'w':
		if s.checkWord("while") {
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
		} else if s.checkNumber() {
			return s.token(token.Number), true
		} else if s.check(".") {
			return s.token(token.Dot), true
		}
	case '+':
		if s.checkNumber() {
			return s.token(token.Number), true
		} else if s.check("+") {
			return s.token(token.BinaryOperator), true
		}
	case '-':
		if s.check("--") { // EOL-comment
			s.skipRemainingLine() // ignore everything until line-end
			goto start
		} else if s.checkNumber() {
			return s.token(token.Number), true
		} else if s.check("-") {
			return s.token(token.UnaryOperator, token.BinaryOperator), true
		}
	case '*':
		if s.check("*") {
			return s.token(token.BinaryOperator), true
		}
	case '/':
		if s.check("//") {
			return s.token(token.BinaryOperator), true
		}
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
	case ',':
		if s.check(",") {
			return s.token(token.Comma), true
		}
	case ':':
		if s.check(":") {
			return s.token(token.Colon), true
		}
	case '"', '\'':
		return s.string_()
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		if s.checkNumber() {
			return s.token(token.Number), true
		}
	case ';':
		if s.check(";") {
			return s.token(token.SemiColon), true
		}
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
