package providers

import (
	"errors"
	"testing"
)

func TestClassifyHTTPStatus(t *testing.T) {
	t.Parallel()
	cases := []struct {
		status int
		want   error
	}{
		{200, nil},
		{500, ErrUnavailable},
		{429, ErrRateLimited},
		{401, ErrAuth},
		{403, ErrAuth},
		{400, ErrInvalidResp},
	}
	for _, tc := range cases {
		got := ClassifyHTTPStatus(tc.status)
		if tc.want == nil {
			if got != nil {
				t.Fatalf("status %d: got %v want nil", tc.status, got)
			}
			continue
		}
		if !errors.Is(got, tc.want) {
			t.Fatalf("status %d: got %v want %v", tc.status, got, tc.want)
		}
	}
}
