package parser

import "testing"

func Test_unescape(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			"empty",
			"",
			"",
			false,
		},
		{
			"trivial",
			"abc ",
			"abc ",
			false,
		},
		{
			"bell escape",
			`\a`,
			"\a",
			false,
		},
		{
			"backspace escape",
			`\a`,
			"\a",
			false,
		},
		{
			"form feed escape",
			`\b`,
			"\b",
			false,
		},
		{
			"newline escape",
			`\f`,
			"\f",
			false,
		},
		{
			"carriage return escape",
			`\n`,
			"\n",
			false,
		},
		{
			"horizontal tab escape",
			`\r`,
			"\r",
			false,
		},
		{
			"vertical tab escape",
			`\t`,
			"\t",
			false,
		},
		{
			"backslash escape",
			`\v`,
			"\v",
			false,
		},
		{
			"double quote escape",
			`\\`,
			"\\",
			false,
		},
		{
			"single quote escape",
			`\"`,
			"\"",
			false,
		},
		{
			"newline escape",
			`\
`,
			"\n",
			false,
		},
		{
			"inside text",
			`\athis is a\
long text, that uses some weii\brd\
mechanics`,
			"\athis is a\nlong text, that uses some weii\brd\nmechanics",
			false,
		},
		{
			"decimal escape",
			`\123`,
			"{",
			false,
		},
		{
			"decimal escape",
			`\255`,
			"\xff",
			false,
		},
		{
			"decimal escape",
			`\256`,
			"",
			true,
		},
		{
			"decimal escape text",
			`\1\2\3`,
			"\x01\x02\x03",
			false,
		},
		{
			"unfinished escape",
			`\`,
			"",
			true,
		},
		{
			"unfinished hex escape",
			`\x`,
			"",
			true,
		},
		{
			"unfinished hex escape",
			`\xf`,
			"",
			true,
		},
		{
			"possibly incomplete dec escape",
			`\123\123\123\12`,
			"{{{\x0c",
			false,
		},
		{
			"skip whitespace escape",
			`abc\z  	  foobar`,
			"abcfoobar",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := unescape(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("unescape() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("unescape() got = %q, want %q", got, tt.want)
			}
		})
	}
}
