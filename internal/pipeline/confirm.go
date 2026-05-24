package pipeline

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// confirmPrompt asks [Y/n] and returns true only for explicit yes.
func confirmPrompt(r io.Reader, w io.Writer) (bool, error) {
	if _, err := fmt.Fprint(w, "Execute? [Y/n] "); err != nil {
		return false, err
	}
	sc := bufio.NewScanner(r)
	if !sc.Scan() {
		return false, sc.Err()
	}
	ans := strings.TrimSpace(strings.ToLower(sc.Text()))
	return ans == "y" || ans == "yes", nil
}
