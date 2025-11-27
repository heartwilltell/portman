# Portman - Port Usage Analyzer

A fast and colorful command-line tool to discover which processes are using specific ports on your system.

## Features

- 🔍 **Port Discovery**: Quickly find which processes are using specific ports
- 🎨 **Colorful Output**: Easy-to-read colored terminal output
- 🖥️ **Interactive TUI**: Beautiful table interface with process management
- 🔧 **Flexible Filtering**: Filter by port number, process name, or connection status
- 📊 **Detailed Information**: Shows port, protocol type, status, PID, process name, addresses, CPU and memory usage
- ⚡ **Process Management**: Kill processes directly from the interface
- 📈 **Real-time Monitoring**: Live CPU and memory usage updates
- 🚀 **Fast**: Built in Go for optimal performance
- 🌐 **Cross-Platform**: Works on macOS, Linux, and Windows

## Installation

### Download Pre-built Binary

Download the latest release from the [releases page](https://github.com/heartwilltell/portman/releases).

### Install via Go

```bash
go install github.com/heartwilltell/portman@latest
```

### Build from Source

```bash
git clone https://github.com/heartwilltell/portman.git
cd portman
go build -o portman .
```

## Usage

### Basic Usage

Show all busy ports and their processes:

```bash
portman
```

### Command Line Options

```
portman - Port Usage Analyzer

Usage:
  portman <flags> [arguments...]

Flags:
  -listen bool     Show only listening ports
  -port uint       Filter by specific port number
  -process string  Filter by process name (case-insensitive partial match)
  -tui bool        Launch interactive TUI mode with process management
  -version bool    Show version information
```

### Examples

#### Launch interactive TUI mode

```bash
portman
```

#### Show only listening ports

```bash
portman -listen
```

#### Find who's using port 8080

```bash
portman -port 8080
```

#### Find all Chrome processes using ports

```bash
portman -process chrome
```

#### Find listening ports used by Node.js

```bash
portman -listen -process node
```

### TUI Features

- **📋 Interactive Table**: Navigate through processes with arrow keys
- **✅ Process Selection**: Use `Space` to select/deselect processes
- **🔥 Kill Processes**: Press `k` to terminate selected processes
- **📊 Real-time Stats**: Live CPU and memory usage monitoring
- **🔍 Quick Actions**: Sort by CPU usage, select all/none
- **🎯 Visual Indicators**: Color-coded resource usage and status

### TUI Keybindings

| Key            | Action                   |
| -------------- | ------------------------ |
| `↑/↓` or `j/k` | Navigate table           |
| `Space`        | Toggle process selection |
| `k`            | Kill selected processes  |
| `a`            | Select all processes     |
| `n`            | Select none              |
| `s`            | Sort by CPU usage        |
| `r`            | Refresh process list     |
| `?/h`          | Toggle help              |
| `q`            | Quit                     |

## Sample Output

### CLI Mode

```text
Scanning for busy ports and their processes...

PORT     TYPE   STATUS      PID          PROCESS              ADDRESS
────────────────────────────────────────────────────────────────────────────
80       TCP    LISTEN      702          nginx                *:80
443      TCP    LISTEN      702          nginx                *:443
3000     TCP    LISTEN      1234         node                 127.0.0.1:3000
3000     TCP    ESTABLISHED 1234         node                 127.0.0.1:3000 -> 127.0.0.1:54321
5432     TCP    LISTEN      567          postgres             127.0.0.1:5432
8080     TCP    LISTEN      890          java                 *:8080
```

### TUI Mode

```
┌─ Portman - Port Usage Analyzer | Processes: 12 | Selected: 2 ──────────────┐
│                                                                            │
│ ✓  PID    Process          Port  Proto Status     CPU    Memory   Local   │
│ ─  ────   ─────────────────────  ───── ────────── ────── ──────── ─────── │
│ ✓  1234   node             3000  TCP   LISTEN     15.2%  120.5MB  *:3000  │
│    702    nginx             80   TCP   LISTEN      2.1%   45.2MB  *:80     │
│ ✓  890    java             8080  TCP   LISTEN     45.7%  512.1MB  *:8080  │
│    567    postgres         5432  TCP   LISTEN      8.3%  256.7MB  *:5432  │
│                                                                            │
│ Space: Select | k: Kill | a: Select All | r: Refresh | ?: Help           │
└────────────────────────────────────────────────────────────────────────────┘
```

### Color Coding

- **Red ports**: System/privileged ports (< 1024)
- **Cyan ports**: User ports (>= 1024)
- **Blue**: TCP connections
- **Purple**: UDP connections
- **Green**: LISTEN status
- **Yellow**: ESTABLISHED status

## Building

The project includes a build script that supports multiple platforms:

### Build for current platform

```bash
./build.sh
```

### Build for all platforms

```bash
./build.sh all
```

### Install locally

```bash
./build.sh install
```

### Other build commands

```bash
./build.sh clean    # Clean build artifacts
./build.sh test     # Run tests
./build.sh fmt      # Format code
./build.sh help     # Show help
```

## How It Works

`portman` uses advanced system analysis to gather comprehensive port and process information:

1. **System Connection Analysis**: Uses the `gopsutil` library to gather detailed connection information
2. **Process Monitoring**: Collects real-time CPU and memory statistics for each process
3. **Interactive Management**: Provides safe process termination capabilities
4. **Real-time Updates**: Continuously refreshes data for live monitoring

The tool focuses specifically on TCP and UDP connections with active ports, providing both quick CLI access and comprehensive TUI management.

## Requirements

- Go 1.21 or later (for building from source)
- Appropriate system permissions to read process information

## Platform Support

- **macOS**: Full support
- **Linux**: Full support
- **Windows**: Full support

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Dependencies

- [scotty](https://github.com/heartwilltell/scotty) - Command-line interface framework
- [gopsutil](https://github.com/shirou/gopsutil) - System and process information library
- [bubbletea](https://github.com/charmbracelet/bubbletea) - Terminal user interface framework
- [lipgloss](https://github.com/charmbracelet/lipgloss) - Style definitions for nice terminal layouts
- [bubble-table](https://github.com/evertras/bubble-table) - Table component for Bubble Tea

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Author

Created by [heartwilltell](https://github.com/heartwilltell)

## Acknowledgments

- Thanks to the `gopsutil` team for the excellent system information library
- Inspired by various network diagnostic tools like `netstat`, `lsof`, and `ss`
