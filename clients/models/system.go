package models

type System struct {
	// Fixed fields
	PacketsSent        string `json:"packets_sent"`
	PacketsReceive     string `json:"packets_receive"`
	AverageChipsetTemp string `json:"average_chipset_temp"`
	CPUTemp            string `json:"cpu_temp"`
	TotalRAM           string `json:"total_ram"`
	FreeRAM            string `json:"free_ram"`
	UsedRAM            string `json:"used_ram"`
	UsedRAMPercentage  string `json:"used_ram_percentage"`
	Hostname           string `json:"hostname"`
	Uptime             string `json:"uptime"`
	LoadAvg1           string `json:"load_1"`
	LoadAvg5           string `json:"load_5"`
	LoadAvg15          string `json:"load_15"`
	ProcessCount       int    `json:"process_count"`
	CPUUsage           string `json:"cpu_usage"`
	CPUMHZ             string `json:"cpu_mhz"`
	DiskTotal          string `json:"disk_total"`
	DiskFree           string `json:"disk_free"`
	DiskUsed           string `json:"disk_used"`
	DiskUsagePercent   string `json:"disk_usage_percent"`
	SwapUsed           string `json:"swap_used"`
	SwapTotal          string `json:"swap_total"`
	SwapPercent        string `json:"swap_percent"`

	// Dynamic fields
	CPUCores map[string]string      `json:"cpu_cores"`
	Custom   map[string]interface{} `json:"custom,omitempty"`
}

// NewSystem creates a new System with initialized maps
func NewSystem() *System {
	return &System{
		CPUCores: make(map[string]string),
		Custom:   make(map[string]interface{}),
	}
}

func (s *System) ToMap() map[string]interface{} {
	result := make(map[string]interface{})

	// fixed fields
	result["packets_sent"] = s.PacketsSent
	result["packets_receive"] = s.PacketsReceive
	result["average_chipset_temp"] = s.AverageChipsetTemp
	result["cpu_temp"] = s.CPUTemp
	result["total_ram"] = s.TotalRAM
	result["free_ram"] = s.FreeRAM
	result["used_ram"] = s.UsedRAM
	result["used_ram_percentage"] = s.UsedRAMPercentage
	result["hostname"] = s.Hostname
	result["uptime"] = s.Uptime
	result["load_1"] = s.LoadAvg1
	result["load_5"] = s.LoadAvg5
	result["load_15"] = s.LoadAvg15
	result["process_count"] = s.ProcessCount
	result["cpu_usage"] = s.CPUUsage
	result["cpu_mhz"] = s.CPUMHZ
	result["disk_total"] = s.DiskTotal
	result["disk_free"] = s.DiskFree
	result["disk_used"] = s.DiskUsed
	result["disk_usage_percent"] = s.DiskUsagePercent
	result["swap_used"] = s.SwapUsed
	result["swap_total"] = s.SwapTotal
	result["swap_percent"] = s.SwapPercent

	for k, v := range s.CPUCores {
		result[k] = v
	}

	// custom fields
	for k, v := range s.Custom {
		result[k] = v
	}

	return result
}
