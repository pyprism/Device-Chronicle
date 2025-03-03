package os

import (
	"device-chronicle-client/models"
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

func Linux() (*models.System, error) {
	s := models.NewSystem()

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
		s.PacketsSent = utils.FormatBytes(network[0].BytesSent - prevNetworkUsage.BytesSent)
		s.PacketsReceive = utils.FormatBytes(network[0].BytesRecv - prevNetworkUsage.BytesRecv)
	} else {
		s.PacketsSent = utils.FormatBytes(network[0].BytesSent)
		s.PacketsReceive = utils.FormatBytes(network[0].BytesRecv)
	}

	// Update previous network usage
	prevNetworkUsage = &network[0]

	average = sum / float64(counter)
	s.AverageChipsetTemp = fmt.Sprintf("%.2f°C", average)

	// CPU usage per core - stored in the CPUCores map
	for i, percentage := range cpu_ {
		s.CPUCores[fmt.Sprintf("cpu_core_%d", i)] = fmt.Sprintf("%.2f", percentage)
	}

	// Get overall CPU percentage
	totalCPU, _ := cpu.Percent(0, false)
	if len(totalCPU) > 0 {
		s.CPUUsage = fmt.Sprintf("%.2f%%", totalCPU[0])
	}

	// Get disk usage
	partitions, _ := disk.Partitions(true)
	var totalDiskSpace uint64
	var usedDiskSpace uint64
	var freeDiskSpace uint64

	for _, partition := range partitions {
		// Skip pseudo filesystems
		if !strings.HasPrefix(partition.Fstype, "ext") &&
			!strings.HasPrefix(partition.Fstype, "xfs") &&
			!strings.HasPrefix(partition.Fstype, "btrfs") &&
			!strings.HasPrefix(partition.Fstype, "ntfs") &&
			partition.Fstype != "vfat" &&
			partition.Fstype != "fat32" {
			continue
		}

		diskUsage, err := disk.Usage(partition.Mountpoint)
		if err != nil {
			continue // Skip partitions with errors
		}
		totalDiskSpace += diskUsage.Total
		usedDiskSpace += diskUsage.Used
		freeDiskSpace += diskUsage.Free
	}

	// Calculate used percentage
	usedPercent := 0.0
	if totalDiskSpace > 0 {
		usedPercent = float64(usedDiskSpace) / float64(totalDiskSpace) * 100.0
	}

	s.DiskTotal = utils.FormatBytes(totalDiskSpace)
	s.DiskFree = utils.FormatBytes(freeDiskSpace)
	s.DiskUsed = utils.FormatBytes(usedDiskSpace)
	s.DiskUsagePercent = fmt.Sprintf("%.2f%%", usedPercent)

	// Load average
	loadAvg, _ := load.Avg()
	s.LoadAvg1 = fmt.Sprintf("%.2f", loadAvg.Load1)
	s.LoadAvg5 = fmt.Sprintf("%.2f", loadAvg.Load5)
	s.LoadAvg15 = fmt.Sprintf("%.2f", loadAvg.Load15)

	// Number of running processes
	processes, _ := process.Processes()
	s.ProcessCount = len(processes)

	// Swap memory
	swap, _ := mem.SwapMemory()
	s.SwapUsed = utils.FormatBytes(swap.Used)
	s.SwapTotal = utils.FormatBytes(swap.Total)
	s.SwapPercent = fmt.Sprintf("%.2f%%", swap.UsedPercent)

	// CPU frequency
	freqs, err := cpu.Percent(100*time.Millisecond, true)
	if err == nil && len(freqs) > 0 {
		cpuInfo, _ := cpu.Info()
		maxFreq := 0.0
		if len(cpuInfo) > 0 {
			maxFreq = cpuInfo[0].Mhz
		}

		currentFreq := 0.0
		for _, f := range freqs {
			currentFreq += (f / 100.0) * maxFreq
		}
		currentFreq /= float64(len(freqs))

		s.CPUMHZ = fmt.Sprintf("%.0f MHz", currentFreq)
	}

	// Format uptime
	uptimeDuration := time.Duration(host_.Uptime) * time.Second
	days := int(uptimeDuration.Hours()) / 24
	hours := int(uptimeDuration.Hours()) % 24
	minutes := int(uptimeDuration.Minutes()) % 60
	s.Uptime = fmt.Sprintf("%dd %dh %dm", days, hours, minutes)

	// Add remaining fields
	s.CPUTemp = fmt.Sprintf("%.2f°C", cpuTemp)
	s.TotalRAM = utils.FormatBytes(memory.Total)
	s.FreeRAM = utils.FormatBytes(memory.Free)
	s.UsedRAM = utils.FormatBytes(memory.Used)
	s.UsedRAMPercentage = fmt.Sprintf("%.2f%%", memory.UsedPercent)
	s.Hostname = host_.Hostname

	return s, nil
}
