package environment

import "strings"

// SchemaVersion is the current system_profile.json schema version.
const SchemaVersion = 2

// ProfileStore holds per-shell system profiles in ~/.clx/system_profile.json.
type ProfileStore struct {
	SchemaVersion int                      `json:"schema_version"`
	Profiles      map[string]SystemProfile `json:"profiles"`
}

// SystemProfile describes the current machine environment for command generation.
type SystemProfile struct {
	SchemaVersion   int               `json:"schema_version"`
	OS              string            `json:"os"`
	OSVersion       string            `json:"os_version"`
	Shell           string            `json:"shell"`
	ShellVersion    string            `json:"shell_version"`
	Terminal        string            `json:"terminal"`
	PackageManagers []string          `json:"package_managers"`
	AvailableTools  []string          `json:"available_tools"`
	WSLEnabled      bool              `json:"wsl_enabled"`
	Paths           map[string]string `json:"paths"`
}

// ProfileKey returns the store map key for os and shell (e.g. "windows-powershell").
func ProfileKey(os, shell string) string {
	return strings.ToLower(os) + "-" + strings.ToLower(shell)
}

// NewProfileStore returns an empty v2 profile store.
func NewProfileStore() ProfileStore {
	return ProfileStore{
		SchemaVersion: SchemaVersion,
		Profiles:      make(map[string]SystemProfile),
	}
}

// UpsertProfile inserts or replaces a profile in the store keyed by os-shell.
func (s *ProfileStore) UpsertProfile(p SystemProfile) {
	if s.Profiles == nil {
		s.Profiles = make(map[string]SystemProfile)
	}
	p.SchemaVersion = SchemaVersion
	s.Profiles[ProfileKey(p.OS, p.Shell)] = p
	s.SchemaVersion = SchemaVersion
}
