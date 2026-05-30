package providers

import (
	"bytes"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"sync"

	"github.com/alibaba40core/clx/internal/builtin"
	"github.com/alibaba40core/clx/internal/intent"
	"github.com/alibaba40core/clx/internal/yamlutil"
)

const (
	maxSkillPromptPacks = 16
	maxSkillPromptBytes = 2048
)

var (
	skillPromptsOnce sync.Once
	skillPrompts     map[string]string
	skillPromptsErr  error
)

// SkillPromptsForEngine returns bounded per-skill prompt text for packs that
// contribute intents present in the engine.
func SkillPromptsForEngine(eng *intent.Engine) map[string]string {
	if eng == nil {
		return nil
	}
	all, err := loadSkillPrompts(builtin.FS, "skills")
	if err != nil || len(all) == 0 {
		return nil
	}
	packs := eng.SkillPacks()
	if len(packs) == 0 {
		return nil
	}
	out := make(map[string]string, len(packs))
	for _, pack := range packs {
		if text, ok := all[pack]; ok && text != "" {
			out[pack] = text
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func loadSkillPrompts(fsys fs.FS, root string) (map[string]string, error) {
	skillPromptsOnce.Do(func() {
		skillPrompts, skillPromptsErr = loadSkillPromptsLocked(fsys, root)
	})
	return skillPrompts, skillPromptsErr
}

func loadSkillPromptsLocked(fsys fs.FS, root string) (map[string]string, error) {
	entries, err := fs.ReadDir(fsys, root)
	if err != nil {
		return nil, err
	}
	out := make(map[string]string)
	for _, e := range entries {
		if !e.IsDir() || len(out) >= maxSkillPromptPacks {
			continue
		}
		pack := e.Name()
		p := filepath.ToSlash(filepath.Join(root, pack, "prompts.yaml"))
		data, err := fs.ReadFile(fsys, p)
		if err != nil {
			continue
		}
		if len(data) > maxSkillPromptBytes {
			return nil, fmt.Errorf("skill prompt %s exceeds %d bytes", pack, maxSkillPromptBytes)
		}
		rootNode, err := yamlutil.Decode(bytes.NewReader(data))
		if err != nil {
			return nil, fmt.Errorf("%s: %w", p, err)
		}
		text, _ := rootNode.GetString("prompt")
		text = trimPrompt(text)
		if text != "" {
			out[pack] = text
		}
	}
	return out, nil
}

func trimPrompt(s string) string {
	const maxRunes = 1500
	if len(s) <= maxRunes {
		return s
	}
	return s[:maxRunes]
}

// SkillPacksSorted returns sorted pack names from a hints map.
func SkillPacksSorted(hints map[string]string) []string {
	if len(hints) == 0 {
		return nil
	}
	out := make([]string, 0, len(hints))
	for k := range hints {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
