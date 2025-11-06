package main

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	md "github.com/nao1215/markdown"
)

func RenderMarkdownTableForProcesses(processes []Process, hideBorders bool) (string, error) {
	if hideBorders {
		return renderPlainTextTable(processes), nil
	}

	var byf bytes.Buffer
	doc := md.NewMarkdown(&byf)
	table := md.TableSet{
		Header: []string{"PID", "Process", "Port", "Protocol", "Status", "Local Address", "Remote Address"},
		Rows:   make([][]string, 0, len(processes)),
	}

	for _, process := range processes {
		table.Rows = append(table.Rows, []string{
			strconv.Itoa(process.PID),
			process.Name,
			strconv.Itoa(process.Port),
			process.Protocol,
			process.Status,
			process.LocalAddr,
		})
	}

	if err := doc.Table(table).Build(); err != nil {
		return "", fmt.Errorf("build table: %w", err)
	}

	return byf.String(), nil
}

// renderPlainTextTable creates a clean table without borders
func renderPlainTextTable(processes []Process) string {
	if len(processes) == 0 {
		return "No processes found.\n"
	}

	headers := []string{"PID", "PROCESS", "PORT", "PROTOCOL", "STATUS", "LOCAL ADDRESS", "REMOTE ADDRESS"}

	// Collect all data including headers
	var allRows [][]string
	allRows = append(allRows, headers)

	for _, process := range processes {
		allRows = append(allRows, []string{
			strconv.Itoa(process.PID),
			process.Name,
			strconv.Itoa(process.Port),
			process.Protocol,
			process.Status,
			process.LocalAddr,
		})
	}

	// Calculate column widths
	colWidths := make([]int, len(headers))
	for _, row := range allRows {
		for i, cell := range row {
			if len(cell) > colWidths[i] {
				colWidths[i] = len(cell)
			}
		}
	}

	// Set a maximum width for the Status column (index 4)
	// The longest status value is "LISTEN" or "ACTIVE" or "CLOSED" (6 chars)
	if len(colWidths) > 4 && colWidths[4] > 6 {
		colWidths[4] = 6
	}

	var result strings.Builder

	// Render each row with proper spacing
	for i, row := range allRows {
		for j, cell := range row {
			// Left-align all columns except add padding
			result.WriteString(fmt.Sprintf("%-*s", colWidths[j], cell))
			if j < len(row)-1 {
				result.WriteString("   ") // 3 spaces between columns
			}
		}
		result.WriteString("\n")

		// Add separator after header
		if i == 0 {
			for j := range row {
				result.WriteString(strings.Repeat("-", colWidths[j]))
				if j < len(row)-1 {
					result.WriteString("   ")
				}
			}
			result.WriteString("\n")
		}
	}

	return result.String()
}
