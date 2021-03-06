package nopanic_test

import (
	"path/filepath"
	"testing"

	"github.com/tsatke/lua/internal/tools/analysis/nopanic"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	dir, err := filepath.Abs("./testdata")
	if err != nil {
		t.Error(err)
	}
	analysistest.Run(t, dir, nopanic.Analyzer, "./...")
}
