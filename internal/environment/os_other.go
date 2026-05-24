//go:build !linux && !windows && !darwin

package environment

func detectOSVersion() string {
	return ""
}
