package generator

import (
	"context"
	"testing"

	"github.com/alibaba40core/clx/internal/capabilities"
	"github.com/alibaba40core/clx/internal/environment"
	"github.com/alibaba40core/clx/internal/intent"
)

func TestInferExecHost(t *testing.T) {
	t.Parallel()
	cases := []struct {
		key  string
		want ExecHost
	}{
		{"powershell", ExecPowerShell},
		{"pwsh", ExecPowerShell},
		{"cmd", ExecCmd},
		{"linux", ExecDirect},
		{"default", ExecDirect},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.key, func(t *testing.T) {
			t.Parallel()
			if got := inferExecHost(tc.key); got != tc.want {
				t.Fatalf("got %v want %v", got, tc.want)
			}
		})
	}
}

func TestRenderExecHostPowerShell(t *testing.T) {
	t.Parallel()
	resolved := intent.ResolvedIntent{Intent: "current_dir", Params: map[string]string{}}
	selected := capabilities.SelectedStrategy{
		Key:      "powershell",
		Strategy: intent.Strategy{Primary: "Get-Location"},
	}
	got, err := Render(context.Background(), resolved, selected, environment.SystemProfile{OS: "windows", Shell: "powershell"})
	if err != nil {
		t.Fatal(err)
	}
	if got.ExecHost != ExecPowerShell {
		t.Fatalf("ExecHost %v want ExecPowerShell", got.ExecHost)
	}
}

func TestRenderExecHostLinuxDirect(t *testing.T) {
	t.Parallel()
	resolved := intent.ResolvedIntent{Intent: "current_dir", Params: map[string]string{}}
	selected := capabilities.SelectedStrategy{
		Key:      "linux",
		Strategy: intent.Strategy{Primary: "pwd"},
	}
	got, err := Render(context.Background(), resolved, selected, environment.SystemProfile{OS: "linux", Shell: "bash"})
	if err != nil {
		t.Fatal(err)
	}
	if got.ExecHost != ExecDirect {
		t.Fatalf("ExecHost %v want ExecDirect", got.ExecHost)
	}
}
