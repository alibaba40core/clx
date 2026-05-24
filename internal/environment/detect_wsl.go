//go:build !windows

package environment

func detectWSL() bool {
	return false
}
