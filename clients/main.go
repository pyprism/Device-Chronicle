package main

import (
	"fmt"
	"github.com/shirou/gopsutil/v4/common"
	"github.com/shirou/gopsutil/v4/sensors"
	"runtime"
	"strings"
)

func main() {
	v := sensors.ReadTemperaturesArm()
	sum := 0.0
	counter := 0
	for _, value := range v {
		//fmt.Printf("Key: %v, Temp: %v\n", value.SensorKey, value.Temperature)
		if strings.Contains(value.SensorKey, "tdie") {
			fmt.Printf("Key: %v Temperature: %v\n", value.SensorKey, value.Temperature)
			sum += value.Temperature
			counter++
		}
	}
	average := sum / float64(counter)
	fmt.Printf("Average: %v\n", average)
	fmt.Println(runtime.GOOS)
	fmt.Println(common.HostProcEnvKey)
	//for _, value := range v {
	//	if strings.Contains(strings.ToLower(strconv.Itoa(key)), "tdie") || strings.Contains(strings.ToLower(strconv.Itoa(key)), "tdie") {
	//		fmt.Printf("Temperature: %v\n", value)
	//	}
	//}
}
