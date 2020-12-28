package parser

/*
Operator precedence as specified in the language reference, from lower to higher.

or
and
<     >     <=    >=    ~=    ==
|
~
&
<<    >>
..
+     -
*     /     //    %
unary operators (not   #     -     ~)
^

*/

type precedence uint8

const (
	precedence0 precedence = iota
	precedence1
	precedence2
	precedence3
	precedence4
	precedence5
	precedence6
	precedence7
	precedence8
	precedence9
	precedence10
	precedence11
)

var (
	precedences = map[string]precedence{
		"or":  precedence0,
		"and": precedence1,
		"<":   precedence2,
		">":   precedence2,
		"<=":  precedence2,
		">=":  precedence2,
		"~=":  precedence2,
		"==":  precedence2,
		"|":   precedence3,
		"~":   precedence4,
		"&":   precedence5,
		"<<":  precedence6,
		">>":  precedence6,
		"..":  precedence7,
		"+":   precedence8,
		"-":   precedence8,
		"*":   precedence9,
		"/":   precedence9,
		"//":  precedence9,
		"%":   precedence9,
		//"not": precedence10, // unary
		//"#":   precedence10, // unary
		//"-":   precedence10, // unary
		//"~":   precedence10, // unary
		"^": precedence11,
	}
)

func precedenceOf(operator string) precedence {
	return precedences[operator]
}

func isRightAssociative(operator string) bool {
	return operator == ".." || operator == "^"
}
