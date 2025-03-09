package main

import (
	"device-chronicle-client/websocket"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

type Config struct {
	Server     string `json:"server"`
	ClientName string `json:"client_name"`
	Interval   int    `json:"interval"`
	DummyData  bool   `json:"dummy_data"`
}

func main() {
	// Define installation directories
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Failed to get user home directory: %v", err)
	}
	configDir := filepath.Join(homeDir, ".config", "chronicle-client")
	configFile := filepath.Join(configDir, "config.json")
	binDir := filepath.Join(homeDir, ".local", "bin")
	binPath := filepath.Join(binDir, "chronicle-client")

	// Define flags
	serverAddr := flag.String("server", "", "Server address, e.g. http://localhost:8000")
	dummyData := flag.Bool("dummy", false, "Use dummy data instead of real data for testing")
	interval := flag.Int("interval", 2, "Interval in seconds to send data to server")
	clientName := flag.String("client", "", "Client name")
	install := flag.Bool("install", false, "Install the client to user's home directory")
	flag.Parse()

	// Handle installation
	if *install {
		if *serverAddr == "" || *clientName == "" {
			log.Fatalln("Server address and client name are required for installation")
		}

		// Create configuration
		config := Config{
			Server:     *serverAddr,
			ClientName: *clientName,
			Interval:   *interval,
			DummyData:  *dummyData,
		}

		// Create directories
		os.MkdirAll(configDir, 0755)
		os.MkdirAll(binDir, 0755)

		// Save config
		configJSON, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			log.Fatalf("Failed to create config: %v", err)
		}
		if err := os.WriteFile(configFile, configJSON, 0644); err != nil {
			log.Fatalf("Failed to write config file: %v", err)
		}

		// Copy binary
		execPath, err := os.Executable()
		if err != nil {
			log.Fatalf("Failed to get executable path: %v", err)
		}
		execData, err := os.ReadFile(execPath)
		if err != nil {
			log.Fatalf("Failed to read executable: %v", err)
		}
		if err := os.WriteFile(binPath, execData, 0755); err != nil {
			log.Fatalf("Failed to copy executable: %v", err)
		}

		// Create user systemd service
		systemdDir := filepath.Join(homeDir, ".config", "systemd", "user")
		os.MkdirAll(systemdDir, 0755)
		serviceFile := filepath.Join(systemdDir, "chronicle-client.service")

		serviceContent := `[Unit]
                            Description=Chronicle Client Service
							After=network.target

							[Service]
 							ExecStart=%s
                            Restart=always
                            RestartSec=5

                            [Install]
                            WantedBy=default.target
                           `
		serviceContent = fmt.Sprintf(serviceContent, binPath)
		if err := os.WriteFile(serviceFile, []byte(serviceContent), 0644); err != nil {
			log.Fatalf("Failed to write service file: %v", err)
		}

		// Enable and start the service
		cmd := exec.Command("systemctl", "--user", "enable", "chronicle-client.service")
		cmd.Run()
		cmd = exec.Command("systemctl", "--user", "start", "chronicle-client.service")
		cmd.Run()

		fmt.Println("Installation completed!")
		fmt.Printf("- Binary installed: %s\n", binPath)
		fmt.Printf("- Config saved: %s\n", configFile)
		fmt.Println("- Service enabled and started")
		fmt.Println("You can now run 'chronicle-client' without parameters")
		os.Exit(0)
	}

	// Load config if exists and no command line args provided
	if (*serverAddr == "" || *clientName == "") && fileExists(configFile) {
		fmt.Println("Loading configuration from file...")
		configData, err := os.ReadFile(configFile)
		if err != nil {
			log.Fatalf("Failed to read config file: %v", err)
		}

		var config Config
		if err := json.Unmarshal(configData, &config); err != nil {
			log.Fatalf("Failed to parse config file: %v", err)
		}

		// Use config values if command line args aren't provided
		if *serverAddr == "" {
			*serverAddr = config.Server
		}
		if *clientName == "" {
			*clientName = config.ClientName
		}
		if flag.Lookup("interval").DefValue == fmt.Sprint(*interval) {
			*interval = config.Interval
		}
		if flag.Lookup("dummy").DefValue == fmt.Sprint(*dummyData) {
			*dummyData = config.DummyData
		}
	}

	// Validate required parameters
	if *serverAddr == "" {
		log.Fatalln("Server address is required. Usage: ./chronicle-client --server http://localhost:8000")
	}

	if *clientName == "" {
		log.Fatalln("Client name is required. Usage: ./chronicle-client --client desktop")
	}

	// Run the client
	websocket.Websocket(serverAddr, dummyData, interval, clientName)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
