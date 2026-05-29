package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"
)

const maxSecretInputBytes = 64 * 1024

func stdinIsTerminal(stdin *os.File) bool {
	if stdin == nil {
		return false
	}
	return term.IsTerminal(int(stdin.Fd()))
}

func readPlainStdin(stdin io.Reader) (string, error) {
	data, err := io.ReadAll(io.LimitReader(stdin, maxSecretInputBytes))
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

// readSecretValue reads a secret from stdin. On an interactive terminal it prompts
// with hidden input; otherwise it reads from a pipe or redirected stdin.
func readSecretValue(stdin *os.File, stderr io.Writer, pathKey string) (string, error) {
	if stdin != nil && stdinIsTerminal(stdin) {
		fmt.Fprintf(stderr, "Enter %s: ", pathKey)
		buf, err := term.ReadPassword(int(stdin.Fd()))
		fmt.Fprintln(stderr)
		if err != nil {
			return "", fmt.Errorf("read secret: %w", err)
		}
		return strings.TrimSpace(string(buf)), nil
	}
	if stdin == nil {
		return "", fmt.Errorf("stdin unavailable")
	}
	value, err := readPlainStdin(stdin)
	if err != nil {
		return "", fmt.Errorf("read stdin: %w", err)
	}
	if value == "" {
		return "", fmt.Errorf("empty secret value")
	}
	return value, nil
}
