package os

import (
	"github.com/stretchr/testify/assert"
	"runtime"
	"testing"
)

func TestIsARM(t *testing.T) {
	expected := runtime.GOARCH == "arm" || runtime.GOARCH == "arm64"
	assert.Equal(t, expected, isARM(), "isARM() should correctly identify ARM architecture")
}
