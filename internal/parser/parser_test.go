package parser

import (
	"github.com/tsatke/lua/internal/ast"
	"github.com/tsatke/lua/internal/token"
)

func (suite *ParserSuite) TestParse() {
	suite.assertBlockString(`
print("Hello, World!")
`, ast.Chunk{
		ast.FunctionCall{
			PrefixExp: ast.PrefixExp{
				Var: ast.Var{
					Name: token.New("print", token.Position{2, 1, 1}, token.Name),
				},
			},
			Args: ast.Args{
				ExpList: []ast.Exp{
					ast.SimpleExp{
						String: token.New(`"Hello, World!"`, token.Position{2, 7, 7}, token.String),
					},
				},
			},
		},
	})
}

func (suite *ParserSuite) TestAssignment() {
	suite.assertBlockString(`
a=x
`, ast.Chunk{
		ast.Assignment{
			VarList: []ast.Var{
				{Name: token.New("a", token.Position{2, 1, 1}, token.Name)},
			},
			ExpList: []ast.Exp{
				ast.PrefixExp{
					Var: ast.Var{Name: token.New("x", token.Position{2, 3, 3}, token.Name)},
				},
			},
		},
	})
}

func (suite *ParserSuite) TestNestedFunctionCall() {
	suite.assertBlockString(`
print(pcall(print, "print message"))
`, ast.Chunk{
		ast.FunctionCall{
			PrefixExp: ast.PrefixExp{
				Var: ast.Var{
					Name: token.New("print", token.Position{2, 1, 1}, token.Name),
				},
			},
			Args: ast.Args{
				ExpList: []ast.Exp{
					ast.PrefixExp{
						FunctionCall: ast.FunctionCall{
							PrefixExp: ast.PrefixExp{
								Var: ast.Var{
									Name: token.New("pcall", token.Position{2, 7, 7}, token.Name),
								},
							},
							Args: ast.Args{
								ExpList: []ast.Exp{
									ast.PrefixExp{
										Var: ast.Var{
											Name: token.New("print", token.Position{2, 13, 13}, token.Name),
										},
									},
									ast.SimpleExp{
										String: token.New(`"print message"`, token.Position{2, 20, 20}, token.String),
									},
								},
							},
						},
					},
				},
			},
		},
	})
}
