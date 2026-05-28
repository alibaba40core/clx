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
	maxRuleFiles      = 32
	maxTotalRuleBytes = 256 * 1024
	maxIntentsPerFile = 64
)

// LoadRulesFromFS loads rules from *.yaml files in dir.
//
// Each file may contain either a single `intent:` document or an `intents:`
// list of multiple rules. Mixing the two shapes in one file is an error.
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
		parsed, err := parseRulesFile(data)
		if err != nil {
			if err == errFileNoIntents {
				continue
			}
			return nil, fmt.Errorf("%s: %w", name, err)
		}
		rules = append(rules, parsed...)
	}
	return rules, nil
}

// errFileNoIntents is returned for files that declare neither shape; callers
// in LoadRulesFromFS treat this as "skip" to preserve the historical behavior
// where placeholder rule files were silently ignored.
var errFileNoIntents = fmt.Errorf("rule file declares neither intent nor intents")

// parseRulesFile parses one rule file, accepting either a single `intent:`
// document or an `intents:` list. Returns the rules in document order.
func parseRulesFile(data []byte) ([]Rule, error) {
	root, err := yamlutil.DecodeLimited(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, err
	}
	hasList := root.Has("intents")
	hasSingle := root.Has("intent")
	if hasList && hasSingle {
		return nil, fmt.Errorf("rule file declares both intent and intents")
	}
	switch {
	case hasList:
		return parseIntentsListNode(root)
	case hasSingle:
		r, err := parseRuleNode(root)
		if err != nil {
			return nil, err
		}
		return []Rule{r}, nil
	default:
		return nil, errFileNoIntents
	}
}

// parseIntentsListNode parses an `intents:` list under root and returns its
// rules. Used by both rule-file loading and skill loading.
func parseIntentsListNode(root *yamlutil.Node) ([]Rule, error) {
	intentsNode, ok := root.GetChild("intents")
	if !ok || intentsNode == nil || len(intentsNode.List) == 0 {
		return nil, fmt.Errorf("intents list is empty")
	}
	if len(intentsNode.List) > maxIntentsPerFile {
		return nil, fmt.Errorf("too many intents in file: %d", len(intentsNode.List))
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
	return parseIntentsListNode(root)
}
