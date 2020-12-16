package parser

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/suite"
	"github.com/tsatke/lua/internal/ast"
	"github.com/tsatke/lua/internal/token"
)

func TestParserSuite(t *testing.T) {
	suite.Run(t, new(ParserSuite))
}

type ParserSuite struct {
	suite.Suite
}

func (suite *ParserSuite) assertChunkString(input string, expected ast.Chunk) {
	suite.assertChunk(strings.NewReader(input), expected)
}

func (suite *ParserSuite) assertChunk(source io.Reader, expected ast.Chunk) {
	parser, err := New(source)
	suite.NoError(err)

	got, ok := parser.Parse()
	suite.True(ok, "Parse failed")

	if len(parser.Errors()) > 0 {
		var errors bytes.Buffer
		for _, err := range parser.Errors() {
			errors.WriteString("\t" + err.Error() + "\n")
		}
		suite.Failf("there are parse errors", "%s", errors.String())
	}

	opts := []cmp.Option{
		cmp.Comparer(func(left, right token.Token) bool {
			if left == nil || right == nil {
				return suite.Equal(left, right)
			}
			if left.Value() != right.Value() {
				return false
			}
			if left.Pos() != right.Pos() {
				return false
			}
			if !suite.EqualValuesf(left.Types(), right.Types(), "types don't match: expected %v but got %v", left.Types(), right.Types()) {
				return false
			}
			return true
		}),
	}

	if !cmp.Equal(expected, got, opts...) {
		suite.Failf("not equal", "%s", cmp.Diff(expected, got, opts...))
	}
}
