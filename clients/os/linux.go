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

// Linux collects system information and returns it as a System struct
func Linux() (*models.System, error) {
	s := models.NewSystem()

	collectTemperatureData(s)
	collectNetworkData(s)
	collectCPUData(s)
	collectDiskData(s)
	collectSystemLoadData(s)
	collectProcessData(s)
	collectMemoryData(s)
	collectSwapData(s)
	collectHostData(s)

	return s, nil
}

// collectTemperatureData gathers temperature information
func collectTemperatureData(s *models.System) {
	temps, _ := sensors.SensorsTemperatures()

	sum := 0.0
	counter := 0
	cpuTemp := 0.0

	for _, value := range temps {
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

	if counter > 0 {
		average := sum / float64(counter)
		s.AverageChipsetTemp = fmt.Sprintf("%.2f°C", average)
	}

	s.CPUTemp = fmt.Sprintf("%.2f°C", cpuTemp)
}

// collectNetworkData gathers network usage information
func collectNetworkData(s *models.System) {
	network, _ := net.IOCounters(false)

	if prevNetworkUsage != nil {
		s.PacketsSent = utils.FormatBytes(network[0].BytesSent - prevNetworkUsage.BytesSent)
		s.PacketsReceive = utils.FormatBytes(network[0].BytesRecv - prevNetworkUsage.BytesRecv)
	} else {
		s.PacketsSent = utils.FormatBytes(network[0].BytesSent)
		s.PacketsReceive = utils.FormatBytes(network[0].BytesRecv)
	}

	prevNetworkUsage = &network[0]
}

// collectCPUData gathers CPU usage and frequency information
func collectCPUData(s *models.System) {
	// Per-core CPU usage
	cpu_, _ := cpu.Percent(1*time.Second, true)
	for i, percentage := range cpu_ {
		s.CPUCores[fmt.Sprintf("cpu_core_%d", i)] = fmt.Sprintf("%.2f", percentage)
	}

	// Overall CPU percentage
	totalCPU, _ := cpu.Percent(0, false)
	if len(totalCPU) > 0 {
		s.CPUUsage = fmt.Sprintf("%.2f%%", totalCPU[0])
	}

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
}

// collectDiskData gathers disk space information
func collectDiskData(s *models.System) {
	partitions, _ := disk.Partitions(true)
	var totalDiskSpace, usedDiskSpace, freeDiskSpace uint64

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
			continue
		}
		totalDiskSpace += diskUsage.Total
		usedDiskSpace += diskUsage.Used
		freeDiskSpace += diskUsage.Free
	}

	usedPercent := 0.0
	if totalDiskSpace > 0 {
		usedPercent = float64(usedDiskSpace) / float64(totalDiskSpace) * 100.0
	}

	s.DiskTotal = utils.FormatBytes(totalDiskSpace)
	s.DiskFree = utils.FormatBytes(freeDiskSpace)
	s.DiskUsed = utils.FormatBytes(usedDiskSpace)
	s.DiskUsagePercent = fmt.Sprintf("%.2f%%", usedPercent)
}

// collectSystemLoadData gathers load average information
func collectSystemLoadData(s *models.System) {
	loadAvg, _ := load.Avg()
	s.LoadAvg1 = fmt.Sprintf("%.2f", loadAvg.Load1)
	s.LoadAvg5 = fmt.Sprintf("%.2f", loadAvg.Load5)
	s.LoadAvg15 = fmt.Sprintf("%.2f", loadAvg.Load15)
}

// collectProcessData gathers information about running processes
func collectProcessData(s *models.System) {
	processes, _ := process.Processes()
	s.ProcessCount = len(processes)
}

// collectMemoryData gathers RAM usage information
func collectMemoryData(s *models.System) {
	memory, _ := mem.VirtualMemory()
	s.TotalRAM = utils.FormatBytes(memory.Total)
	s.FreeRAM = utils.FormatBytes(memory.Free)
	s.UsedRAM = utils.FormatBytes(memory.Used)
	s.UsedRAMPercentage = fmt.Sprintf("%.2f%%", memory.UsedPercent)
}

// collectSwapData gathers swap memory information
func collectSwapData(s *models.System) {
	swap, _ := mem.SwapMemory()
	s.SwapUsed = utils.FormatBytes(swap.Used)
	s.SwapTotal = utils.FormatBytes(swap.Total)
	s.SwapPercent = fmt.Sprintf("%.2f%%", swap.UsedPercent)
}

// collectHostData gathers host information like hostname and uptime
func collectHostData(s *models.System) {
	hostInfo, _ := host.Info()
	s.Hostname = hostInfo.Hostname

	uptimeDuration := time.Duration(hostInfo.Uptime) * time.Second
	days := int(uptimeDuration.Hours()) / 24
	hours := int(uptimeDuration.Hours()) % 24
	minutes := int(uptimeDuration.Minutes()) % 60
	s.Uptime = fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
}
