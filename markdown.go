package main

import (
	"github.com/a-h/templ"
	"github.com/gomarkdown/markdown"
	"github.com/microcosm-cc/bluemonday"
)

func renderMarkdownComponent(body string) templ.Component {
	if body == "" {
		return templ.NopComponent
	}
	html := markdown.ToHTML([]byte(body), nil, nil)
	sanitized := bluemonday.UGCPolicy().Sanitize(string(html))
	return templ.Raw(sanitized)
}