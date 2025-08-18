package main

import (
	"fmt"
	"os"

	"github.com/heartwilltell/scotty"
)

func main() {
	var filterPort uint
	var filterProcess string
	var showListenOnly bool
	var showVersion bool
	var tuiMode bool

	cmd := scotty.Command{
		Name:  "wutp",
		Short: "Who Use This Port - A fast port usage analyzer",
		Long: `Who Use This Port (wutp) is a fast and colorful command-line tool to discover
which processes are using specific ports on your system. It provides detailed information
about port usage including process names, PIDs, connection types, and addresses.

Use --tui flag for an interactive table interface with process management capabilities.`,
		SetFlags: func(flags *scotty.FlagSet) {
			flags.UintVarE(&filterPort, "port", "", 0, "Filter by specific port number")
			flags.StringVarE(&filterProcess, "process", "", "", "Filter by process name (case-insensitive partial match)")
			flags.BoolVarE(&showListenOnly, "listen", "", false, "Show only listening ports")
			flags.BoolVarE(&showVersion, "version", "", false, "Show version information")
			flags.BoolVarE(&tuiMode, "tui", "", false, "Launch interactive TUI mode with process management")
		},

		Run: func(cmd *scotty.Command, args []string) error {
			processes, err := GetProcessInfo()
			if err != nil {
				return err
			}

			RenderProcesses(processes)

			return nil
		},
	}

	if err := cmd.Exec(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
