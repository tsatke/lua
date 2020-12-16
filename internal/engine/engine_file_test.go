package engine

import (
	"bytes"
	"path/filepath"

	"github.com/spf13/afero"
	"github.com/tsatke/lua/internal/engine/value"
)

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
			"false	error message\nprint message\ntrue\n",
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
