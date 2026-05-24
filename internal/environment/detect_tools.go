package environment

// Bounded tool catalog (LookPath only).
var toolCatalog = []string{
	"git", "docker", "node", "npm",
	"python", "python3",
	"kubectl", "helm",
	"rg", "fd", "grep",
	"curl", "wget",
	"make", "go", "ollama",
}

func detectTools() []string {
	return findOnPath(toolCatalog)
}
