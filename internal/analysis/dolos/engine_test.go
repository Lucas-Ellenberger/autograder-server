package dolos

import (
	"runtime"
	"testing"

	"github.com/edulinq/autograder/internal/analysis/core"
	"github.com/edulinq/autograder/internal/model"
)

func TestDolosComputeFileSimilarityBase(test *testing.T) {
	if runtime.GOARCH != "amd64" {
		test.Skip("Dolos only runs on amd64.")
	}

	expected := &model.FileSimilarity{
		Filename: "submission.py",
		Tool:     NAME,
		Version:  VERSION,
		Score:    0.717949,
	}

	core.RunEngineTestComputeFileSimilarityBase(test, GetEngine(), false, expected)
}

func TestDolosComputeFileSimilarityWithIgnoreBase(test *testing.T) {
	if runtime.GOARCH != "amd64" {
		test.Skip("Dolos only runs on amd64.")
	}

	expected := &model.FileSimilarity{
		Filename: "submission.py",
		Tool:     NAME,
		Version:  VERSION,
		Score:    0.702703,
	}

	core.RunEngineTestComputeFileSimilarityBase(test, GetEngine(), true, expected)
}
