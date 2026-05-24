package intent

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"

	"github.com/alibaba40core/clx/internal/yamlutil"
)

const (
	maxRuleFiles     = 32
	maxTotalRuleBytes = 256 * 1024
)

// LoadRulesFromFS loads one intent per *.yaml file in dir.
func LoadRulesFromFS(fsys fs.FS, dir string) ([]Rule, error) {
	entries, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".yaml") {
			continue
		}
		names = append(names, e.Name())
	}
	sort.Strings(names)
	if len(names) > maxRuleFiles {
		return nil, fmt.Errorf("too many rule files: %d", len(names))
	}

	var rules []Rule
	var totalBytes int64
	for _, name := range names {
		path := filepath.ToSlash(filepath.Join(dir, name))
		data, err := fs.ReadFile(fsys, path)
		if err != nil {
			return nil, err
		}
		if len(strings.TrimSpace(string(data))) == 0 {
			continue
		}
		totalBytes += int64(len(data))
		if totalBytes > maxTotalRuleBytes {
			return nil, fmt.Errorf("rules exceed total size budget")
		}
		r, err := parseRuleFile(data)
		if err != nil {
			if strings.Contains(err.Error(), "rule missing intent") {
				continue
			}
			return nil, fmt.Errorf("%s: %w", name, err)
		}
		rules = append(rules, r)
	}
	return rules, nil
}

func parseRuleFile(data []byte) (Rule, error) {
	root, err := yamlutil.DecodeLimited(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return Rule{}, err
	}
	if _, ok := root.GetString("intent"); !ok {
		return Rule{}, fmt.Errorf("rule missing intent")
	}
	return parseRuleNode(root)
}

// LoadRulesFromReader loads a single rule YAML document.
func LoadRulesFromReader(r io.Reader) (Rule, error) {
	root, err := yamlutil.Decode(r)
	if err != nil {
		return Rule{}, err
	}
	return parseRuleNode(root)
}

// ParseSkillIntents parses a skill intents.yaml document (intents: list).
func ParseSkillIntents(data []byte) ([]Rule, error) {
	root, err := yamlutil.DecodeLimited(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, err
	}
	intentsNode, ok := root.GetChild("intents")
	if !ok || len(intentsNode.List) == 0 {
		return nil, fmt.Errorf("skill file missing intents list")
	}
	rules := make([]Rule, 0, len(intentsNode.List))
	for i, item := range intentsNode.List {
		r, err := parseRuleNode(item)
		if err != nil {
			return nil, fmt.Errorf("intent[%d]: %w", i, err)
		}
		rules = append(rules, r)
	}
	return rules, nil
}
