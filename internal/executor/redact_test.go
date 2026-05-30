package executor

import "testing"

func TestContainsSecret(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in   string
		want bool
	}{
		{"filename.log", false},
		{"token=sk-abcdefghijklmnopqrst", true},
		{"api_key=supersecret", true},
		{"Bearer abc.def-123", true},
	}
	for _, tc := range cases {
		if got := ContainsSecret(tc.in); got != tc.want {
			t.Errorf("ContainsSecret(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
}
