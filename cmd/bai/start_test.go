package main

import (
	"fmt"
	"testing"

	test_utils "github.com/biisal/bai/utils/tests"
)

func TestStart(t *testing.T) {
	t.Run("Should throw error when invalid config path provided",
		func(t *testing.T) {
			err := start("invalid path", false)
			test_utils.AssertError(t, err, fmt.Errorf("failed to load config: config file does not exists: invalid path"))
		})
}
