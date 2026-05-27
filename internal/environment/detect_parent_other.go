//go:build !windows

package environment

func parentProcessBaseName() string { return "" }

func shellFromParentExecutable(string) string { return "" }
