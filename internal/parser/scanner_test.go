package parser

import (
	"github.com/tsatke/lua/internal/token"
)

func (suite *ScannerSuite) TestKeywordTypes() {
	suite.assertTokensString("and break do else elseif end false for function if in local nil not or repeat return then true until while",
		[]token.Token{
			token.New("and", token.Position{1, 1, 0}, token.And, token.BinaryOperator),
			token.New("break", token.Position{1, 5, 4}, token.Break),
			token.New("do", token.Position{1, 11, 10}, token.Do),
			token.New("else", token.Position{1, 14, 13}, token.Else),
			token.New("elseif", token.Position{1, 19, 18}, token.Elseif),
			token.New("end", token.Position{1, 26, 25}, token.End),
			token.New("false", token.Position{1, 30, 29}, token.False),
			token.New("for", token.Position{1, 36, 35}, token.For),
			token.New("function", token.Position{1, 40, 39}, token.Function),
			token.New("if", token.Position{1, 49, 48}, token.If),
			token.New("in", token.Position{1, 52, 51}, token.In),
			token.New("local", token.Position{1, 55, 54}, token.Local),
			token.New("nil", token.Position{1, 61, 60}, token.Nil),
			token.New("not", token.Position{1, 65, 64}, token.Not, token.UnaryOperator),
			token.New("or", token.Position{1, 69, 68}, token.Or, token.BinaryOperator),
			token.New("repeat", token.Position{1, 72, 71}, token.Repeat),
			token.New("return", token.Position{1, 79, 78}, token.Return),
			token.New("then", token.Position{1, 86, 85}, token.Then),
			token.New("true", token.Position{1, 91, 90}, token.True),
			token.New("until", token.Position{1, 96, 95}, token.Until),
			token.New("while", token.Position{1, 102, 101}, token.While),
		})
}

func (suite *ScannerSuite) TestOperatorTypes() {
	suite.assertTokensString("+ - * / ^ % .. < <= > >= == ~= #",
		[]token.Token{
			token.New("+", token.Position{1, 1, 0}, token.BinaryOperator),
			token.New("-", token.Position{1, 3, 2}, token.UnaryOperator, token.BinaryOperator),
			token.New("*", token.Position{1, 5, 4}, token.BinaryOperator),
			token.New("/", token.Position{1, 7, 6}, token.BinaryOperator),
			token.New("^", token.Position{1, 9, 8}, token.BinaryOperator),
			token.New("%", token.Position{1, 11, 10}, token.BinaryOperator),
			token.New("..", token.Position{1, 13, 12}, token.BinaryOperator),
			token.New("<", token.Position{1, 16, 15}, token.BinaryOperator),
			token.New("<=", token.Position{1, 18, 17}, token.BinaryOperator),
			token.New(">", token.Position{1, 21, 20}, token.BinaryOperator),
			token.New(">=", token.Position{1, 23, 22}, token.BinaryOperator),
			token.New("==", token.Position{1, 26, 25}, token.BinaryOperator),
			token.New("~=", token.Position{1, 29, 28}, token.BinaryOperator),
			token.New("#", token.Position{1, 32, 31}, token.UnaryOperator),
		})
}

func (suite *ScannerSuite) TestLinefeed() {
	suite.assertTokensString(`
break
 break
		break

do`,
		[]token.Token{
			token.New("break", token.Position{2, 1, 1}, token.Break),
			token.New("break", token.Position{3, 2, 8}, token.Break),
			token.New("break", token.Position{4, 3, 16}, token.Break),
			token.New("do", token.Position{6, 1, 23}, token.Do),
		})
}