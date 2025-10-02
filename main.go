package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// Shortcut represents a single tmux shortcut.
type Shortcut struct {
	Key         string
	Description string
	Category    string
}

// getStaticShortcuts returns a hardcoded list of tmux shortcuts.
func getStaticShortcuts() []Shortcut {
	return []Shortcut{
		// Sessions
		{"Prefix d", "Detach from session", "Sessions"},
		{"Prefix s", "List sessions", "Sessions"},
		{"Prefix $", "Rename session", "Sessions"},

		// Windows
		{"Prefix c", "Create new window", "Windows"},
		{"Prefix ,", "Rename window", "Windows"},
		{"Prefix w", "List windows", "Windows"},
		{"Prefix f", "Find window", "Windows"},
		{"Prefix &", "Kill window", "Windows"},
		{"Prefix .", "Move window", "Windows"},
		{"Prefix p", "Previous window", "Windows"},
		{"Prefix n", "Next window", "Windows"},
		{"Prefix 0-9", "Select window by number", "Windows"},
		{"Prefix l", "Last selected window", "Windows"},

		// Pane Splitting & Nav
		{"Prefix %", "Split horizontally", "Pane Splitting & Nav"},
		{"Prefix \"", "Split vertically", "Pane Splitting & Nav"},
		{"Prefix o", "Next pane", "Pane Splitting & Nav"},
		{"Prefix ;", "Last pane", "Pane Splitting & Nav"},
		{"Prefix Arrow", "Navigate panes", "Pane Splitting & Nav"},

		// Pane Management
		{"Prefix x", "Kill pane", "Pane Management"},
		{"Prefix {", "Move pane left", "Pane Management"},
		{"Prefix }", "Move pane right", "Pane Management"},
		{"Prefix Space", "Next layout", "Pane Management"},
		{"Prefix z", "Toggle zoom pane", "Pane Management"},
		{"Prefix !", "Break pane to new window", "Pane Management"},
		{"Prefix q", "Show pane numbers", "Pane Management"},

		// Copy Mode
		{"Prefix [", "Enter copy mode", "Copy Mode"},
		{"v", "Start selection (vi)", "Copy Mode"},
		{"y", "Copy selection (vi)", "Copy Mode"},
		{"Prefix ]", "Paste buffer", "Copy Mode"},

		// Misc
		{"Prefix t", "Show clock", "Misc"},
		{"Prefix :", "Enter command prompt", "Misc"},
		{"Prefix ?", "List all shortcuts", "Misc"},
	}
}

// Gets the terminal width.
func getTerminalWidth() int {
	cmd := exec.Command("stty", "size")
	cmd.Stdin = os.Stdin
	out, err := cmd.Output()
	if err != nil {
		return 80 // Default width
	}
	parts := strings.Split(strings.TrimSpace(string(out)), " ")
	if len(parts) == 2 {
		width, err := strconv.Atoi(parts[1])
		if err == nil {
			return width
		}
	}
	return 80 // Default width
}

// center pads a string with spaces to center it within a given width.
func center(s string, width int) string {
	if len(s) >= width {
		return s
	}
	padding := (width - len(s)) / 2
	rightPadding := width - len(s) - padding
	return strings.Repeat(" ", padding) + s + strings.Repeat(" ", rightPadding)
}

// formatCell pads a string with spaces to fit in a cell of a given width.
func formatCell(s string, width int) string {
	return fmt.Sprintf("%-*s", width, s)
}

// wordWrap wraps a string to a given line width.
func wordWrap(text string, lineWidth int) []string {
	words := strings.Fields(strings.TrimSpace(text))
	if len(words) == 0 {
		return nil
	}
	var lines []string
	var currentLine string
	for _, word := range words {
		if len(currentLine)+len(word)+1 > lineWidth {
			lines = append(lines, currentLine)
			currentLine = ""
		}
		if currentLine == "" {
			currentLine = word
		} else {
			currentLine += " " + word
		}
	}
	lines = append(lines, currentLine)
	return lines
}

// displayShortcutsInColumns categorizes and prints shortcuts in a dynamic column layout.
func displayShortcutsInColumns(allShortcuts []Shortcut) {
	categorized := make(map[string][]Shortcut)
	for _, s := range allShortcuts {
		categorized[s.Category] = append(categorized[s.Category], s)
	}

	orderedCategories := []string{"Sessions", "Windows", "Pane Splitting & Nav", "Pane Management", "Misc", "Copy Mode"}

	columnWidth := 25
	columnSpacing := 2
	fullColumnWidth := columnWidth + columnSpacing
	terminalWidth := getTerminalWidth()
	numDisplayColumns := terminalWidth / fullColumnWidth
	if numDisplayColumns == 0 {
		numDisplayColumns = 1
	}

	// Block represents an atomic unit with metadata
	type block struct {
		lines      []string
		isHeader   bool
		categoryId int
	}

	// Generate all lines for all categories and shortcuts as atomic blocks
	var allBlocks []block
	var totalLines int
	for i, catName := range orderedCategories {
		// Category header block
		var headerBlock []string
		if i > 0 {
			headerBlock = append(headerBlock, "")
		}
		headerBlock = append(headerBlock, center(catName, columnWidth), strings.Repeat("─", columnWidth), "")
		allBlocks = append(allBlocks, block{lines: headerBlock, isHeader: true, categoryId: i})
		totalLines += len(headerBlock)

		// Each shortcut is its own atomic block
		for _, shortcut := range categorized[catName] {
			var shortcutBlock []string
			shortcutBlock = append(shortcutBlock, formatCell(" "+shortcut.Key, columnWidth))
			descLines := wordWrap(shortcut.Description, columnWidth-2)
			for _, line := range descLines {
				shortcutBlock = append(shortcutBlock, formatCell("  "+line, columnWidth))
			}
			shortcutBlock = append(shortcutBlock, "") // Blank line after each entry
			allBlocks = append(allBlocks, block{lines: shortcutBlock, isHeader: false, categoryId: i})
			totalLines += len(shortcutBlock)
		}
	}
	if totalLines == 0 {
		return
	}

	// Distribute atomic blocks into columns
	columns := make([][]string, numDisplayColumns)
	columnHeight := (totalLines + numDisplayColumns - 1) / numDisplayColumns
	currentCol := 0

	for i, blk := range allBlocks {
		blockLines := blk.lines

		// If this is a header, check if there's at least one shortcut following it
		minSpaceNeeded := len(blockLines)
		if blk.isHeader && i+1 < len(allBlocks) && allBlocks[i+1].categoryId == blk.categoryId {
			// Reserve space for header + at least the first shortcut
			minSpaceNeeded += len(allBlocks[i+1].lines)
		}

		// If the block won't fit in the current column, move to the next
		if len(columns[currentCol]) > 0 && len(columns[currentCol])+minSpaceNeeded > columnHeight {
			currentCol++
			if currentCol >= numDisplayColumns {
				break
			}
			// Skip leading empty lines when starting a new column
			if len(blockLines) > 0 && strings.TrimSpace(blockLines[0]) == "" {
				blockLines = blockLines[1:]
			}
		}

		// Add the entire block to the current column
		columns[currentCol] = append(columns[currentCol], blockLines...)
	}

	// Print the columns
	for row := 0; row < columnHeight; row++ {
		for col := 0; col < numDisplayColumns; col++ {
			line := ""
			if row < len(columns[col]) {
				line = columns[col][row]
			}
			fmt.Print(formatCell(line, columnWidth))
			fmt.Print(strings.Repeat(" ", columnSpacing))
		}
		fmt.Println()
	}
}

func main() {
	fmt.Println()
	shortcuts := getStaticShortcuts()
	for i := range shortcuts {
		shortcuts[i].Key = strings.Replace(shortcuts[i].Key, "Prefix", "ctrl+b", -1)
	}

	displayShortcutsInColumns(shortcuts)
	fmt.Println()
}
