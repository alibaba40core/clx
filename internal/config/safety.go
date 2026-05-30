package config

import (
	"fmt"
	"strings"
)

// SafetyAction describes what the pipeline should do after risk classification.
type SafetyAction struct {
	ShowExplain bool // enriched explanation in display
	Preview     bool // print dry-run invocation line
	Confirm     bool // y/n prompt before exec
	BlockYes    bool // -y cannot skip Confirm (high mode, medium/high risk)
}

// SafetyFlagOverrides carries per-invocation CLI flags that adjust the action.
type SafetyFlagOverrides struct {
	DryRun bool // --dry-run: preview only, never exec
	Yes    bool // -y: skip confirm when allowed
}

// DecideSafetyAction maps safety mode + command risk to pipeline behavior.
// riskLevel is the command risk name: "low", "medium", or "high".
func DecideSafetyAction(cfg Config, riskLevel string, flags SafetyFlagOverrides) SafetyAction {
	var action SafetyAction
	mode := strings.ToLower(strings.TrimSpace(cfg.Safety.Mode))
	if mode == "" {
		mode = "medium"
	}
	riskLevel = strings.ToLower(strings.TrimSpace(riskLevel))

	switch mode {
	case "custom":
		action.ShowExplain = cfg.Features.Explain
		action.Preview = cfg.Safety.DryRun
		action.Confirm = cfg.Safety.RequireConfirmation
	default:
		action = presetAction(mode, riskLevel)
	}

	if flags.DryRun {
		action.Preview = true
		action.Confirm = false
	}

	return action
}

// PreviewOnly reports whether preview should stop the pipeline (no confirm, no exec).
func (a SafetyAction) PreviewOnly(cfg Config, flags SafetyFlagOverrides) bool {
	if flags.DryRun {
		return true
	}
	if !a.Preview || a.Confirm {
		return false
	}
	// custom dry_run without confirm: config-driven preview cannot be bypassed by -y
	return strings.EqualFold(cfg.Safety.Mode, "custom") && cfg.Safety.DryRun
}

// ShouldConfirm reports whether the y/n prompt should run.
func (a SafetyAction) ShouldConfirm(cfg Config, flags SafetyFlagOverrides) bool {
	if !a.Confirm || cfg.Execution.AutoExecute {
		return false
	}
	if flags.Yes && !a.BlockYes {
		return false
	}
	return true
}

func presetAction(mode, riskLevel string) SafetyAction {
	switch mode {
	case "low":
		switch riskLevel {
		case "high":
			return SafetyAction{Confirm: true}
		default:
			return SafetyAction{}
		}
	case "medium":
		switch riskLevel {
		case "low":
			return SafetyAction{}
		default:
			return SafetyAction{ShowExplain: true, Confirm: true}
		}
	case "high":
		switch riskLevel {
		case "low":
			return SafetyAction{ShowExplain: true, Confirm: true}
		default:
			return SafetyAction{ShowExplain: true, Preview: true, Confirm: true, BlockYes: true}
		}
	default:
		return presetAction("medium", riskLevel)
	}
}

// ApplySafetyMode sets cfg.Safety.Mode to a validated preset name.
func ApplySafetyMode(cfg *Config, mode string) error {
	mode = strings.ToLower(strings.TrimSpace(mode))
	if _, ok := validSafetyModes[mode]; !ok {
		return fmt.Errorf("invalid safety mode %q: must be low, medium, high, or custom", mode)
	}
	cfg.Safety.Mode = mode
	return nil
}

// ApplySafetyCustomToggle sets a custom-mode boolean and switches mode to custom.
func ApplySafetyCustomToggle(cfg *Config, key string, value bool) error {
	key = strings.ToLower(strings.TrimSpace(key))
	switch key {
	case "require_confirmation":
		cfg.Safety.RequireConfirmation = value
	case "dry_run":
		cfg.Safety.DryRun = value
	case "explain":
		cfg.Features.Explain = value
	default:
		return fmt.Errorf("unknown safety toggle %q", key)
	}
	cfg.Safety.Mode = "custom"
	return nil
}

// SafetyShowLines returns human-readable safety settings for clx safety show.
func SafetyShowLines(cfg Config) []string {
	mode := strings.ToLower(strings.TrimSpace(cfg.Safety.Mode))
	if mode == "" {
		mode = "medium"
	}

	var lines []string
	lines = append(lines, fmt.Sprintf("Safety mode: %s", mode))
	lines = append(lines, "")
	lines = append(lines, "  Command risk (low / medium / high) comes from the risk engine for each command.")
	lines = append(lines, "  Safety mode decides what happens after that classification.")
	lines = append(lines, "")

	if mode == "custom" {
		lines = append(lines, "  Custom toggles apply to every command (preset matrix is ignored):")
		lines = append(lines, fmt.Sprintf("    require_confirmation: %t  %s",
			cfg.Safety.RequireConfirmation, toggleHint("require_confirmation", cfg.Safety.RequireConfirmation)))
		lines = append(lines, fmt.Sprintf("    dry_run:              %t  %s",
			cfg.Safety.DryRun, toggleHint("dry_run", cfg.Safety.DryRun)))
		lines = append(lines, fmt.Sprintf("    explain:              %t  %s",
			cfg.Features.Explain, toggleHint("explain", cfg.Features.Explain)))
	} else {
		if summary := modeSummary(mode); summary != "" {
			lines = append(lines, "  "+summary)
			lines = append(lines, "")
		}
		lines = append(lines, "  Command risk → what CLX does:")
		for _, rl := range []string{"low", "medium", "high"} {
			a := presetAction(mode, rl)
			lines = append(lines, fmt.Sprintf("    %-6s %s", rl+":", actionShowLine(a)))
		}
		if mode == "high" {
			lines = append(lines, "")
			lines = append(lines, "  In high mode, -y cannot skip confirmation for medium/high-risk commands.")
		}
	}

	lines = append(lines, "")
	lines = append(lines, "  Change preset:  clx safety set mode=low|medium|high")
	lines = append(lines, "  Custom toggles: clx safety set require_confirmation|dry_run|explain=<true|false>")
	lines = append(lines, "  Per run:        --dry-run (preview only)  -y (skip confirm when allowed)  --explain (display only)")
	return lines
}

func modeSummary(mode string) string {
	switch mode {
	case "low":
		return "Low tolerance: most commands run immediately; only high-risk commands prompt for confirmation."
	case "medium":
		return "Balanced (default): low-risk commands run; medium/high-risk show an explanation and ask before executing."
	case "high":
		return "Strict: all commands at least explain + confirm; medium/high-risk also show a dry-run preview first."
	default:
		return ""
	}
}

func actionShowLine(a SafetyAction) string {
	label := describeAction(a)
	if detail := actionDetail(a); detail != "" {
		return label + " — " + detail
	}
	return label
}

func actionDetail(a SafetyAction) string {
	switch {
	case !a.ShowExplain && !a.Preview && !a.Confirm:
		return "execute after risk and policy checks"
	case a.Preview && a.Confirm && a.ShowExplain:
		return "show explanation, print dry-run preview, ask y/n, then execute if confirmed"
	case a.ShowExplain && a.Confirm && !a.Preview:
		return "show explanation, ask y/n before execute"
	case a.Confirm && !a.ShowExplain && !a.Preview:
		return "ask y/n before execute"
	default:
		return ""
	}
}

func toggleHint(key string, on bool) string {
	if !on {
		switch key {
		case "require_confirmation":
			return "no confirmation prompt"
		case "dry_run":
			return "no dry-run preview"
		case "explain":
			return "minimal display (no enriched explanation)"
		}
	}
	switch key {
	case "require_confirmation":
		return "ask y/n before every command"
	case "dry_run":
		return "print dry-run preview; without confirm, never executes (-y cannot bypass)"
	case "explain":
		return "show enriched explanation in display"
	default:
		return ""
	}
}

func describeAction(a SafetyAction) string {
	if !a.ShowExplain && !a.Preview && !a.Confirm {
		return "run"
	}
	var parts []string
	if a.ShowExplain {
		parts = append(parts, "explain")
	}
	if a.Preview {
		parts = append(parts, "preview")
	}
	if a.Confirm {
		parts = append(parts, "confirm")
	}
	return strings.Join(parts, " + ")
}
