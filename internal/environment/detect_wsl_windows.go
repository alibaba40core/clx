//go:build windows

package environment

func detectWSL() bool {
	_, err := lookPath("wsl.exe")
	return err == nil
}
