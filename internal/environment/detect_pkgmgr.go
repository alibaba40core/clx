package environment

// Bounded package manager catalog (LookPath only).
var packageManagerCatalog = []string{
	"winget", "choco", "scoop",
	"brew",
	"apt", "apt-get", "dnf", "yum", "pacman", "snap", "flatpak",
	"npm",
}

func detectPackageManagers() []string {
	return findOnPath(packageManagerCatalog)
}
