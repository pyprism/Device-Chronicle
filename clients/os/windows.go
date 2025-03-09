package os

import (
	"device-chronicle-client/models"
	"device-chronicle-client/utils"
	"fmt"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/process"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var prevWindowsNetworkUsage *net.IOCountersStat

// Windows collects system information on Windows OS and returns it as a System struct
func Windows() (*models.System, error) {
	s := models.NewSystem()

	collectWindowsTemperatureData(s)
	collectWindowsNetworkData(s)
	collectWindowsCPUData(s)
	collectWindowsDiskData(s)
	collectWindowsSystemLoadData(s)
	collectWindowsProcessData(s)
	collectWindowsMemoryData(s)
	collectWindowsSwapData(s)
	collectWindowsHostData(s)

	return s, nil
}

// collectWindowsTemperatureData gathers temperature information from Windows
func collectWindowsTemperatureData(s *models.System) {
	// Try to get temperature using wmic
	cmd := exec.Command("powershell", "-Command", "Get-WmiObject MSAcpi_ThermalZoneTemperature -Namespace \"root/wmi\"")
	output, err := cmd.Output()
	if err == nil {
		// Parse the temperature values
		tempRegex := regexp.MustCompile(`CurrentTemperature\s*:\s*(\d+)`)
		matches := tempRegex.FindAllStringSubmatch(string(output), -1)

		if len(matches) > 0 {
			// Windows reports temperature in tenths of Kelvin, convert to Celsius
			temps := []float64{}

			for i, match := range matches {
				if len(match) >= 2 {
					if tempVal, err := strconv.ParseFloat(match[1], 64); err == nil {
						// Convert from tenths of Kelvin to Celsius
						tempCelsius := (tempVal / 10.0) - 273.15
						temps = append(temps, tempCelsius)

						// Store individual sensor readings
						s.Custom[fmt.Sprintf("thermal_zone_%d", i)] = fmt.Sprintf("%.2f°C", tempCelsius)

						// Use the first reading for CPU temperature (approximation)
						if i == 0 && s.CPUTemp == "" {
							s.CPUTemp = fmt.Sprintf("%.2f°C", tempCelsius)
						}
					}
				}
			}

			// Calculate average temperature
			if len(temps) > 0 {
				sum := 0.0
				for _, temp := range temps {
					sum += temp
				}
				averageTemp := sum / float64(len(temps))
				s.AverageChipsetTemp = fmt.Sprintf("%.2f°C", averageTemp)
			}
		}
	}

	// Try using OpenHardwareMonitor if available
	cmd = exec.Command("powershell", "-Command",
		"If (Get-Module -ListAvailable -Name \"OpenHardwareMonitorLib\") { "+
			"Import-Module OpenHardwareMonitorLib; "+
			"$hw = New-Object OpenHardwareMonitor.Hardware.Computer; "+
			"$hw.CPUEnabled = $true; $hw.Open(); "+
			"$hw.Hardware | ForEach-Object { "+
			"$_.Sensors | Where-Object { $_.SensorType -eq 'Temperature' } | "+
			"ForEach-Object { $_.Name + ': ' + $_.Value } }"+
			"}")

	output, err = cmd.Output()
	if err == nil && len(output) > 0 {
		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		for _, line := range lines {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				name := strings.TrimSpace(parts[0])
				valStr := strings.TrimSpace(parts[1])
				if val, err := strconv.ParseFloat(valStr, 64); err == nil {
					s.Custom[fmt.Sprintf("sensor_%s", name)] = fmt.Sprintf("%.2f°C", val)

					// Use CPU package or core temps for CPU temperature
					if strings.Contains(strings.ToLower(name), "cpu") {
						if s.CPUTemp == "" || strings.Contains(strings.ToLower(name), "package") {
							s.CPUTemp = fmt.Sprintf("%.2f°C", val)
						}
					}
				}
			}
		}
	}
}

// collectWindowsNetworkData gathers network usage information
func collectWindowsNetworkData(s *models.System) {
	network, err := net.IOCounters(false)
	if err != nil || len(network) == 0 {
		return
	}

	if prevNetworkUsage != nil {
		s.PacketsSent = utils.FormatNetworkBytes(network[0].BytesSent - prevWindowsNetworkUsage.BytesSent)
		s.PacketsReceive = utils.FormatNetworkBytes(network[0].BytesRecv - prevWindowsNetworkUsage.BytesRecv)
	} else {
		s.PacketsSent = utils.FormatNetworkBytes(network[0].BytesSent)
		s.PacketsReceive = utils.FormatNetworkBytes(network[0].BytesRecv)
	}

	prevNetworkUsage = &network[0]

	// Add details for individual interfaces
	interfaces, err := net.IOCounters(true)
	if err == nil {
		for _, iface := range interfaces {
			// Skip uninteresting interfaces
			if iface.BytesSent == 0 && iface.BytesRecv == 0 {
				continue
			}

			s.Custom[fmt.Sprintf("net_%s_sent", iface.Name)] = utils.FormatNetworkBytes(iface.BytesSent)
			s.Custom[fmt.Sprintf("net_%s_recv", iface.Name)] = utils.FormatNetworkBytes(iface.BytesRecv)
		}
	}
}

// collectWindowsCPUData gathers CPU usage and frequency information
func collectWindowsCPUData(s *models.System) {
	cpuInfo, err := cpu.Info()
	if err == nil && len(cpuInfo) > 0 {
		s.Custom["cpu_core_count"] = strconv.Itoa(len(cpuInfo))
		if cpuInfo[0].Mhz > 0 {
			s.CPUMHZ = fmt.Sprintf("%.0f MHz", cpuInfo[0].Mhz)
		}
	}

	// Per-core CPU usage
	cpuPercentages, err := cpu.Percent(1*time.Second, true)
	if err == nil {
		s.CPUCores = make(map[string]string)
		for i, percentage := range cpuPercentages {
			s.CPUCores[fmt.Sprintf("cpu_core_%d", i)] = fmt.Sprintf("%.2f", percentage)
		}
	}

	// Overall CPU percentage
	totalCPU, err := cpu.Percent(0, false)
	if err == nil && len(totalCPU) > 0 {
		s.CPUUsage = fmt.Sprintf("%.2f%%", totalCPU[0])
	}
}

// collectWindowsDiskData gathers disk space information
func collectWindowsDiskData(s *models.System) {
	partitions, err := disk.Partitions(true)
	if err != nil {
		return
	}

	var totalDiskSpace, usedDiskSpace, freeDiskSpace uint64

	for _, partition := range partitions {
		// Skip non-fixed drives (like CD-ROMs, network drives)
		if partition.Fstype == "CDFS" || partition.Fstype == "" {
			continue
		}

		diskUsage, err := disk.Usage(partition.Mountpoint)
		if err != nil {
			continue
		}

		// Include significant partitions only (>100MB)
		if diskUsage.Total > 1024*1024*100 {
			totalDiskSpace += diskUsage.Total
			usedDiskSpace += diskUsage.Used
			freeDiskSpace += diskUsage.Free

			// Add details for each drive
			driveLetter := strings.TrimRight(partition.Mountpoint, ":\\")
			s.Custom[fmt.Sprintf("disk_%s_total", driveLetter)] = utils.FormatBytes(diskUsage.Total)
			s.Custom[fmt.Sprintf("disk_%s_used", driveLetter)] = utils.FormatBytes(diskUsage.Used)
			s.Custom[fmt.Sprintf("disk_%s_free", driveLetter)] = utils.FormatBytes(diskUsage.Free)
			s.Custom[fmt.Sprintf("disk_%s_percent", driveLetter)] = fmt.Sprintf("%.2f%%", diskUsage.UsedPercent)
		}
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

// collectWindowsSystemLoadData gathers load average information
// Note: Windows doesn't have load averages like Unix systems
func collectWindowsSystemLoadData(s *models.System) {
	// For Windows, we'll use CPU utilization as an approximation of system load
	cpuPercent, err := cpu.Percent(0, false)
	if err == nil && len(cpuPercent) > 0 {
		loadValue := cpuPercent[0] / 100.0 * float64(runtime.NumCPU())
		s.LoadAvg1 = fmt.Sprintf("%.2f", loadValue)
		s.LoadAvg5 = s.LoadAvg1  // Windows doesn't have 5 min average
		s.LoadAvg15 = s.LoadAvg1 // Windows doesn't have 15 min average :/
	}
}

// collectProcessData gathers information about running processes
func collectWindowsProcessData(s *models.System) {
	processes, err := process.Processes()
	if err == nil {
		s.ProcessCount = len(processes)
	}
}

// collectWindowsMemoryData gathers RAM usage information
func collectWindowsMemoryData(s *models.System) {
	memory, err := mem.VirtualMemory()
	if err == nil {
		s.TotalRAM = utils.FormatBytes(memory.Total)
		s.FreeRAM = utils.FormatBytes(memory.Free)
		s.UsedRAM = utils.FormatBytes(memory.Used)
		s.UsedRAMPercentage = fmt.Sprintf("%.2f%%", memory.UsedPercent)
	}
}

// collectWindowsSwapData gathers page file (swap) information
func collectWindowsSwapData(s *models.System) {
	swap, err := mem.SwapMemory()
	if err == nil {
		s.SwapUsed = utils.FormatBytes(swap.Used)
		s.SwapTotal = utils.FormatBytes(swap.Total)
		s.SwapPercent = fmt.Sprintf("%.2f%%", swap.UsedPercent)
	}
}

// collectWindowsHostData gathers host information like hostname and uptime
func collectWindowsHostData(s *models.System) {
	hostInfo, err := host.Info()
	if err == nil {
		s.Hostname = hostInfo.Hostname

		uptimeDuration := time.Duration(hostInfo.Uptime) * time.Second
		days := int(uptimeDuration.Hours()) / 24
		hours := int(uptimeDuration.Hours()) % 24
		minutes := int(uptimeDuration.Minutes()) % 60

		s.Uptime = fmt.Sprintf("%dd %dh %dm", days, hours, minutes)

		// Add OS version information
		s.Custom["os_version"] = fmt.Sprintf("Windows %s", hostInfo.PlatformVersion)
		s.Custom["windows_build"] = hostInfo.KernelVersion
	}
}
