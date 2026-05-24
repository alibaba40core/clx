//go:build linux

package environment

import (
	"bufio"
	"io"
	"os"
	"strings"
)

const osReleasePath = "/etc/os-release"

func detectOSVersion() string {
	f, err := os.Open(osReleasePath)
	if err != nil {
		return ""
	}
	defer f.Close()

	sc := bufio.NewScanner(io.LimitReader(f, 16*1024))
	var versionID, pretty string
	for sc.Scan() {
		line := sc.Text()
		if strings.HasPrefix(line, "VERSION_ID=") {
			versionID = unquoteOSRelease(strings.TrimPrefix(line, "VERSION_ID="))
		} else if strings.HasPrefix(line, "PRETTY_NAME=") {
			pretty = unquoteOSRelease(strings.TrimPrefix(line, "PRETTY_NAME="))
		}
	}
	if versionID != "" {
		return versionID
	}
	return pretty
}

func unquoteOSRelease(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		return s[1 : len(s)-1]
	}
	return s
}
