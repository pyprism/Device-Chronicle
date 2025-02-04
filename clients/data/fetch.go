package data

import (
	"device-chronicle-client/os"
	"fmt"
	"runtime"
)

func FetchData() (map[string]interface{}, error) {
	if runtime.GOOS == "linux" {
		return os.Linux()
	}
	return nil, fmt.Errorf("unsupported OS")
}
