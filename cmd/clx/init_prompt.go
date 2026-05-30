package main

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

const maxInitPromptAttempts = 5

// parseInitChoice interprets one menu line. Empty input selects defaultIdx.
func parseInitChoice(line string, numOptions int, defaultIdx int) (idx int, ok bool) {
	line = strings.TrimSpace(line)
	if line == "" {
		return defaultIdx, true
	}
	if numOptions <= 0 {
		return 0, false
	}
	var n int
	if _, err := fmt.Sscanf(line, "%d", &n); err != nil || n < 1 || n > numOptions {
		return 0, false
	}
	return n - 1, true
}

// parseYesNoInput interprets y/n input. Empty means defaultNo.
func parseYesNoInput(line string, defaultNo bool) (yes bool, ok bool) {
	line = strings.ToLower(strings.TrimSpace(line))
	if line == "" {
		return !defaultNo, true
	}
	switch line {
	case "y", "yes":
		return true, true
	case "n", "no":
		return false, true
	default:
		return false, false
	}
}

func promptChoice(stdout, stderr io.Writer, in *bufio.Reader, title string, options []string, defaultIdx int) int {
	if len(options) == 0 {
		return 0
	}
	if defaultIdx < 0 || defaultIdx >= len(options) {
		defaultIdx = 0
	}
	for attempt := 0; attempt < maxInitPromptAttempts; attempt++ {
		fmt.Fprintf(stdout, "\n%s:\n", title)
		for i, opt := range options {
			mark := " "
			if i == defaultIdx {
				mark = "*"
			}
			fmt.Fprintf(stdout, "  %s %d) %s\n", mark, i+1, opt)
		}
		fmt.Fprintf(stdout, "Choice [%d]: ", defaultIdx+1)
		line, err := in.ReadString('\n')
		if err != nil && strings.TrimSpace(line) == "" {
			return defaultIdx
		}
		idx, ok := parseInitChoice(line, len(options), defaultIdx)
		if ok {
			return idx
		}
		fmt.Fprintf(stderr, "invalid choice: enter a number 1-%d or press Enter for default [%d]\n",
			len(options), defaultIdx+1)
	}
	fmt.Fprintf(stderr, "too many invalid attempts; using default [%d]\n", defaultIdx+1)
	return defaultIdx
}

func promptYesNo(stdout, stderr io.Writer, in *bufio.Reader, prompt string, defaultNo bool) bool {
	for attempt := 0; attempt < maxInitPromptAttempts; attempt++ {
		fmt.Fprint(stdout, prompt)
		line, err := in.ReadString('\n')
		if err != nil && strings.TrimSpace(line) == "" {
			return defaultNo
		}
		yes, ok := parseYesNoInput(line, defaultNo)
		if ok {
			return yes
		}
		fmt.Fprintf(stderr, "invalid input: enter y/yes, n/no, or press Enter for default (%s)\n",
			yesNoDefaultLabel(defaultNo))
	}
	fmt.Fprintf(stderr, "too many invalid attempts; using default (%s)\n", yesNoDefaultLabel(defaultNo))
	return defaultNo
}

func yesNoDefaultLabel(defaultNo bool) string {
	if defaultNo {
		return "no"
	}
	return "yes"
}
