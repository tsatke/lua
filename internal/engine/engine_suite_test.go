package engine

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestEngineSuite(t *testing.T) {
	suite.Run(t, new(EngineSuite))
}

type EngineSuite struct {
	suite.Suite

	engine *Engine
	stdin  *bytes.Buffer
	stdout *bytes.Buffer
	stderr *bytes.Buffer
}

func (suite *EngineSuite) SetupTest() {
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
	suite.T().Logf("stdout (%d bytes):\n%s", len(suite.stdout.Bytes()), suite.stdout.String())
	suite.T().Logf("stderr (%d bytes):\n%s", len(suite.stderr.Bytes()), suite.stderr.String())
}
