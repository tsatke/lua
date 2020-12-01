package parser

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/tsatke/lua/internal/ast"
)

func TestParserSuite(t *testing.T) {
	suite.Run(t, new(ParserSuite))
}

type ParserSuite struct {
	suite.Suite
}

func (suite *ParserSuite) assertBlockString(input string, expected ast.Block) {
	suite.assertBlock(strings.NewReader(input), expected)
}

func (suite *ParserSuite) assertBlock(source io.Reader, expected ast.Block) {
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

	suite.EqualValues(expected, got)
}
