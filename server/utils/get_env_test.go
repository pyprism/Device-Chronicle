package utils

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestGetEnv(t *testing.T) {
	// Test case: No environment variable set
	os.Unsetenv("MY_ENV")
	value := GetEnv("MY_ENV", "default_value")
	assert.Equal(t, "default_value", value)

	// Test case: Environment variable set
	os.Setenv("MY_ENV", "actual_value")
	defer os.Unsetenv("MY_ENV")
	value = GetEnv("MY_ENV", "default_value")
	assert.Equal(t, "actual_value", value)
}
