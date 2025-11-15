package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/heartwilltell/scotty"
)

const (
	cmdName  = "portman"
	cmdShort = "Portman - Port Usage Analyzer"
	cmdLong  = "Portman is a fast and colorful command-line tool to discover which processes are using specific ports on your system."
	version  = "0.1.0"
)

func main() {
	var (
		filterPort     uint
		filterProcess  string
		showListenOnly bool
		hideBorders    bool
	)

	cmd := scotty.Command{
		Name:  cmdName,
		Short: cmdShort,
		Long:  cmdLong,
		SetFlags: func(flags *scotty.FlagSet) {
			flags.UintVar(&filterPort, "port", 0, "Filter by specific port number")
			flags.StringVar(&filterProcess, "process", "", "Filter by process name (case-insensitive partial match)")
			flags.BoolVar(&showListenOnly, "listen", false, "Show only listening ports")
			flags.BoolVar(&hideBorders, "no-borders", false, "Hide table borders for cleaner output")
		},

		Run: func(cmd *scotty.Command, args []string) error {
			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer stop()

			processManager, err := NewProcessManager(ctx)
			if err != nil {
				return fmt.Errorf("new process manager: %w", err)
			}

			m := newTableModel(processManager)

			p := tea.NewProgram(m,
				tea.WithOutput(os.Stdout),
				tea.WithContext(ctx),
				tea.WithAltScreen(),
				tea.WithMouseCellMotion(),
			)

			if _, err := p.Run(); err != nil {
				return fmt.Errorf("run program: %w", err)
			}

			return nil
		},
	}

	if err := cmd.Exec(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
