package generator

var intentExplanations = map[string]string{
	"find_file":           "Search for a file by name",
	"search_text_in_file": "Search for text inside a file",
	"list_dir":            "List directory contents (ls/ll/dir)",
	"remove_file":         "Delete a file",
	"remove_dir":          "Delete a directory",
	"show_ip_addresses":   "Show local IP addresses",
	"current_dir":         "Print current working directory",
	"disk_usage":          "Show disk usage",
	"find_large_files":    "Find files larger than a size threshold",
	"list_env":            "List environment variables",
	"list_recycle_bin":    "List files in the recycle bin",
	"print_text":          "Print text to stdout",
	"show_installed_ram":  "Show total installed physical memory",
	"list_scheduled_tasks_today": "List scheduled tasks due to run today",
	"show_disk_performance":  "Show disk read and write performance counters",
	"list_local_users":     "List local user accounts",
	"extract_tar_gz":      "Extract a tar.gz archive into the current folder",
	"compare_folders":     "Compare two folders and list differences",
	"create_symlink":      "Create a symbolic link to a directory",
	"find_pdf_downloads":  "Find PDF files in the Downloads folder",
	"git_status":          "Show git working tree status",
	"git_log":             "Show recent git commits (oneline)",
	"git_diff":            "Show unstaged changes in working tree",
	"git_diff_path":       "Show unstaged changes for a path",
	"git_branch_list":     "List local git branches",
	"docker_ps":           "List running Docker containers",
	"docker_images":       "List local Docker images",
	"docker_logs":         "Show recent logs for a Docker container",
	"ping_host":           "Send a small ICMP ping batch to a host",
	"curl_url":            "Fetch HTTP response headers for a URL",
	"netstat_listening":   "List listening TCP sockets on this machine",
}

func explanationFor(intent string) string {
	if s, ok := intentExplanations[intent]; ok {
		return s
	}
	return "Run command for intent " + intent
}
