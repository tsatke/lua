package parser

import (
	"github.com/tsatke/lua/internal/ast"
	"github.com/tsatke/lua/internal/token"
)

func (suite *ParserSuite) TestParse() {
	suite.assertBlockString(`
print("Hello, World!")
`, ast.Chunk{
		Name: "<unknown input>",
		Block: ast.Block{
			ast.FunctionCall{
				PrefixExp: ast.PrefixExp{
					Name: token.New("print", token.Position{2, 1, 1}, token.Name),
					Fragments: []ast.PrefixExpFragment{
						{
							Args: &ast.Args{
								ExpList: []ast.Exp{
									ast.SimpleExp{
										String: token.New("Hello, World!", token.Position{2, 7, 7}, token.String),
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

func (suite *ParserSuite) TestFunctionCallColon() {
	suite.assertBlockString(`
io.stderr:write("foobar")
`, ast.Chunk{
		Name: "<unknown input>",
		Block: ast.Block{
			ast.FunctionCall{
				PrefixExp: ast.PrefixExp{
					Name: token.New("io", token.Position{2, 1, 1}, token.Name),
					Fragments: []ast.PrefixExpFragment{
						{
							Name: token.New("stderr", token.Position{2, 4, 4}, token.Name),
						},
						{
							Name: token.New("write", token.Position{2, 11, 11}, token.Name),
							Args: &ast.Args{
								ExpList: []ast.Exp{
									ast.SimpleExp{
										String: token.New("foobar", token.Position{2, 17, 17}, token.String),
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

func (suite *ParserSuite) TestAssignment() {
	suite.assertBlockString(`
a=x
`, ast.Chunk{
		Name: "<unknown input>",
		Block: ast.Block{
			ast.Assignment{
				VarList: []ast.Var{
					{
						PrefixExp: ast.PrefixExp{
							Name: token.New("a", token.Position{2, 1, 1}, token.Name),
						},
					},
				},
				ExpList: []ast.Exp{
					ast.PrefixExp{
						Name: token.New("x", token.Position{2, 3, 3}, token.Name),
					},
				},
			},
		},
	})
}

func (suite *ParserSuite) TestFunctionDeclaration() {
	suite.assertBlockString(`
function foo.bar()
	some = code
end
`, ast.Chunk{
		Name: "<unknown input>",
		Block: ast.Block{
			ast.Function{
				FuncName: &ast.FuncName{
					Name1: []token.Token{
						token.New("foo", token.Position{2, 10, 10}, token.Name),
						token.New("bar", token.Position{2, 14, 14}, token.Name),
					},
				},
				FuncBody: ast.FuncBody{
					Block: ast.Block{
						ast.Assignment{
							VarList: []ast.Var{
								{
									PrefixExp: ast.PrefixExp{
										Name: token.New("some", token.Position{3, 2, 21}, token.Name),
									},
								},
							},
							ExpList: []ast.Exp{
								ast.PrefixExp{
									Name: token.New("code", token.Position{3, 9, 28}, token.Name),
								},
							},
						},
					},
				},
			},
		},
	})
}

func (suite *ParserSuite) TestNestedFunctionCall() {
	suite.assertBlockString(`
print(pcall(print, "print message"))
`, ast.Chunk{
		Name: "<unknown input>",
		Block: ast.Block{
			ast.FunctionCall{
				PrefixExp: ast.PrefixExp{
					Name: token.New("print", token.Position{2, 1, 1}, token.Name),
					Fragments: []ast.PrefixExpFragment{
						{
							Args: &ast.Args{
								ExpList: []ast.Exp{
									ast.PrefixExp{
										Name: token.New("pcall", token.Position{2, 7, 7}, token.Name),
										Fragments: []ast.PrefixExpFragment{
											{
												Args: &ast.Args{
													ExpList: []ast.Exp{
														ast.PrefixExp{
															Name: token.New("print", token.Position{2, 13, 13}, token.Name),
														},
														ast.SimpleExp{
															String: token.New("print message", token.Position{2, 20, 20}, token.String),
														},
													},
												},
											},
										},
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
