package helpers

import (
	"strings"

	"github.com/a-h/templ"
	"github.com/gomarkdown/markdown"
	"github.com/microcosm-cc/bluemonday"
)

func deferImages(html string) string {
	var buf strings.Builder
	buf.Grow(len(html) + 256)
	i := 0
	for {
		start := strings.Index(html[i:], "<img")
		if start == -1 {
			buf.WriteString(html[i:])
			break
		}
		start += i
		buf.WriteString(html[i:start])
		end := strings.IndexByte(html[start:], '>')
		if end == -1 {
			buf.WriteString(html[start:])
			break
		}
		end += start + 1
		tag := html[start:end]
		tag = strings.Replace(tag, " src=", " data-src=", 1)
		tag = strings.Replace(tag, "\tsrc=", "\tdata-src=", 1)
		if strings.Contains(tag, "src=") && !strings.Contains(tag, "data-src=") {
			tag = strings.Replace(tag, "src=", "data-src=", 1)
		}
		buf.WriteString(tag)
		i = end
	}
	return buf.String()
}

func RenderMarkdownComponent(body string) templ.Component {
	if body == "" {
		return templ.NopComponent
	}
	html := markdown.ToHTML([]byte(body), nil, nil)
	sanitized := bluemonday.UGCPolicy().Sanitize(string(html))
	return templ.Raw(deferImages(sanitized))
}
