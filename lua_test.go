package lua

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLua5_3_4(t *testing.T) {
	t.Skip("engine not fully functional yet")

	assert := assert.New(t)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	e := NewEngine(
		WithStdout(&stdout),
		WithStderr(&stderr),
		WithWorkingDirectory("testdata/lua-test"),
	)
	results, err := e.EvalFile("all.lua")
	assert.NoError(err)
	assert.Len(results, 0)
}
