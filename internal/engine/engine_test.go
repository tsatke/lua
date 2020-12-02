package engine

func (suite *EngineSuite) TestTrivial() {
	results, err := suite.engine.EvalString(`
a = "Hello, World!"
print(a)
`)
	suite.Len(results, 0)
	suite.NoError(err)
	suite.Equal("Hello, World!\n", suite.stdout.String())
}
