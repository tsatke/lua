package ctxfunc_test

import (
	"path/filepath"
	"testing"

	"github.com/TimSatke/app/internal/tools/analysis/ctxfunc"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	dir, err := filepath.Abs("./testdata")
	if err != nil {
		t.Error(err)
	}
	analysistest.Run(t, dir, ctxfunc.Analyzer, "./...")
}
