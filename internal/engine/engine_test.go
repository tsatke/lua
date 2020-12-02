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

func (suite *EngineSuite) TestExample01() {
	file, err := suite.testdata.Open("example01.lua")
	suite.NoError(err)

	results, err := suite.engine.Eval(file)
	suite.EqualError(err, "expected error message")
	suite.Len(results, 0)
	suite.Equal("line 1 on stdout\nline 2 on stdout\nline 3 on stdout\n", suite.stdout.String())
	suite.Equal("", suite.stderr.String())
}

func (suite *EngineSuite) TestError01() {
	file, err := suite.testdata.Open("errors/error01.lua")
	suite.NoError(err)

	results, err := suite.engine.Eval(file)
	suite.EqualError(err, "error called with <nil>") // default error message, if no args are given
	suite.Len(results, 0)
	suite.Equal("", suite.stdout.String())
	suite.Equal("", suite.stderr.String())
}

func (suite *EngineSuite) TestError02() {
	file, err := suite.testdata.Open("errors/error02.lua")
	suite.NoError(err)

	results, err := suite.engine.Eval(file)
	suite.EqualError(err, "custom message")
	suite.Len(results, 0)
	suite.Equal("", suite.stdout.String())
	suite.Equal("", suite.stderr.String())
}
