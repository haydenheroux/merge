package helpers

import (
	"regexp"
	"strings"

	"github.com/a-h/templ"
	"github.com/gomarkdown/markdown"
	"github.com/microcosm-cc/bluemonday"
)

var prRefRe = regexp.MustCompile(`#(\d+)\b`)

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

// TODO: @keyframes spin may not be visible if SCSS isn't recompiled — verify after build
func linkPRRefs(html string) string {
	var buf strings.Builder
	buf.Grow(len(html) + 128)
	inTag := false
	inAnchor := false

	for i := 0; i < len(html); i++ {
		c := html[i]
		if c == '<' && i+1 < len(html) {
			if html[i+1] == '/' {
				if strings.HasPrefix(html[i:], "</a>") {
					inAnchor = false
				}
				inTag = true
			} else if html[i+1] == 'a' && (i+2 >= len(html) || html[i+2] == ' ' || html[i+2] == '>') {
				inAnchor = true
				inTag = true
			} else {
				inTag = true
			}
		}
		if inTag {
			end := strings.IndexByte(html[i:], '>')
			if end == -1 {
				buf.WriteString(html[i:])
				return buf.String()
			}
			buf.WriteString(html[i : i+end+1])
			i += end
			inTag = false
			continue
		}
		if c == '#' && !inAnchor && i+1 < len(html) && html[i+1] >= '0' && html[i+1] <= '9' {
			if loc := prRefRe.FindStringSubmatch(html[i:]); loc != nil {
				buf.WriteString(`<a class="pr-ref" data-pr-number="`)
				buf.WriteString(loc[1])
				buf.WriteString(`">`)
				buf.WriteString(loc[0])
				buf.WriteString(`</a>`)
				i += len(loc[0]) - 1
				continue
			}
		}
		buf.WriteByte(c)
	}
	return buf.String()
}

func RenderMarkdownComponent(body string) templ.Component {
	if body == "" {
		return templ.NopComponent
	}
	html := markdown.ToHTML([]byte(body), nil, nil)
	sanitized := bluemonday.UGCPolicy().Sanitize(string(html))
	return templ.Raw(linkPRRefs(deferImages(sanitized)))
}
