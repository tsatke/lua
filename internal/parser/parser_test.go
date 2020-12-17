package parser

import (
	"github.com/tsatke/lua/internal/ast"
	"github.com/tsatke/lua/internal/token"
)

func (suite *ParserSuite) TestParse() {
	suite.assertChunkString(`
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
	suite.assertChunkString(`
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
	suite.assertChunkString(`
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
	suite.assertChunkString(`
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
	suite.assertChunkString(`
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

func (suite *ParserSuite) TestReturn() {
	suite.assertChunkString(`
function foo()
	local a = 5
	return a
end
`, ast.Chunk{
		Name: "<unknown input>",
		Block: ast.Block{
			ast.Function{
				FuncName: &ast.FuncName{
					Name1: []token.Token{
						token.New("foo", token.Position{2, 10, 10}, token.Name),
					},
				},
				FuncBody: ast.FuncBody{
					Block: ast.Block{
						ast.Local{
							NameList: []token.Token{
								token.New("a", token.Position{3, 8, 23}, token.Name),
							},
							ExpList: []ast.Exp{
								ast.SimpleExp{
									Number: token.New("5", token.Position{3, 12, 27}, token.Number),
								},
							},
						},
						ast.LastStatement{
							ExpList: []ast.Exp{
								ast.PrefixExp{
									Name: token.New("a", token.Position{4, 9, 37}, token.Name),
								},
							},
						},
					},
				},
			},
		},
	})

	suite.assertChunkString(`
return 5
`, ast.Chunk{
		Name: "<unknown input>",
		Block: ast.Block{
			ast.LastStatement{
				ExpList: []ast.Exp{
					ast.SimpleExp{
						Number: token.New("5", token.Position{2, 8, 8}, token.Number),
					},
				},
			},
		},
	})
}

func (suite *ParserSuite) TestBinaryExpression() {
	suite.assertChunkString(`
return x * 5
`, ast.Chunk{
		Name: "<unknown input>",
		Block: ast.Block{
			ast.LastStatement{
				ExpList: []ast.Exp{
					ast.BinopExp{
						Left: ast.PrefixExp{
							Name: token.New("x", token.Position{2, 8, 8}, token.Name),
						},
						Binop: token.New("*", token.Position{2, 10, 10}, token.BinaryOperator),
						Right: ast.SimpleExp{
							Number: token.New("5", token.Position{2, 12, 12}, token.Number),
						},
					},
				},
			},
		},
	})
}
