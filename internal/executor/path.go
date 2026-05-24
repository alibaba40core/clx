package executor

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var (
	ErrInvalidPath = errors.New("invalid path")
	ErrPathEscape  = errors.New("path escapes allowed roots")
)

// PathOption configures path validation.
type PathOption func(*pathOpts)

type pathOpts struct {
	allowAbs bool
	roots    []string
}

// WithAllowedRoots permits absolute paths under the given roots (cleaned).
func WithAllowedRoots(roots ...string) PathOption {
	return func(o *pathOpts) {
		o.allowAbs = true
		for _, r := range roots {
			if r == "" {
				continue
			}
			clean, err := filepath.Abs(filepath.Clean(r))
			if err == nil {
				o.roots = append(o.roots, clean)
			}
		}
	}
}

// CleanAndValidatePath rejects dangerous path inputs.
func CleanAndValidatePath(p string, opts ...PathOption) (string, error) {
	if p == "" {
		return "", ErrInvalidPath
	}
	if strings.ContainsRune(p, 0) {
		return "", ErrInvalidPath
	}
	for _, c := range []string{"`", "$", ";", "|", "&", ">", "<", "\n", "\r"} {
		if strings.Contains(p, c) {
			return "", ErrInvalidPath
		}
	}

	o := pathOpts{}
	for _, fn := range opts {
		fn(&o)
	}

	clean := filepath.Clean(p)
	if clean == "." && p != "." && p != "./" {
		// filepath.Clean collapses some bad inputs; still check segments
	}
	if strings.Contains(clean, "..") {
		return "", ErrPathEscape
	}

	if filepath.IsAbs(clean) {
		if !o.allowAbs {
			return "", ErrPathEscape
		}
		abs, err := filepath.Abs(clean)
		if err != nil {
			return "", ErrInvalidPath
		}
		if !underAllowedRoot(abs, o.roots) {
			return "", ErrPathEscape
		}
		return abs, nil
	}

	return clean, nil
}

func underAllowedRoot(path string, roots []string) bool {
	if len(roots) == 0 {
		return true
	}
	for _, root := range roots {
		if path == root {
			return true
		}
		rel, err := filepath.Rel(root, path)
		if err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
			return true
		}
	}
	return false
}

// DefaultPathRoots returns cwd, home, and temp for allowlist checks.
func DefaultPathRoots() ([]string, error) {
	var roots []string
	if cwd, err := os.Getwd(); err == nil {
		roots = append(roots, cwd)
	}
	if home, err := os.UserHomeDir(); err == nil {
		roots = append(roots, home)
	}
	roots = append(roots, os.TempDir())
	return roots, nil
}

// PathOptionsFromEnv builds validation options using default roots.
func PathOptionsFromEnv() ([]PathOption, error) {
	roots, err := DefaultPathRoots()
	if err != nil {
		return nil, err
	}
	return []PathOption{WithAllowedRoots(roots...)}, nil
}

// ValidateParamValue validates a single path-like parameter.
func ValidateParamValue(key, value string) error {
	if value == "" {
		return nil
	}
	switch key {
	case "path", "file", "filename":
		opts, err := PathOptionsFromEnv()
		if err != nil {
			return err
		}
		_, err = CleanAndValidatePath(value, opts...)
		return err
	default:
		return nil
	}
}

// ValidateIntentParams validates path-like keys on a resolved intent.
func ValidateIntentParams(params map[string]string) error {
	for k, v := range params {
		if err := ValidateParamValue(k, v); err != nil {
			return fmt.Errorf("param %q: %w", k, err)
		}
	}
	return nil
}
