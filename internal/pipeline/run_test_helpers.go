package pipeline

import (
	"context"
	"testing"

	"github.com/alibaba40core/clx/internal/config"
	"github.com/alibaba40core/clx/internal/environment"
)

func testProfile(t *testing.T, osName, shell string) {
	t.Helper()
	if osName == "" || shell == "" {
		mp := environment.MinimalProfile()
		osName = mp.OS
		shell = mp.Shell
	}
	path, err := config.SystemProfilePath()
	if err != nil {
		t.Fatal(err)
	}
	store := environment.NewProfileStore()
	store.UpsertProfile(environment.SystemProfile{
		OS:             osName,
		Shell:          shell,
		AvailableTools: []string{"grep"},
	})
	if err := environment.SaveStore(context.Background(), path, store); err != nil {
		t.Fatal(err)
	}
}
