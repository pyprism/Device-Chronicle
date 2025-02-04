package os

import (
	"device-chronicle-client/utils"
	"fmt"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/sensors"
	"strings"
	"time"
)

func Linux() (map[string]interface{}, error) {
	data := make(map[string]interface{})

	// Get data from sensors and system
	v, _ := sensors.SensorsTemperatures()
	memory, _ := mem.VirtualMemory()
	network, _ := net.IOCounters(false)
	cpu_, _ := cpu.Percent(1*time.Second, true)
	sum := 0.0
	counter := 0
	average := 0.0
	cpuTemp := 0.0
	host_, _ := host.Info()

	// Calculate average chipset temperature
	for _, value := range v {
		// chipset sensors
		if strings.Contains(value.SensorKey, "wmi") { // ex: gigabyte_wmi
			sum += value.Temperature
			counter++
		}
		// cpu temp
		if strings.Contains(value.SensorKey, "tctl") {
			cpuTemp = value.Temperature
		}
	}

	average = sum / float64(counter)
	formattedAverage := fmt.Sprintf("%.2f", average)

	// CPU usage per core
	for i, percentage := range cpu_ {
		data[fmt.Sprintf("cpu_core_%d", i)] = fmt.Sprintf("%.2f", percentage)
	}
	// Add data to map
	data["average_chipset_temp"] = formattedAverage
	data["cpu_temp"] = cpuTemp
	data["total_ram"] = utils.FormatBytes(memory.Total)
	data["free_ram"] = utils.FormatBytes(memory.Free)
	data["used_ram"] = utils.FormatBytes(memory.Used)
	data["used_ram_percentage"] = fmt.Sprintf("%.2f", memory.UsedPercent)
	data["packets_sent"] = utils.FormatBytes(network[0].PacketsSent)
	data["packets_receive"] = utils.FormatBytes(network[0].PacketsRecv)
	data["hostname"] = host_.Hostname
	return data, nil
}
