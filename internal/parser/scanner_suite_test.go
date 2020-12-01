package parser

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/tsatke/lua/internal/token"
)

func TestScannerSuite(t *testing.T) {
	t.Run("type=inmemory", func(t *testing.T) {
		suite.Run(t, &ScannerSuite{
			scannerGenerator: func(rd io.Reader) (scanner, error) {
				return newInMemoryScanner(rd)
			},
		})
	})
}

type ScannerSuite struct {
	suite.Suite

	scannerGenerator func(io.Reader) (scanner, error)
}

func (suite *ScannerSuite) assertTokensString(input string, tokens []token.Token) {
	suite.assertTokens(strings.NewReader(input), tokens)
}

func (suite *ScannerSuite) assertTokens(source io.Reader, expected []token.Token) {
	sc, err := suite.scannerGenerator(source)
	suite.NoError(err)

	got := make([]token.Token, 0)
	var tk token.Token
	ok := true
	for ok {
		tk, ok = sc.next()
		if ok {
			got = append(got, tk)
		}
	}

	for _, next := range got {
		suite.T().Logf("%s", next)
	}

	suite.Equalf(len(expected), len(got), "did not receive as much got as expected (expected %d, but got %d)", len(expected), len(got))

	limit := len(expected)
	if len(got) < limit {
		limit = len(got)
	}
	suite.NotNil(got)
	for i := 0; i < limit; i++ {
		suite.Equal(expected[i].Pos(), got[i].Pos(), "Position doesn't match")
		suite.EqualValues(expected[i].Types(), got[i].Types(), "Types don't match, expected %v, but got %v", expected[i].Types(), got[i].Types())
		suite.Equal(expected[i].Value(), got[i].Value(), "Value doesn't match")
	}
}
