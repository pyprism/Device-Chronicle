# Device Chronicle [![Client](https://github.com/pyprism/Device-Chronicle/actions/workflows/client.yaml/badge.svg)](https://github.com/pyprism/Device-Chronicle/actions/workflows/client.yaml) [![Server](https://github.com/pyprism/Device-Chronicle/actions/workflows/server.yaml/badge.svg)](https://github.com/pyprism/Device-Chronicle/actions/workflows/server.yaml)
Device Chronicle is a lightweight system monitoring tool that collects and visualizes realtime performance metrics from your devices. It consists of a server component (web dashboard) and client agents that run on monitored systems.

## Features

- **Realtime monitoring** via WebSocket communication
- **Performance visualization** with interactive charts:
    - CPU usage and temperature
    - Memory and swap usage
    - Network traffic
    - Disk space and usage
- **Multi-device support** - monitor multiple systems from a single dashboard
- **User-level installation** - no root privileges required
- **Automatic startup** via systemd user service

## Quick Start

### Server Setup

1. Start the server:
   ```bash
   # run this commands from the server directory
   cp .env.example .env
   # change the .env file to your needs
   docker compose -f docker-compose.production.yaml build
   docker compose -f docker-compose.production.yaml up -d
   ```
   
   ###### Note: You can also configure nginx to serve the dashboard on a custom domain

2. Access the dashboard at http://localhost:8000

### Client Setup

1. Install the client on each system you want to monitor:
   ```bash
   chmod +x ./chronicle-client
   ./chronicle-client --install --server http://SERVER_IP:8000 --client DEVICE_NAME
   ```
    ###### Note:
     - Download the client binary from the [releases page](https://github.com/pyprism/Device-Chronicle/releases)
     - Replace `SERVER_IP` with the IP address of the server or domain name and `DEVICE_NAME` with a unique name for the device
     - In case for updating the client, stop the service first `systemctl --user stop chronicle-client`


2. The client will automatically start and connect to the server

## Client Installation Options

```
Usage: chronicle-client [options]

Options:
  --install         Install the client to user's home directory
  --server string   Server address, e.g. http://localhost:8000
  --client string   Client name (e.g., desktop, laptop, server)
  --interval int    Data collection interval in seconds (default: 2)
  --dummy           Use dummy data for testing
```

## How It Works

1. The client collects system metrics using the gopsutil library
2. Data is sent to the server via WebSocket connection
3. The server processes and visualizes the data using interactive charts
4. All charts update in real-time as data arrives

## Configuration

After installation, configuration is stored in:
```
~/.config/chronicle-client/config.json
```

The client binary is installed to:
```
~/.local/bin/chronicle-client
```

## System Service Management

The client runs as a systemd user service that starts automatically on login:

```bash
# Check status
systemctl --user status chronicle-client

# Stop service
systemctl --user stop chronicle-client

# Start service
systemctl --user start chronicle-client
```

## Why Device Chronicle Exists

Device Chronicle was born out of a practical need, monitoring system resources on my linux gaming (aka potato) PC from another device. When playing games on my less powerful "potato PC," I wanted to keep an eye on system temperatures, CPU usage, and memory consumption without tabbing out of the game or impacting performance.

By creating a lightweight monitoring tool with a web interface accessible from any device, I could:

1. Monitor my Linux system resources in realtime from my phone or tablet
2. Track performance metrics during gaming sessions
3. Keep an eye on temperature levels during resource intensive tasks
4. Identify potential bottlenecks affecting game performance

## TODO
 - [ ] Code cleanup and more tests
 - [ ] Add more system metrics
 - [ ] Windows support
 - [ ] Maybe add a database to store historical data

## Status
 Beta
## Screenshot
<img src="/screenshot.png" width="80%" alt="ugly screenshot">

## Development

This project uses:
- Go for the server and client services
- JavaScript with ECharts for visualization
- WebSockets for realtime communication

## License

MIT

- Icon credit: [Icon by Iconic Panda](https://www.freepik.com/icon/torah_11288355)