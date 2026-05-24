package skills

import (
	"io/fs"

	"github.com/alibaba40core/clx/internal/intent"
)

// LoadFromFS loads rules from skills/*/intents.yaml (delegates to intent package).
func LoadFromFS(fsys fs.FS, root string) ([]intent.Rule, error) {
	return intent.LoadSkillsFromFS(fsys, root)
}
