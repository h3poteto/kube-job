package job

import (
	"testing"

	"github.com/pkg/errors"
)

func TestShouldCleanupAllWithError(t *testing.T) {
	err := errors.New("test")
	result := shouldCleanup("all", err)
	if !result {
		t.Error("should be true when specified 'all'")
	}
}

func TestShouldCleanupAllWithoutError(t *testing.T) {
	result := shouldCleanup("all", nil)
	if !result {
		t.Error("should be true when specified 'all'")
	}
}

func TestShouldCleanupSucceededWithError(t *testing.T) {
	err := errors.New("test")
	result := shouldCleanup("succeeded", err)
	if result {
		t.Error("should be false when specified 'succeeded' with error")
	}
}

func TestShouldCleanupSucceededWithoutError(t *testing.T) {
	result := shouldCleanup("succeeded", nil)
	if !result {
		t.Error("should be true when specified 'succeeded' without error")
	}
}

func TestShouldCleanupFailedWithError(t *testing.T) {
	err := errors.New("test")
	result := shouldCleanup("failed", err)
	if !result {
		t.Error("should be true when specified 'failed' with error")
	}
}

func TestShouldCleanupFailedWithoutError(t *testing.T) {
	result := shouldCleanup("failed", nil)
	if result {
		t.Error("should be false when specified 'failed' without error")
	}
}
