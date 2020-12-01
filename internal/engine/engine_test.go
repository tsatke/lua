package engine

func (suite *EngineSuite) TestTrivial() {
	err := suite.engine.EvalString(`
a = "Hello, World!"
print(a)
`)
	suite.NoError(err)
	suite.Equal("Hello, World!\n", suite.stdout.String())
	suite.engine.dumpState()
}
