package engine

import (
	"bytes"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/suite"
)

func TestEngineSuite(t *testing.T) {
	suite.Run(t, new(EngineSuite))
}

type EngineSuite struct {
	suite.Suite

	testdata afero.Fs

	engine *Engine
	stdin  *bytes.Buffer
	stdout *bytes.Buffer
	stderr *bytes.Buffer
}

func (suite *EngineSuite) SetupTest() {
	suite.testdata = afero.NewBasePathFs(afero.NewOsFs(), "testdata")

	suite.stdin = new(bytes.Buffer)
	suite.stdout = new(bytes.Buffer)
	suite.stderr = new(bytes.Buffer)

	suite.engine = New(
		WithStdin(suite.stdin),
		WithStdout(suite.stdout),
		WithStderr(suite.stderr),
		WithClock(mockClock{}),
	)
}

func (suite *EngineSuite) TearDownTest() {
	suite.T().Logf("stdout (%d bytes):\n%q", len(suite.stdout.Bytes()), suite.stdout.String())
	suite.T().Logf("stderr (%d bytes):\n%q", len(suite.stderr.Bytes()), suite.stderr.String())
}
