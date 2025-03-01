package os

import (
	"device-chronicle-client/utils"
	"fmt"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/process"
	"github.com/shirou/gopsutil/v4/sensors"
	"strings"
	"time"
)

var prevNetworkUsage *net.IOCountersStat

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

	// Calculate network usage difference
	if prevNetworkUsage != nil {
		data["packets_sent"] = utils.FormatBytes(network[0].BytesSent - prevNetworkUsage.BytesSent)
		data["packets_receive"] = utils.FormatBytes(network[0].BytesRecv - prevNetworkUsage.BytesRecv)
	} else {
		data["packets_sent"] = utils.FormatBytes(network[0].BytesSent)
		data["packets_receive"] = utils.FormatBytes(network[0].BytesRecv)
	}

	// Update previous network usage
	prevNetworkUsage = &network[0]

	average = sum / float64(counter)
	formattedAverage := fmt.Sprintf("%.2f°C", average)

	// CPU usage per core
	for i, percentage := range cpu_ {
		data[fmt.Sprintf("cpu_core_%d", i)] = fmt.Sprintf("%.2f", percentage)
	}

	// Get overall CPU percentage
	totalCPU, _ := cpu.Percent(0, false)
	if len(totalCPU) > 0 {
		data["cpu_usage"] = fmt.Sprintf("%.2f%%", totalCPU[0])
	}

	// Get disk usage
	diskUsage, _ := disk.Usage("/")
	data["disk_total"] = utils.FormatBytes(diskUsage.Total)
	data["disk_free"] = utils.FormatBytes(diskUsage.Free)
	data["disk_used"] = utils.FormatBytes(diskUsage.Used)
	data["disk_usage_percent"] = fmt.Sprintf("%.2f%%", diskUsage.UsedPercent)

	// load average
	loadAvg, _ := load.Avg()
	data["load_1"] = fmt.Sprintf("%.2f", loadAvg.Load1)
	data["load_5"] = fmt.Sprintf("%.2f", loadAvg.Load5)
	data["load_15"] = fmt.Sprintf("%.2f", loadAvg.Load15)

	// number iof running processes
	processes, _ := process.Processes()
	data["process_count"] = len(processes)

	// swap memory
	swap, _ := mem.SwapMemory()
	data["swap_used"] = utils.FormatBytes(swap.Used)
	data["swap_total"] = utils.FormatBytes(swap.Total)
	data["swap_percent"] = fmt.Sprintf("%.2f%%", swap.UsedPercent)

	// cpu freq
	// Get current CPU frequencies for all cores
	freqs, err := cpu.Percent(100*time.Millisecond, true)
	if err == nil && len(freqs) > 0 {
		// Get max frequency as reference
		cpuInfo, _ := cpu.Info()
		maxFreq := 0.0
		if len(cpuInfo) > 0 {
			maxFreq = cpuInfo[0].Mhz
		}

		// Calculate current frequency based on utilization percentage
		currentFreq := 0.0
		for _, f := range freqs {
			currentFreq += (f / 100.0) * maxFreq
		}
		currentFreq /= float64(len(freqs))

		data["cpu_mhz"] = fmt.Sprintf("%.0f MHz", currentFreq)
	}

	// Format uptime
	uptimeDuration := time.Duration(host_.Uptime) * time.Second
	days := int(uptimeDuration.Hours()) / 24
	hours := int(uptimeDuration.Hours()) % 24
	minutes := int(uptimeDuration.Minutes()) % 60
	data["uptime"] = fmt.Sprintf("%dd %dh %dm", days, hours, minutes)

	// Add data to map
	data["average_chipset_temp"] = formattedAverage
	data["cpu_temp"] = fmt.Sprintf("%.2f°C", cpuTemp)
	data["total_ram"] = utils.FormatBytes(memory.Total)
	data["free_ram"] = utils.FormatBytes(memory.Free)
	data["used_ram"] = utils.FormatBytes(memory.Used)
	data["used_ram_percentage"] = fmt.Sprintf("%.2f%%", memory.UsedPercent)
	data["hostname"] = host_.Hostname
	return data, nil
}
