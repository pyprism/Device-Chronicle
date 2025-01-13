package main

import (
	"fmt"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/sensors"
	"runtime"
	"strings"
)

func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func main() {
	//v := sensors.ReadTemperaturesArm()
	//sum := 0.0
	//counter := 0
	//for _, value := range v {
	//	//fmt.Printf("Key: %v, Temp: %v\n", value.SensorKey, value.Temperature)
	//	if strings.Contains(value.SensorKey, "tdie") {
	//		fmt.Printf("Key: %v Temperature: %v\n", value.SensorKey, value.Temperature)
	//		sum += value.Temperature
	//		counter++
	//	}
	//}
	//average := sum / float64(counter)
	//fmt.Printf("Average: %v\n", average)
	//fmt.Println(runtime.GOOS)
	//fmt.Println(common.HostProcEnvKey)

	//for _, value := range v {
	//	if strings.Contains(strings.ToLower(strconv.Itoa(key)), "tdie") || strings.Contains(strings.ToLower(strconv.Itoa(key)), "tdie") {
	//		fmt.Printf("Temperature: %v\n", value)
	//	}
	//}
	// if linux
	if runtime.GOOS == "linux" {
		v, _ := sensors.SensorsTemperatures()
		memory, _ := mem.VirtualMemory()
		network, _ := net.IOCounters(false)
		cpu, _ := cpu.Info()
		sum := 0.0
		counter := 0
		average := 0.0
		cpuTemp := 0.0

		for _, value := range v {
			fmt.Printf("Key: %v, Temp: %v\n", value.SensorKey, value.Temperature)
			// chipset sensors
			if strings.Contains(value.SensorKey, "wmi") {
				sum += value.Temperature
				counter++
			}
			// cpu temp
			if strings.Contains(value.SensorKey, "tctl") {
				cpuTemp = value.Temperature
			}
		}
		average = sum / float64(counter)
		fmt.Printf("Average chipset temp: %.2f\n", average)
		fmt.Printf("CPU temp: %v\n", cpuTemp)
		fmt.Printf("Total: %v\n", formatBytes(memory.Total))
		fmt.Printf("Free: %v\n", formatBytes(memory.Free))
		fmt.Printf("Used: %v\n", formatBytes(memory.Used))
		fmt.Printf("Used percentage: %.2f\n", memory.UsedPercent)
		fmt.Println(formatBytes(network[0].PacketsSent))
		fmt.Println(formatBytes(network[0].PacketsRecv))
		fmt.Println(cpu)

	}
}
