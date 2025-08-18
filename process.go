package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/process"
)

// RenderProcesses renders a list of processes into a formatted table
func RenderProcesses(processes []ProcessInfo) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleRounded)
	t.AppendHeader(table.Row{
		"PID", "Name", "Port", "Protocol", "Status", "CPU", "Memory", "User",
	})

	for _, p := range processes {
		t.AppendRow(table.Row{
			p.PID,
			p.Name,
			p.Port,
			p.Protocol,
			p.Status,
			FormatCPU(p.CPUPercent),
			FormatMemory(p.MemoryMB),
			p.Username,
		})
	}

	t.Render()
}

// ProcessInfo represents detailed information about a process using a port
type ProcessInfo struct {
	PID        int32   `json:"pid"`
	Name       string  `json:"name"`
	Port       uint32  `json:"port"`
	Protocol   string  `json:"protocol"`
	Status     string  `json:"status"`
	LocalAddr  string  `json:"local_addr"`
	RemoteAddr string  `json:"remote_addr"`
	CPUPercent float64 `json:"cpu_percent"`
	MemoryMB   float64 `json:"memory_mb"`
	MemPercent float64 `json:"mem_percent"`
	CreateTime int64   `json:"create_time"`
	Username   string  `json:"username"`
	Selected   bool    `json:"selected"`
}

// GetProcessInfo gathers detailed information about all processes using TCP/UDP ports
func GetProcessInfo() ([]ProcessInfo, error) {
	connections, err := net.Connections("all")
	if err != nil {
		return nil, fmt.Errorf("failed to get connections: %w", err)
	}

	var processes []ProcessInfo
	processMap := make(map[int32]*process.Process) // Cache processes to avoid duplicate calls

	for _, conn := range connections {
		// Skip non-TCP/UDP connections
		if conn.Type != 1 && conn.Type != 2 { // 1=TCP, 2=UDP
			continue
		}

		// Skip connections without ports
		if conn.Laddr.Port == 0 {
			continue
		}

		// Skip non-listening and non-established connections
		if conn.Status != "LISTEN" && conn.Status != "ESTABLISHED" {
			continue
		}

		// Skip connections without PID
		if conn.Pid == 0 {
			continue
		}

		var proc *process.Process
		var ok bool

		// Check if we already have this process cached
		if proc, ok = processMap[conn.Pid]; !ok {
			proc, err = process.NewProcess(conn.Pid)
			if err != nil {
				continue // Skip if we can't get process info
			}
			processMap[conn.Pid] = proc
		}

		processInfo := ProcessInfo{
			PID:      conn.Pid,
			Port:     conn.Laddr.Port,
			Status:   conn.Status,
			Selected: false,
		}

		// Set protocol
		switch conn.Type {
		case 1:
			processInfo.Protocol = "TCP"
		case 2:
			processInfo.Protocol = "UDP"
		default:
			processInfo.Protocol = "UNK"
		}

		// Set addresses
		processInfo.LocalAddr = conn.Laddr.IP + ":" + strconv.Itoa(int(conn.Laddr.Port))
		if conn.Raddr.IP != "" {
			processInfo.RemoteAddr = conn.Raddr.IP + ":" + strconv.Itoa(int(conn.Raddr.Port))
		} else {
			processInfo.RemoteAddr = "-"
		}

		// Get process name
		if name, err := proc.Name(); err == nil {
			processInfo.Name = name
		} else {
			processInfo.Name = "unknown"
		}

		// Get CPU usage
		if cpuPercent, err := proc.CPUPercent(); err == nil {
			processInfo.CPUPercent = cpuPercent
		}

		// Get memory info
		if memInfo, err := proc.MemoryInfo(); err == nil {
			processInfo.MemoryMB = float64(memInfo.RSS) / 1024 / 1024 // Convert bytes to MB
		}

		// Get memory percentage
		if memPercent, err := proc.MemoryPercent(); err == nil {
			processInfo.MemPercent = float64(memPercent)
		}

		// Get create time
		if createTime, err := proc.CreateTime(); err == nil {
			processInfo.CreateTime = createTime
		}

		// Get username
		if username, err := proc.Username(); err == nil {
			processInfo.Username = username
		} else {
			processInfo.Username = "unknown"
		}

		processes = append(processes, processInfo)
	}

	return processes, nil
}

// RefreshProcessInfo updates CPU usage for existing processes
func RefreshProcessInfo(processes []ProcessInfo) []ProcessInfo {
	for i := range processes {
		if proc, err := process.NewProcess(processes[i].PID); err == nil {
			// Get fresh CPU usage
			if cpuPercent, err := proc.CPUPercent(); err == nil {
				processes[i].CPUPercent = cpuPercent
			}

			// Get fresh memory info
			if memInfo, err := proc.MemoryInfo(); err == nil {
				processes[i].MemoryMB = float64(memInfo.RSS) / 1024 / 1024
			}

			// Get fresh memory percentage
			if memPercent, err := proc.MemoryPercent(); err == nil {
				processes[i].MemPercent = float64(memPercent)
			}
		}
	}
	return processes
}

// KillProcess terminates a process by PID
func KillProcess(pid int32) error {
	proc, err := process.NewProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to find process %d: %w", pid, err)
	}

	name, _ := proc.Name()

	// Try graceful termination first
	if err := proc.Terminate(); err != nil {
		// If graceful termination fails, force kill
		if killErr := proc.Kill(); killErr != nil {
			return fmt.Errorf("failed to kill process %d (%s): %w", pid, name, killErr)
		}
	}

	// Wait a bit to ensure process is terminated
	time.Sleep(100 * time.Millisecond)

	return nil
}

// KillSelectedProcesses kills all selected processes
func KillSelectedProcesses(processes []ProcessInfo) ([]string, []string) {
	var killed []string
	var failed []string

	for _, proc := range processes {
		if proc.Selected {
			if err := KillProcess(proc.PID); err != nil {
				failed = append(failed, fmt.Sprintf("%s (PID: %d): %v", proc.Name, proc.PID, err))
			} else {
				killed = append(killed, fmt.Sprintf("%s (PID: %d)", proc.Name, proc.PID))
			}
		}
	}

	return killed, failed
}

// FilterProcesses filters processes based on various criteria
func FilterProcesses(processes []ProcessInfo, filters ProcessFilters) []ProcessInfo {
	var filtered []ProcessInfo

	for _, proc := range processes {
		// Filter by port
		if filters.Port != 0 && proc.Port != filters.Port {
			continue
		}

		// Filter by process name
		if filters.ProcessName != "" && !strings.Contains(strings.ToLower(proc.Name), strings.ToLower(filters.ProcessName)) {
			continue
		}

		// Filter by protocol
		if filters.Protocol != "" && !strings.EqualFold(proc.Protocol, filters.Protocol) {
			continue
		}

		// Filter by status
		if filters.Status != "" && !strings.EqualFold(proc.Status, filters.Status) {
			continue
		}

		// Filter by listening only
		if filters.ListenOnly && proc.Status != "LISTEN" {
			continue
		}

		filtered = append(filtered, proc)
	}

	return filtered
}

// ProcessFilters represents filtering criteria for processes
type ProcessFilters struct {
	Port        uint32
	ProcessName string
	Protocol    string
	Status      string
	ListenOnly  bool
}

// FormatCPU formats CPU percentage for display
func FormatCPU(cpu float64) string {
	if cpu < 0.1 {
		return "0.0%"
	}
	return fmt.Sprintf("%.1f%%", cpu)
}

// FormatMemory formats memory usage for display
func FormatMemory(memMB float64) string {
	if memMB < 1 {
		return fmt.Sprintf("%.1fKB", memMB*1024)
	} else if memMB > 1024 {
		return fmt.Sprintf("%.1fGB", memMB/1024)
	}
	return fmt.Sprintf("%.1fMB", memMB)
}

// FormatMemoryPercent formats memory percentage for display
func FormatMemoryPercent(percent float64) string {
	if percent < 0.1 {
		return "0.0%"
	}
	return fmt.Sprintf("%.1f%%", percent)
}

// GetUniqueProcesses removes duplicate processes (same PID) and keeps the most relevant one
func GetUniqueProcesses(processes []ProcessInfo) []ProcessInfo {
	seen := make(map[int32]ProcessInfo)

	for _, proc := range processes {
		existing, exists := seen[proc.PID]
		if !exists {
			seen[proc.PID] = proc
			continue
		}

		// Prefer LISTEN status over ESTABLISHED
		if proc.Status == "LISTEN" && existing.Status == "ESTABLISHED" {
			seen[proc.PID] = proc
		}
	}

	var unique []ProcessInfo
	for _, proc := range seen {
		unique = append(unique, proc)
	}

	return unique
}
