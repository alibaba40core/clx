//go:build !windows

package environment

func parentProcessBaseName() string { return "" }
