package generator

var intentExplanations = map[string]string{
	"find_file":           "Search for a file by name",
	"search_text_in_file": "Search for text inside a file",
	"list_dir":            "List directory contents",
	"current_dir":         "Print current working directory",
	"disk_usage":          "Show disk usage",
	"git_status":          "Show git working tree status",
}

func explanationFor(intent string) string {
	if s, ok := intentExplanations[intent]; ok {
		return s
	}
	return "Run command for intent " + intent
}
