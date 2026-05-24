//go:build darwin

package environment

import (
	"bufio"
	"io"
	"os"
	"regexp"
	"strings"
)

const systemVersionPlist = "/System/Library/CoreServices/SystemVersion.plist"

var productVersionRE = regexp.MustCompile(`<key>ProductVersion</key>\s*<string>([^<]+)</string>`)

func detectOSVersion() string {
	f, err := os.Open(systemVersionPlist)
	if err != nil {
		return ""
	}
	defer f.Close()

	sc := bufio.NewScanner(io.LimitReader(f, 8*1024))
	var buf strings.Builder
	for sc.Scan() {
		buf.WriteString(sc.Text())
	}
	if err := sc.Err(); err != nil {
		return ""
	}
	m := productVersionRE.FindStringSubmatch(buf.String())
	if len(m) >= 2 {
		return m[1]
	}
	return ""
}
