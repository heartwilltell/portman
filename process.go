package main

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	netutil "github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/process"
)

// Error represents a set of errors related to processes and connection management.
// Implements the error interface.
type Error string

// Error returns the error message.
func (e Error) Error() string { return string(e) }

const (
	// ErrNoConnectionsFound indicates that no connections were found.
	ErrNoConnectionsFound = Error("no connections found")
)

// Enumerated constants for protocol types and status.
const (
	ProtocolTCP  = "TCP"
	ProtocolUDP  = "UDP"
	ProtocolAll  = "all"
	ProtocolTCP4 = "tcp4"
	ProtocolTCP6 = "tcp6"
	ProtocolUDP4 = "udp4"
	ProtocolUDP6 = "udp6"

	StatusActive = "ACTIVE"
	StatusListen = "LISTEN"
	StatusClosed = "CLOSED"
)

// Process represents a process that is using a port.
type Process struct {
	PID        int
	Name       string
	Port       int
	Protocol   string
	Status     string
	LocalAddr  string
	RemoteAddr string
	Cmdline    string
}

// Options represents the options for the GetOcupiedPorts function.
type Options struct {
	FilterPort     uint
	FilterProcess  string
	FilterProtocol string
	ShowListenOnly bool
}

// Option represents an option for the GetOcupiedPorts function.
type Option func(*Options)

// WithFilterPort returns an option that filters the processes by port.
func WithFilterPort(port uint) Option {
	return func(o *Options) { o.FilterPort = port }
}

// WithFilterProcess returns an option that filters the processes by process name.
func WithFilterProcess(process string) Option {
	return func(o *Options) { o.FilterProcess = process }
}

// WithShowListenOnly returns an option that shows only listening ports.
func WithShowListenOnly(show bool) Option {
	return func(o *Options) { o.ShowListenOnly = show }
}

// WithFilterProtocol returns an option that filters the processes by protocol.
func WithFilterProtocol(protocol string) Option {
	return func(o *Options) { o.FilterProtocol = protocol }
}

// ProcessManager is a manager for processes.
type ProcessManager struct {
	mu        sync.RWMutex
	pidIndex  map[int]int
	processes []Process
	cancel    context.CancelFunc
	ticker    *time.Ticker
	err       error
}

// NewProcessManager creates a new ProcessManager.
func NewProcessManager(ctx context.Context) (*ProcessManager, error) {
	ctx, cancel := context.WithCancel(ctx)

	manager := &ProcessManager{
		pidIndex:  make(map[int]int),
		processes: make([]Process, 0),
		cancel:    cancel,
		ticker:    time.NewTicker(time.Second * 5),
	}

	// Fetch initial data immediately and wait for it to complete.
	if err := manager.fetchProcesses(ctx, WithFilterProtocol("all")); err != nil {
		manager.mu.Lock()
		manager.err = fmt.Errorf("initial fetch processes: %w", err)
		manager.mu.Unlock()
	}

	// Start the background monitoring.
	go manager.monitorProcesses(ctx)

	return manager, nil
}

func (m *ProcessManager) Stop() {
	m.cancel()
	m.ticker.Stop()
}

func (m *ProcessManager) Processes(ctx context.Context, options ...Option) ([]Process, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	listOptions, err := parseOptions(options...)
	if err != nil {
		return nil, fmt.Errorf("parse options: %w", err)
	}

	filtered := make([]Process, 0, len(m.processes))

	for _, process := range m.processes {
		if listOptions.FilterProtocol != ProtocolAll && process.Protocol != listOptions.FilterProtocol {
			continue
		}

		if listOptions.FilterPort != 0 && uint(process.Port) != listOptions.FilterPort {
			continue
		}

		if listOptions.ShowListenOnly && process.Status != StatusListen {
			continue
		}

		if listOptions.FilterProcess != "" && !strings.Contains(
			strings.ToLower(process.Name),
			strings.ToLower(listOptions.FilterProcess),
		) {
			continue
		}

		filtered = append(filtered, process)
	}

	return filtered, nil
}

func (m *ProcessManager) monitorProcesses(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return

		case <-m.ticker.C:
			if err := m.fetchProcesses(ctx, WithFilterProtocol("all")); err != nil {
				m.mu.Lock()
				m.err = fmt.Errorf("fetch processes: %w", err)
				m.mu.Unlock()
				continue
			}
		}
	}
}

func (m *ProcessManager) fetchProcesses(ctx context.Context, options ...Option) error {
	listOptions, err := parseOptions(options...)
	if err != nil {
		return fmt.Errorf("parse options: %w", err)
	}

	connections, err := m.connections(ctx, listOptions)
	if err != nil {
		return fmt.Errorf("get connections: %w", err)
	}

	if len(connections) == 0 {
		return ErrNoConnectionsFound
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.processes == nil {
		m.processes = make([]Process, 0, len(connections))
	}

	if m.pidIndex == nil {
		m.pidIndex = make(map[int]int, len(connections))
	}

	// Only clear if we have new data to replace it with.
	if len(connections) > 0 {
		clear(m.pidIndex)
		clear(m.processes)
	}

	processes := make([]Process, 0, len(connections))

	for _, conn := range connections {
		select {
		case <-ctx.Done():
			return ctx.Err()

		default:
			proc, err := process.NewProcessWithContext(ctx, int32(conn.Pid))
			if err != nil {
				// Skip process that we can't get.
				continue
			}

			name, nameErr := proc.Name()
			if nameErr != nil {
				// Skip process if we can't get the name.
				continue
			}

			var (
				protocol string
				status   = conn.Status
			)

			switch conn.Family {
			case 2: // AF_INET.
				switch conn.Type {
				case 1: // SOCK_STREAM.
					protocol = ProtocolTCP

				case 2: // SOCK_DGRAM.
					protocol = ProtocolUDP
					if conn.Status == "" {
						status = StatusActive
					}

				default:
					continue
				}

			case 10: // AF_INET6.
				switch conn.Type {
				case 1: // SOCK_STREAM.
					protocol = ProtocolTCP6

				case 2: // SOCK_DGRAM.
					protocol = ProtocolUDP6
					if conn.Status == "" {
						status = StatusActive
					}

				default:
					continue
				}

			default:
				continue
			}

			if listOptions.FilterProcess != "" && !strings.Contains(
				strings.ToLower(name),
				strings.ToLower(listOptions.FilterProcess),
			) {
				continue
			}

			if listOptions.FilterPort != 0 && uint(conn.Laddr.Port) != listOptions.FilterPort {
				continue
			}

			if listOptions.ShowListenOnly && status != StatusListen {
				continue
			}

			process := Process{
				PID:        int(conn.Pid),
				Name:       name,
				Port:       int(conn.Laddr.Port),
				Protocol:   protocol,
				Status:     status,
				LocalAddr:  fmt.Sprintf("%s:%d", conn.Laddr.IP, conn.Laddr.Port),
				RemoteAddr: fmt.Sprintf("%s:%d", conn.Raddr.IP, conn.Raddr.Port),
			}

			processes = append(processes, process)
		}
	}

	m.processes = processes
	clear(m.pidIndex)

	for i, process := range processes {
		m.pidIndex[process.PID] = i
	}

	return nil
}

func (m *ProcessManager) connections(ctx context.Context, options Options) ([]netutil.ConnectionStat, error) {
	connections, err := netutil.ConnectionsWithContext(ctx, options.FilterProtocol)
	if err != nil {
		return nil, fmt.Errorf("get tcp connections: %w", err)
	}

	return connections, nil
}

func parseOptions(options ...Option) (Options, error) {
	listOptions := Options{
		FilterProtocol: "all",
		ShowListenOnly: false,
	}

	for _, option := range options {
		option(&listOptions)
	}

	switch listOptions.FilterProtocol {
	case "tcp", "udp", "all", "tcp4", "tcp6", "udp4", "udp6":
	default:
		return listOptions, fmt.Errorf("invalid protocol: %s", listOptions.FilterProtocol)
	}

	return listOptions, nil
}
