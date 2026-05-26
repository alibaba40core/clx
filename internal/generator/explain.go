package generator

var intentExplanations = map[string]string{
	"find_file":           "Search for a file by name",
	"search_text_in_file": "Search for text inside a file",
	"list_dir":            "List directory contents",
	"current_dir":         "Print current working directory",
	"disk_usage":          "Show disk usage",
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
