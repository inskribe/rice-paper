package templategen

var TemplateCommentDelimiters = map[string]string{
	"vim":        "\"",
	"lua":        "--",
	"conf":       "#",
	"kitty":      "#",
	"xresources": "!",
	"polybar":    ";",
	"go":         "//",
	"yaml":       "#",
	"toml":       "#",
}

func GetCommentDelimiter(templateName string) string {
	if delim, ok := TemplateCommentDelimiters[templateName]; ok {
		return delim
	}
	return "#"
}
