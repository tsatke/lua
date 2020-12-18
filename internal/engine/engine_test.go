package engine

import (
	"bytes"
	"github.com/spf13/afero"
	"github.com/tsatke/lua/internal/engine/value"
	"path/filepath"
	"strings"
	"time"
)

func (suite *EngineSuite) TestTrivial() {
	results, err := suite.engine.Eval(strings.NewReader(`
a = "Hello, World!"
print(a)
`))
	suite.Len(results, 0)
	suite.NoError(err)
	suite.Equal("Hello, World!\n", suite.stdout.String())
}

func (suite *EngineSuite) TestStack() {
	results, err := suite.engine.Eval(strings.NewReader(`
function a()
	b()
end

function b()
	c()
end

function c()
	error("Message")
end

a()
`))
	suite.Len(results, 0)
	suite.IsType(Error{}, err)
	suite.Equal("Message", err.(Error).Message.(value.String).String())
	suite.Equal([]StackFrame{
		{
			Name: "error",
		},
		{
			Name: "c",
		},
		{
			Name: "b",
		},
		{
			Name: "a",
		},
		{
			Name: "<unknown input>",
		},
	}, err.(Error).Stack)
}

func (suite *EngineSuite) TestStackOverflow() {
	maxStackSize := 5000
	start := time.Now()

	e := New(WithMaxStackSize(maxStackSize))
	results, err := e.Eval(strings.NewReader(`
function infiniteRecursion()
	infiniteRecursion()
end

infiniteRecursion()
`))

	suite.T().Logf("Stack overflow took %s to occur", time.Since(start))

	suite.Len(results, 0)
	suite.EqualError(err, "Stack overflow while calling 'infiniteRecursion'")
}

func (suite *EngineSuite) TestLuaSuite() {
	basePath := "suite"
	mainFile := "main.lua"

	stdin := new(bytes.Buffer)
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	engine := New(
		WithStdin(stdin),
		WithStdout(stdout),
		WithStderr(stderr),
		WithClock(mockClock{}),
		WithFs(afero.NewBasePathFs(suite.testdata, basePath)),
	)

	file, err := suite.testdata.Open(filepath.Join(basePath, mainFile))
	suite.Require().NoError(err)
	defer func() { _ = file.Close() }()

	results, err := engine.Eval(file)
	if err != nil {
		if luaErr, ok := err.(Error); ok {
			suite.T().Logf("lua suite called error():\n%s", luaErr.String())
		} else {
			suite.T().Logf("execution failed\n%s", err)
		}
		suite.Fail("lua tests failed")
	}

	suite.T().Logf("stdout (%d bytes):\n%s", len(stdout.Bytes()), stdout.String())
	suite.T().Logf("stderr (%d bytes):\n%s", len(stderr.Bytes()), stderr.String())

	suite.Require().Len(results, 1)
	rc := results[0]
	suite.EqualValues(rc.(value.Number).Value(), 0, "RC != 0")
}
