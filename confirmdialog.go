package main

import (
	"strconv"
	"strings"
)

func displayName(name string) string {
	if name == "" {
		return "process"
	}
	return name
}

func renderConfirmBox(target Process) string {
	lines := []string{
		"Are you sure you want to kill?",
		"PID: " + strconv.Itoa(target.PID),
	}
	if target.Name != "" {
		lines = append(lines, "Process: "+target.Name)
	}
	return confirmBoxStyle.Render(strings.Join(lines, "\n"))
}
