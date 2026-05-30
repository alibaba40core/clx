package intent

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
)

// LoadSkillsFromFS loads rules from skills/*/intents.yaml under root.
func LoadSkillsFromFS(fsys fs.FS, root string) ([]Rule, error) {
	entries, err := fs.ReadDir(fsys, root)
	if err != nil {
		return nil, err
	}
	var paths []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		p := filepath.ToSlash(filepath.Join(root, e.Name(), "intents.yaml"))
		if _, err := fs.Stat(fsys, p); err != nil {
			continue
		}
		paths = append(paths, p)
	}
	sort.Strings(paths)
	if len(paths) > maxSkillFiles {
		return nil, fmt.Errorf("too many skill packs: %d", len(paths))
	}

	var all []Rule
	for _, p := range paths {
		pack := filepath.Base(filepath.Dir(p))
		data, err := fs.ReadFile(fsys, p)
		if err != nil {
			return nil, err
		}
		rules, err := ParseSkillIntents(data)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", p, err)
		}
		for i := range rules {
			rules[i].SkillPack = pack
		}
		all = append(all, rules...)
	}
	return all, nil
}

const maxSkillFiles = 16
