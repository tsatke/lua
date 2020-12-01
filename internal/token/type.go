package token

//go:generate stringer -type=Type

// Type is a token type.
type Type uint8

// Known types.
const (
	TypeUnknown Type = iota

	// And is the token type for the keyword 'and'.
	And
	// Break is the token type for the keyword 'break'.
	Break
	// Do is the token type for the keyword 'do'.
	Do
	// Else is the token type for the keyword 'else'.
	Else
	// Elseif is the token type for the keyword 'elseif'.
	Elseif
	// End is the token type for the keyword 'end'.
	End
	// False is the token type for the keyword 'false'.
	False
	// For is the token type for the keyword 'for'.
	For
	// Function is the token type for the keyword 'function'.
	Function
	// If is the token type for the keyword 'if'.
	If
	// In is the token type for the keyword 'in'.
	In
	// Local is the token type for the keyword 'local'.
	Local
	// Nil is the token type for the keyword 'nil'.
	Nil
	// Not is the token type for the keyword 'not'.
	Not
	// Or is the token type for the keyword 'or'.
	Or
	// Repeat is the token type for the keyword 'repeat'.
	Repeat
	// Return is the token type for the keyword 'return'.
	Return
	// Then is the token type for the keyword 'then'.
	Then
	// True is the token type for the keyword 'true'.
	True
	// Until is the token type for the keyword 'until'.
	Until
	// While is the token type for the keyword 'while'.
	While

	// BinaryOperator is the token type for any binary operator.
	// Please note, that unary operators that can also be binary operators,
	// such as '-', will also have this type (in addition to the unary operator type).
	BinaryOperator
	// UnaryOperator is the token type for any unary operator.
	// Please note, that binary operators that can also be unary operators,
	// such as '-', will also have this type (in addition to the binary operator type).
	UnaryOperator
	// Assign is the token type for the operator '='.
	Assign

	// Number is the token type for any number.
	Number
	// String is the token type for any string.
	String
	// Name is the token type for any identifier.
	Name

	// ParLeft is the token type for an opening parenthesis '('.
	ParLeft
	// ParRight is the token type for a closing parenthesis ')'.
	ParRight
	// CurlyLeft is the token type for an opening curly bracket '{'.
	CurlyLeft
	// CurlyRight is the token type for a closing curly bracket '}'.
	CurlyRight
	// BracketLeft is the token type for an opening bracket '['.
	BracketLeft
	// BracketRight is the token type for a closing bracket ']'.
	BracketRight

	// SemiColon is the token type for a semicolon ';'.
	SemiColon
	// Colon is the token type for a colon ':'.
	Colon
	// Comma is the token type for a comma ','.
	Comma
	// Dot is the token type for a period '.'.
	Dot
	// DoubleDot is the token type for a double dot '..'.
	DoubleDot
	// Ellipsis is the token type for an ellipsis (or triple dot) '...'.
	Ellipsis

	// Error is a token type that indicates that this token represents
	// a corrupt or incorrect structure. The value of this token is the error message.
	Error
)
