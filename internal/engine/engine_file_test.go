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
			if len(gotResults) > 0 {
				panic("results not yet supported")
			}

			suite.Equal(test.wantStdout, stdout.String())
			suite.Equal(test.wantStderr, stderr.String())

			suite.T().Logf("stdout (%d bytes):\n%q", len(stdout.Bytes()), stdout.String())
			suite.T().Logf("stderr (%d bytes):\n%q", len(stderr.Bytes()), stderr.String())
		})
	}
}
