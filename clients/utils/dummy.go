package utils

import (
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
func DummyData() map[string]interface{} {
	data := make(map[string]interface{})
	data["packets_sent"], _ = randomNumber(500, 1000)
	data["cpu_core_1"], _ = randomNumber(50.0, 60.99)
	data["packets_receive"], _ = randomNumber(500, 1000)
	data["average_chipset_temp"], _ = randomNumber(40.0, 70.0)
	data["cpu_temp"], _ = randomNumber(50.0, 60.0)
	data["total_ram"], _ = randomNumber(50, 80)
	data["free_ram"], _ = randomNumber(60, 80)
	data["user_ram"], _ = randomNumber(50, 90)
	data["used_ram_percentage"], _ = randomNumber(60.0, 80.0)
	data["hostname"] = "dummy-hostname"
	data["load_avg"], _ = randomNumber(0.1, 5.0)
	data["process_count"], _ = randomNumber(100, 300)
	data["swap_usage"], _ = randomNumber(0, 100)
	data["cpu_freq"], _ = randomNumber(1.0, 3.5)
	data["uptime"] = "1d 2h 3m"
	return data
}
