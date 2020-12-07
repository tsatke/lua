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

func (suite *ParserSuite) assertBlockString(input string, expected ast.Chunk) {
	suite.assertBlock(strings.NewReader(input), expected)
}

func (suite *ParserSuite) assertBlock(source io.Reader, expected ast.Chunk) {
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
			return suite.Equal(left.Types(), right.Types()) &&
				suite.Equal(left.Types(), right.Types()) &&
				suite.EqualValues(left.Types(), right.Types())
		}),
	}

	if !cmp.Equal(expected, got, opts...) {
		suite.Failf("not equal", "%s", cmp.Diff(expected, got, opts...))
	}
}
