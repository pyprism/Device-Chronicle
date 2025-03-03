package utils

import (
	"device-chronicle-client/models"
	"fmt"
	"math/rand"
	"time"
)

// randoNumber returns a random number between min and max.
// It accepts either int or float64 as parameters.
func randomNumber(min, max interface{}) (interface{}, error) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	switch min_ := min.(type) {
	case int:
		max_ := max.(int)
		return r.Intn(max_-min_+1) + min_, nil
	case float64:
		max_ := max.(float64)
		return min_ + r.Float64()*(max_-min_), nil
	default:
		return nil, fmt.Errorf("unsupported type")
	}
}

// DummyData returns dummy data for testing purposes.
func DummyData() *models.System {
	s := models.NewSystem()

	// Set fixed fields
	num, _ := randomNumber(500, 1000)
	s.PacketsSent = fmt.Sprintf("%v", num)

	num, _ = randomNumber(500, 1000)
	s.PacketsReceive = fmt.Sprintf("%v", num)

	num, _ = randomNumber(40.0, 70.0)
	s.AverageChipsetTemp = fmt.Sprintf("%.2f°C", num)

	// Set CPU cores
	for i := 0; i < 8; i++ {
		val, _ := randomNumber(50.0, 60.99)
		s.CPUCores[fmt.Sprintf("cpu_core_%d", i)] = fmt.Sprintf("%.2f", val)
	}

	s.CPUTemp = "50.0°C"
	s.TotalRAM = "16GB"
	s.FreeRAM = "8GB"
	s.UsedRAM = "8GB"
	s.UsedRAMPercentage = "50%"
	s.Hostname = "dummy-host"
	s.Uptime = "1d 2h 3m"
	s.LoadAvg1 = "0.5"
	s.LoadAvg5 = "0.6"
	s.LoadAvg15 = "0.7"
	s.ProcessCount = 100
	s.CPUUsage = "50%"
	s.CPUMHZ = "3.2GHz"
	s.DiskTotal = "1TB"
	s.DiskFree = "500GB"
	s.DiskUsed = "500GB"
	s.DiskUsagePercent = "50%"
	s.SwapUsed = "1GB"
	s.SwapTotal = "2GB"
	s.SwapPercent = "50%"
	return s
}
