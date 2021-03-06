// Code generated by "stringer -type=Type"; DO NOT EDIT.

package token

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[TypeUnknown-0]
	_ = x[And-1]
	_ = x[Break-2]
	_ = x[Do-3]
	_ = x[Else-4]
	_ = x[Elseif-5]
	_ = x[End-6]
	_ = x[False-7]
	_ = x[For-8]
	_ = x[Function-9]
	_ = x[If-10]
	_ = x[In-11]
	_ = x[Local-12]
	_ = x[Nil-13]
	_ = x[Not-14]
	_ = x[Or-15]
	_ = x[Repeat-16]
	_ = x[Return-17]
	_ = x[Then-18]
	_ = x[True-19]
	_ = x[Until-20]
	_ = x[While-21]
	_ = x[BinaryOperator-22]
	_ = x[UnaryOperator-23]
	_ = x[Assign-24]
	_ = x[Number-25]
	_ = x[String-26]
	_ = x[Name-27]
	_ = x[ParLeft-28]
	_ = x[ParRight-29]
	_ = x[CurlyLeft-30]
	_ = x[CurlyRight-31]
	_ = x[BracketLeft-32]
	_ = x[BracketRight-33]
	_ = x[SemiColon-34]
	_ = x[Colon-35]
	_ = x[Comma-36]
	_ = x[Dot-37]
	_ = x[DoubleDot-38]
	_ = x[Ellipsis-39]
	_ = x[Error-40]
}

const _Type_name = "TypeUnknownAndBreakDoElseElseifEndFalseForFunctionIfInLocalNilNotOrRepeatReturnThenTrueUntilWhileBinaryOperatorUnaryOperatorAssignNumberStringNameParLeftParRightCurlyLeftCurlyRightBracketLeftBracketRightSemiColonColonCommaDotDoubleDotEllipsisError"

var _Type_index = [...]uint8{0, 11, 14, 19, 21, 25, 31, 34, 39, 42, 50, 52, 54, 59, 62, 65, 67, 73, 79, 83, 87, 92, 97, 111, 124, 130, 136, 142, 146, 153, 161, 170, 180, 191, 203, 212, 217, 222, 225, 234, 242, 247}

func (i Type) String() string {
	if i >= Type(len(_Type_index)-1) {
		return "Type(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _Type_name[_Type_index[i]:_Type_index[i+1]]
}
