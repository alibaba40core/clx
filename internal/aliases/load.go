package aliases

import (
	"context"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/alibaba40core/clx/internal/yamlutil"
)

const maxAliasFileBytes = 64 * 1024

func parseFile(root *yamlutil.Node) (File, error) {
	out := File{SchemaVersion: SchemaVersion, Entries: nil}
	if root == nil {
		return out, nil
	}
	if v, ok := root.GetString("schema_version"); ok && v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			out.SchemaVersion = n
		}
	}
	aliasesNode, ok := root.GetChild("aliases")
	if !ok || aliasesNode == nil || len(aliasesNode.List) == 0 {
		return out, nil
	}
	entries := make([]Entry, 0, len(aliasesNode.List))
	for _, item := range aliasesNode.List {
		name, _ := item.GetString("name")
		value, _ := item.GetString("value")
		name = strings.TrimSpace(name)
		value = strings.TrimSpace(value)
		if name == "" || value == "" {
			continue
		}
		entries = append(entries, Entry{Name: normalizeName(name), Value: value})
	}
	out.Entries = entries
	return out, nil
}

func readFile(ctx context.Context, path string) (File, error) {
	if err := ctx.Err(); err != nil {
		return File{}, err
	}
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return File{SchemaVersion: SchemaVersion}, nil
		}
		return File{}, err
	}
	defer f.Close()

	root, err := yamlutil.Decode(io.LimitReader(f, maxAliasFileBytes))
	if err != nil {
		return File{}, err
	}
	return parseFile(root)
}
