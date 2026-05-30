package yamlutil

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// MaxYAMLBytes is the default per-file decode limit.
const MaxYAMLBytes = 64 * 1024

// Decode reads strict-subset YAML (2-space indent, maps, scalar lists).
func Decode(r io.Reader) (*Node, error) {
	return DecodeLimited(r, MaxYAMLBytes)
}

// DecodeLimited decodes with a custom byte limit.
func DecodeLimited(r io.Reader, maxBytes int64) (*Node, error) {
	limited := io.LimitReader(r, maxBytes)
	sc := bufio.NewScanner(limited)
	sc.Buffer(make([]byte, 0, 4096), int(maxBytes))

	root := &Node{Children: make(map[string]*Node)}
	stack := []frame{{indent: -1, node: root}}

	lineNum := 0
	for sc.Scan() {
		lineNum++
		line := stripInlineComment(sc.Text())
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if strings.Contains(line, "\t") {
			return nil, fmt.Errorf("yaml line %d: tabs are not allowed", lineNum)
		}
		indent := leadingSpaces(line)
		if indent%2 != 0 {
			return nil, fmt.Errorf("yaml line %d: indent must be a multiple of 2", lineNum)
		}
		if strings.ContainsAny(trimmed, "&*!|>") || strings.Contains(trimmed, ": {") {
			return nil, fmt.Errorf("yaml line %d: unsupported syntax", lineNum)
		}

		for len(stack) > 1 && indent <= stack[len(stack)-1].indent {
			stack = stack[:len(stack)-1]
		}
		parent := stack[len(stack)-1].node

		if strings.HasPrefix(trimmed, "-") {
			itemText := strings.TrimSpace(trimmed[1:])
			if itemText == "" {
				return nil, fmt.Errorf("yaml line %d: empty list item", lineNum)
			}
			if strings.Contains(itemText, ":") {
				key, value, hasValue, err := parseKeyValue(itemText, lineNum)
				if err != nil {
					return nil, err
				}
				if hasValue {
					item := &Node{Children: make(map[string]*Node)}
					item.Children[key] = &Node{Scalar: value}
					parent.List = append(parent.List, item)
					stack = append(stack, frame{indent: indent, node: item})
					continue
				}
			}
			parent.List = append(parent.List, &Node{Scalar: itemText})
			continue
		}

		key, value, hasValue, err := parseKeyValue(trimmed, lineNum)
		if err != nil {
			return nil, err
		}

		if !hasValue {
			child := &Node{Children: make(map[string]*Node)}
			parent.Children[key] = child
			stack = append(stack, frame{indent: indent, node: child})
			continue
		}

		parent.Children[key] = &Node{Scalar: value}
	}

	if err := sc.Err(); err != nil {
		return nil, err
	}
	return root, nil
}

type frame struct {
	indent int
	node   *Node
}

func leadingSpaces(s string) int {
	n := 0
	for n < len(s) && s[n] == ' ' {
		n++
	}
	return n
}

func parseKeyValue(line string, lineNum int) (key, value string, hasValue bool, err error) {
	idx := strings.Index(line, ":")
	if idx < 0 {
		return "", "", false, fmt.Errorf("yaml line %d: expected key: value", lineNum)
	}
	key = strings.TrimSpace(line[:idx])
	if key == "" {
		return "", "", false, fmt.Errorf("yaml line %d: empty key", lineNum)
	}
	rest := strings.TrimSpace(line[idx+1:])
	if rest == "" {
		return key, "", false, nil
	}
	return key, unquoteScalar(rest), true, nil
}

func unquoteScalar(s string) string {
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}

// stripInlineComment removes a trailing YAML inline comment (# preceded by
// whitespace or at column 0). # inside quoted scalars and unquoted a#b tokens
// are preserved.
func stripInlineComment(line string) string {
	inSingle := false
	inDouble := false
	for i := 0; i < len(line); i++ {
		switch line[i] {
		case '\'':
			if !inDouble {
				inSingle = !inSingle
			}
		case '"':
			if !inSingle {
				inDouble = !inDouble
			}
		case '#':
			if inSingle || inDouble {
				continue
			}
			if i == 0 {
				return ""
			}
			prev := line[i-1]
			if prev == ' ' || prev == '\t' {
				return strings.TrimRight(line[:i], " \t")
			}
		}
	}
	return line
}
