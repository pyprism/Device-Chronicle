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
	"os"
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

// readSysfsTemp reads temperature from sysfs (for ARM devices)
func readSysfsTemp(path string) (float64, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	var temp float64
	_, err = fmt.Sscanf(string(data), "%f", &temp)
	// Convert from milliCelsius to Celsius
	return temp / 1000.0, err
}

// collectTemperatureData gathers temperature information
func collectTemperatureData(s *models.System) {
	if isARM() {
		// ARM-specific code - scan multiple thermal zones
		cpuTemp := 0.0
		chipsetTemps := []float64{}

		// Try to find all thermal zones
		for i := 0; i < 10; i++ {
			zonePath := fmt.Sprintf("/sys/class/thermal/thermal_zone%d/temp", i)
			typePath := fmt.Sprintf("/sys/class/thermal/thermal_zone%d/type", i)

			if _, err := os.Stat(zonePath); os.IsNotExist(err) {
				continue
			}

			temp, err := readSysfsTemp(zonePath)
			if err != nil {
				continue
			}

			zoneType := "unknown"
			if typeData, err := os.ReadFile(typePath); err == nil {
				zoneType = strings.TrimSpace(string(typeData))
			}

			if strings.Contains(strings.ToLower(zoneType), "cpu") || i == 0 {
				cpuTemp = temp
			} else {
				chipsetTemps = append(chipsetTemps, temp)
				s.Custom[fmt.Sprintf("thermal_zone_%d_%s", i, zoneType)] = fmt.Sprintf("%.2f°C", temp)
			}
		}

		s.CPUTemp = fmt.Sprintf("%.2f°C", cpuTemp)

		if len(chipsetTemps) > 0 {
			sum := 0.0
			for _, temp := range chipsetTemps {
				sum += temp
			}
			s.AverageChipsetTemp = fmt.Sprintf("%.2f°C", sum/float64(len(chipsetTemps)))
		}
	} else {
		// Enhanced x86 approach for various motherboard manufacturers
		temps, _ := sensors.SensorsTemperatures()

		// For chipset/motherboard temperatures
		chipsetTemps := []float64{}
		cpuTemp := 0.0

		// Track if we've found at least one sensor
		foundCPUSensor := false

		// Common sensor patterns by manufacturer
		chipsetPatterns := []string{
			"wmi",     // Gigabyte
			"pch_",    // Intel PCH
			"system",  // Common name
			"board",   // Common name
			"chipset", // Generic
			"sbr",     // South Bridge
			"nbr",     // North Bridge
			"asus",    // ASUS
			"msi",     // MSI
			"asrock",  // ASRock
			"mb",      // Motherboard
		}

		cpuPatterns := []string{
			"tctl",    // AMD Tctl
			"tdie",    // AMD Tdie
			"core",    // Intel Core temps
			"cpu",     // Generic CPU
			"package", // CPU package
			"k10temp", // AMD K10
		}

		// First pass: look for CPU temperature
		for _, value := range temps {
			sensorKey := strings.ToLower(value.SensorKey)

			// Check for CPU temperature sensors
			for _, pattern := range cpuPatterns {
				if strings.Contains(sensorKey, pattern) {
					cpuTemp = value.Temperature
					foundCPUSensor = true
					// Prefer core/package sensors if found
					if strings.Contains(sensorKey, "core") ||
						strings.Contains(sensorKey, "package") {
						break
					}
				}
			}
		}

		// Second pass: look for chipset temperatures
		for _, value := range temps {
			sensorKey := strings.ToLower(value.SensorKey)

			for _, pattern := range chipsetPatterns {
				if strings.Contains(sensorKey, pattern) {
					chipsetTemps = append(chipsetTemps, value.Temperature)
					// Add individual sensor to custom fields
					s.Custom[fmt.Sprintf("sensor_%s", value.SensorKey)] = fmt.Sprintf("%.2f°C", value.Temperature)
					break
				}
			}
		}

		// Calculate average chipset temperature
		if len(chipsetTemps) > 0 {
			sum := 0.0
			for _, temp := range chipsetTemps {
				sum += temp
			}
			s.AverageChipsetTemp = fmt.Sprintf("%.2f°C", sum/float64(len(chipsetTemps)))
		}

		// If we found a CPU temperature, use it
		if foundCPUSensor {
			s.CPUTemp = fmt.Sprintf("%.2f°C", cpuTemp)
		}
	}
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
	// Per-core CPU usage (works for both architectures)
	cpu_, _ := cpu.Percent(1*time.Second, true)
	for i, percentage := range cpu_ {
		s.CPUCores[fmt.Sprintf("cpu_core_%d", i)] = fmt.Sprintf("%.2f", percentage)
	}

	// Overall CPU percentage
	totalCPU, _ := cpu.Percent(0, false)
	if len(totalCPU) > 0 {
		s.CPUUsage = fmt.Sprintf("%.2f%%", totalCPU[0])
	}

	// Get CPU frequency based on architecture
	if isARM() {
		// ARM-specific frequency reading
		freqPath := "/sys/devices/system/cpu/cpu0/cpufreq/scaling_cur_freq"
		if freqData, err := os.ReadFile(freqPath); err == nil {
			var freq float64
			fmt.Sscanf(string(freqData), "%f", &freq)
			s.CPUMHZ = fmt.Sprintf("%.0f MHz", freq/1000) // Convert KHz to MHz
		}
	} else {
		// x86 frequency calculation
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
