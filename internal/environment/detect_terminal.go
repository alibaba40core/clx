package environment

import (
	"os"
	"strings"
)

func detectTerminal() string {
	if os.Getenv("WT_SESSION") != "" {
		return "windows_terminal"
	}
	// Git Bash / MSYS2 mintty (common on Windows).
	if os.Getenv("MSYSTEM") != "" {
		return "mintty"
	}
	if prog := os.Getenv("TERM_PROGRAM"); prog != "" {
		switch strings.ToLower(prog) {
		case "iterm.app", "iterm2":
			return "iterm"
		case "apple_terminal":
			return "apple_terminal"
		case "vscode", "cursor":
			return "vscode"
		case "hyper":
			return "hyper"
		case "alacritty":
			return "alacritty"
		case "wezterm":
			return "wezterm"
		default:
			return strings.ToLower(prog)
		}
	}
	if os.Getenv("KITTY_WINDOW_ID") != "" {
		return "kitty"
	}
	if os.Getenv("ALACRITTY_LOG") != "" || os.Getenv("ALACRITTY_SOCKET") != "" {
		return "alacritty"
	}
	if term := os.Getenv("TERM"); term != "" && term != "dumb" {
		return "terminal"
	}
	return "unknown"
}
