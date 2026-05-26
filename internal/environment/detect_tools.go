package environment

// Bounded tool catalog (LookPath only).
// Network tooling (ping/ss/netstat) is probed so requires_tool gating in
// rules can prefer ss over netstat on Linux, etc.
var toolCatalog = []string{
	"git", "docker", "node", "npm",
	"python", "python3",
	"kubectl", "helm",
	"rg", "fd", "grep",
	"curl", "wget",
	"make", "go", "ollama",
	"ping", "ss", "netstat",
}

func detectTools() []string {
	return findOnPath(toolCatalog)
}
