package fetch

import (
	"device-chronicle-client/models"
	"device-chronicle-client/os"
	"fmt"
	"runtime"
)

func FetchData() (*models.System, error) {
	if runtime.GOOS == "linux" {
		return os.Linux()
	}
	return nil, fmt.Errorf("unsupported OS")
}
