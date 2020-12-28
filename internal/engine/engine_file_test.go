package engine

import (
	"bytes"
	"path/filepath"

	"github.com/spf13/afero"
	"github.com/tsatke/lua/internal/engine/value"
)

func (suite *EngineSuite) TestRawget() {
	suite.runFileTests("rawget", []fileTest{
		{
			"rawget01.lua",
			nil,
			"",
			"a\nb\nc\nnil\nnil\n",
			"",
		},
	})
}

func (suite *EngineSuite) TestBreak() {
	suite.runFileTests("break", []fileTest{
		{
			"break01.lua",
			nil,
			"",
			"1\n2\n3\nend\n",
			"",
		},
		{
			"break02.lua",
			nil,
			"",
			"1\t1\n2\t2\n3\t3\nend\n",
			"",
		},
		{
			"break03.lua",
			nil,
			"",
			"1\n2\n3\n4\nend\n",
			"",
		},
		{
			"break04.lua",
			nil,
			"",
			"1\n2\n3\n4\nend\n",
			"",
		},
	})
}

func (suite *EngineSuite) TestIpairs() {
	suite.runFileTests("ipairs", []fileTest{
		{
			"ipairs01.lua",
			nil,
			"",
			"1\ta\n2\tb\n3\tc\n",
			"",
		},
		{
			"ipairs02.lua",
			nil,
			"",
			"1\n2\n3\n",
			"",
		},
	})
}

func (suite *EngineSuite) TestFor() {
	suite.runFileTests("for", []fileTest{
		{
			"for01.lua",
			nil,
			"",
			"0\n1\n2\n3\n4\n5\nend\nnil\n",
			"",
		},
		{
			"for02.lua",
			nil,
			"",
			"1\n3\n5\nend\n",
			"",
		},
		{
			"for03.lua",
			nil,
			"",
			"1\nend\n",
			"",
		},
		{
			"for04.lua",
			nil,
			"",
			"0\n1\n2\n3\n4\n5\nend\n",
			"",
		},
	})
}

func (suite *EngineSuite) TestLocal() {
	suite.runFileTests("local", []fileTest{
		{
			"local01.lua",
			nil,
			"",
			"6\n",
			"",
		},
		{
			"local02.lua",
			nil,
			"",
			"6\nnil\n",
			"",
		},
		{
			"local03.lua",
			nil,
			"",
			"local foo()\nnil\n",
			"",
		},
	})
}

func (suite *EngineSuite) TestWhile() {
	suite.runFileTests("while", []fileTest{
		{
			"while01.lua",
			nil,
			"",
			"a=0\na=1\na=2\na=3\na=4\nfinally a=5\n",
			"",
		},
	})
}

func (suite *EngineSuite) TestRepeat() {
	suite.runFileTests("repeat", []fileTest{
		{
			"repeat01.lua",
			nil,
			"",
			"a=0\na=1\na=2\na=3\na=4\nfinally a=5\n",
			"",
		},
	})
}

func (suite *EngineSuite) TestComment() {
	suite.runFileTests("comment", []fileTest{
		{
			"comment01.lua",
			nil,
			"",
			"",
			"",
		},
	})
}

func (suite *EngineSuite) TestTable() {
	suite.runFileTests("table", []fileTest{
		{
			"table01.lua",
			nil,
			"",
			"",
			"",
		},
		{
			"table02.lua",
			nil,
			"",
			"9\n",
			"",
		},
		{
			"table03.lua",
			nil,
			"",
			"9\n10\n",
			"",
		},
		{
			"table04.lua",
			nil,
			"",
			"nil\nfoobar\n",
			"",
		},
		{
			"table05.lua",
			nil,
			"",
			"foobar\n",
			"",
		},
	})
}

func (suite *EngineSuite) TestAssert() {
	suite.runFileTests("assert", []fileTest{
		{
			"assert01.lua",
			nil,
			"",
			"",
			"",
		},
		{
			"assert02.lua",
			nil,
			"must happen",
			"",
			"",
		},
	})
}

func (suite *EngineSuite) TestBinop() {
	suite.runFileTests("binop", []fileTest{
		{
			"binop01.lua",
			nil,
			"",
			"450\n",
			"",
		},
		{
			"binop02.lua",
			nil,
			"",
			"3\n-1\n2\n0.5\n0\n-1\n",
			"",
		},
	})
}

func (suite *EngineSuite) TestDo() {
	suite.runFileTests("do", []fileTest{
		{
			"do01.lua",
			nil,
			"",
			"Hello\nnil\n",
			"",
		},
	})
}

func (suite *EngineSuite) TestAssign() {
	suite.runFileTests("assign", []fileTest{
		{
			"assign01.lua",
			nil,
			"",
			"b\ta\n",
			"",
		},
		{
			"assign02.lua",
			nil,
			"",
			"1\n",
			"",
		},
	})
}

func (suite *EngineSuite) TestSemicolon() {
	suite.runFileTests("semicolon", []fileTest{
		{
			"semicolon01.lua",
			nil,
			"",
			"nil\nnil\nnil\n",
			"",
		},
	})
}

func (suite *EngineSuite) TestFunction() {
	suite.runFileTests("function", []fileTest{
		{
			"function01.lua",
			nil,
			"",
			"Hello\n",
			"",
		},
		{
			"function02.lua",
			nil,
			"",
			"bye\nhello\n",
			"",
		},
		{
			"function03.lua",
			nil,
			"",
			"Hello, World!\n",
			"",
		},
	})
}

func (suite *EngineSuite) TestReturn() {
	suite.runFileTests("return", []fileTest{
		{
			"return01.lua",
			[]value.Value{
				value.NewString("hello"),
			},
			"",
			"",
			"",
		},
		{
			"return02.lua",
			[]value.Value{
				value.NewString("hello"),
			},
			"",
			"",
			"",
		},
	})
}

func (suite *EngineSuite) TestSelect() {
	suite.runFileTests("select", []fileTest{
		{
			"select01.lua",
			nil,
			"",
			"3\n",
			"",
		},
	})
}

func (suite *EngineSuite) TestIf() {
	suite.runFileTests("if", []fileTest{
		{
			"if01.lua",
			nil,
			"",
			"Hello\tWorld\n",
			"",
		},
	})
}

func (suite *EngineSuite) TestPcall() {
	suite.runFileTests("pcall", []fileTest{
		{
			"pcall01.lua",
			nil,
			"",
			"false	error message\n",
			"",
		},
		{
			"pcall02.lua",
			nil,
			"",
			"print message\ntrue\n",
			"",
		},
	})
}

func (suite *EngineSuite) TestDofile() {
	suite.runFileTests("dofile", []fileTest{
		{
			"dofile01.lua",
			nil,
			"",
			"Goodbye\n",
			"",
		},
		{
			"dofile02.lua",
			nil,
			"",
			"Hello\n",
			"",
		},
	})
}

func (suite *EngineSuite) TestErrors() {
	suite.runFileTests("errors", []fileTest{
		{
			"error01.lua",
			nil,
			"error called with <nil>",
			"",
			"",
		},
		{
			"error02.lua",
			nil,
			"custom message",
			"",
			"",
		},
		{
			"error03.lua",
			nil,
			"expected error message",
			"line 1 on stdout\nline 2 on stdout\nline 3 on stdout\n",
			"",
		},
	})
}

type fileTest struct {
	file        string
	wantResults []value.Value
	wantErr     string
	wantStdout  string
	wantStderr  string
}

func (suite *EngineSuite) runFileTests(basePath string, tests []fileTest) {
	for _, test := range tests {
		suite.Run("file="+test.file, func() {
			stdin := new(bytes.Buffer)
			stdout := new(bytes.Buffer)
			stderr := new(bytes.Buffer)

			engine := New(
				WithStdin(stdin),
				WithStdout(stdout),
				WithStderr(stderr),
				WithClock(mockClock{}),
				WithFs(afero.NewBasePathFs(suite.testdata, basePath)),
			)

			file, err := suite.testdata.Open(filepath.Join(basePath, test.file))
			suite.Require().NoError(err)
			defer func() { _ = file.Close() }()

			gotResults, gotErr := engine.Eval(file)
			if gotErr != nil {
				if test.wantErr != "" {
					suite.IsType(Error{}, gotErr)
				}
				suite.EqualErrorf(gotErr, test.wantErr, "%s", gotErr)
			}

			suite.Equal(len(test.wantResults), len(gotResults), "amount of results not equal")
			resultsLen := len(test.wantResults)
			if len(gotResults) < resultsLen {
				resultsLen = len(gotResults)
			}

			for i := 0; i < resultsLen; i++ {
				expected := test.wantResults[i]
				got := gotResults[i]
				suite.Equal(expected.Type(), got.Type(), "expected %s, but got %s", expected.Type(), got.Type())
				if expected.Type() != got.Type() {
					continue
				}
				switch expected.Type() {
				case value.TypeString:
					suite.EqualValues(expected.(value.String), got.(value.String))
					suite.Equal(expected.(value.String).String(), got.(value.String).String())
				case value.TypeNumber:
					suite.EqualValues(expected.(value.Number), got.(value.Number))
					suite.Equal(expected.(value.Number).Value(), got.(value.Number).Value())
				case value.TypeBoolean:
					suite.EqualValues(expected.(value.Boolean), got.(value.Boolean))
					suite.Equal(expected.(value.Boolean).String(), got.(value.Boolean).String())
				default:
					suite.Failf("unsupported type", "type %s not supported yet", expected.Type())
				}
			}

			suite.Equal(test.wantStdout, stdout.String())
			suite.Equal(test.wantStderr, stderr.String())

			suite.T().Logf("stdout (%d bytes):\n%q", len(stdout.Bytes()), stdout.String())
			suite.T().Logf("stderr (%d bytes):\n%q", len(stderr.Bytes()), stderr.String())
		})
	}
}
