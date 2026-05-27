// Package builtin exposes the embedded default rule and skill catalog.
// Defaults ship inside the binary so clx works from any cwd without
// access to the source tree. User overlays live in ~/.clx/ and are
// loaded separately by the engine constructor.
package builtin

import "embed"

//go:embed all:rules all:skills
var FS embed.FS
