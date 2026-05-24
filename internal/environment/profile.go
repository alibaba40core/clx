package environment

// SchemaVersion is the current system_profile.json schema version.
const SchemaVersion = 1

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
