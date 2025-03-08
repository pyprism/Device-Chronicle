package fetch

import (
	"device-chronicle-client/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"runtime"
	"testing"
)

// MockOS is a mock that simulates the OS interface
type MockOS struct {
	mock.Mock
}

// Linux mocks the os.Linux function
func (m *MockOS) Linux() (*models.System, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.System), args.Error(1)
}

// TestFetchData tests the FetchData function
func TestFetchData(t *testing.T) {
	// We can only test the expected behavior for the current OS
	system, err := FetchData()

	if runtime.GOOS == "linux" {
		assert.NoError(t, err)
		assert.NotNil(t, system)

		// Verify system has expected structure
		assert.NotEmpty(t, system.Hostname)
		assert.NotNil(t, system.CPUCores)
	} else {
		assert.Error(t, err)
		assert.Nil(t, system)
		assert.Contains(t, err.Error(), "unsupported OS")
	}
}
