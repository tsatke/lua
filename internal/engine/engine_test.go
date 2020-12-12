package engine

import (
	"github.com/tsatke/lua/internal/engine/value"
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
