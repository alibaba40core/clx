package capabilities

import "github.com/alibaba40core/clx/internal/intent"

// SelectedStrategy is the chosen rule strategy for the current environment.
type SelectedStrategy struct {
	Key      string
	Strategy intent.Strategy
}
