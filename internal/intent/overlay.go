package intent

import (
	"context"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/alibaba40core/clx/internal/builtin"
	"github.com/alibaba40core/clx/internal/config"
)

// loadBuiltinRulesAndSkills returns built-in rules and skills from the embedded FS.
func loadBuiltinRulesAndSkills() ([]Rule, error) {
	rules, err := LoadRulesFromFS(builtin.FS, "rules")
	if err != nil {
		return nil, err
	}
	skills, err := LoadSkillsFromFS(builtin.FS, "skills")
	if err != nil {
		return nil, err
	}
	return append(rules, skills...), nil
}

// appendUserOverlay loads optional user rules and skills from CLX_HOME.
// Missing directories are silent. Malformed files are skipped with a warning.
func appendUserOverlay(ctx context.Context, logger *slog.Logger, rules []Rule) []Rule {
	if logger == nil {
		logger = slog.Default()
	}
	rulesDir, err := config.UserRulesDir()
	if err == nil {
		rules = append(rules, loadOverlayRules(ctx, logger, rulesDir)...)
	}
	skillsDir, err := config.UserSkillsDir()
	if err == nil {
		rules = append(rules, loadOverlaySkills(ctx, logger, skillsDir)...)
	}
	return rules
}

func loadOverlayRules(ctx context.Context, logger *slog.Logger, dir string) []Rule {
	if _, err := os.Stat(dir); err != nil {
		if !os.IsNotExist(err) {
			logger.Warn("overlay rules dir unavailable", "dir", dir, "err", err)
		}
		return nil
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		logger.Warn("overlay rules dir unreadable", "dir", dir, "err", err)
		return nil
	}
	fsys := os.DirFS(dir)
	var out []Rule
	for _, e := range entries {
		if err := ctx.Err(); err != nil {
			return out
		}
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".yaml") {
			continue
		}
		data, err := fs.ReadFile(fsys, e.Name())
		if err != nil {
			logger.Warn("overlay rule file unreadable", "file", e.Name(), "err", err)
			continue
		}
		parsed, err := parseRulesFile(data)
		if err != nil {
			if err == errFileNoIntents {
				continue
			}
			logger.Warn("overlay rule file skipped", "file", e.Name(), "err", err)
			continue
		}
		out = append(out, parsed...)
	}
	return out
}

func loadOverlaySkills(ctx context.Context, logger *slog.Logger, root string) []Rule {
	if _, err := os.Stat(root); err != nil {
		if !os.IsNotExist(err) {
			logger.Warn("overlay skills dir unavailable", "dir", root, "err", err)
		}
		return nil
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		logger.Warn("overlay skills dir unreadable", "dir", root, "err", err)
		return nil
	}
	var out []Rule
	for _, e := range entries {
		if err := ctx.Err(); err != nil {
			return out
		}
		if !e.IsDir() {
			continue
		}
		path := filepath.ToSlash(filepath.Join(e.Name(), "intents.yaml"))
		fsys := os.DirFS(root)
		data, err := fs.ReadFile(fsys, path)
		if err != nil {
			continue
		}
		parsed, err := ParseSkillIntents(data)
		if err != nil {
			logger.Warn("overlay skill pack skipped", "pack", e.Name(), "err", err)
			continue
		}
		out = append(out, parsed...)
	}
	return out
}
