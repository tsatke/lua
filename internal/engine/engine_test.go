package engine

import "strings"

func (suite *EngineSuite) TestTrivial() {
	results, err := suite.engine.Eval(strings.NewReader(`
a = "Hello, World!"
print(a)
`))
	suite.Len(results, 0)
	suite.NoError(err)
	suite.Equal("Hello, World!\n", suite.stdout.String())
}
