package engine

import (
	"bytes"
	"path/filepath"

	"github.com/tsatke/lua/internal/engine/value"
)

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
			)

			file, err := suite.testdata.Open(filepath.Join(basePath, test.file))
			suite.Require().NoError(err)
			defer func() { _ = file.Close() }()

			gotResults, gotErr := engine.Eval(file)
			suite.EqualError(gotErr, test.wantErr)
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