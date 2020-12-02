package parser

type state struct {
	start     int
	startLine int
	startCol  int

	pos  int
	line int
	col  int
}
