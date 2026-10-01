// Package markdown renders post/project bodies from Markdown to HTML with
// server-side syntax highlighting, and derives reading metadata.
package markdown

import (
	"bytes"
	"regexp"
	"strings"

	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// md is safe for concurrent use once constructed.
var md = goldmark.New(
	goldmark.WithExtensions(
		extension.GFM,
		extension.Footnote,
		highlighting.NewHighlighting(
			// Emit CSS classes (not inline styles) so themes can restyle code
			// via web/static/css/chroma.css.
			highlighting.WithFormatOptions(chromahtml.WithClasses(true)),
		),
	),
	goldmark.WithParserOptions(
		parser.WithAutoHeadingID(),
		parser.WithASTTransformers(util.Prioritized(postTransformer{}, 1000)),
	),
	goldmark.WithRendererOptions(
		html.WithUnsafe(), // trusted single-author content
	),
)

// postTransformer adds a self-link to every h2–h4 (so headings can be shared
// by URL) and marks images as lazily loaded.
type postTransformer struct{}

func (postTransformer) Transform(doc *ast.Document, _ text.Reader, _ parser.Context) {
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch v := n.(type) {
		case *ast.Image:
			v.SetAttributeString("loading", []byte("lazy"))
			v.SetAttributeString("decoding", []byte("async"))
		case *ast.Heading:
			if v.Level < 2 || v.Level > 4 {
				return ast.WalkContinue, nil
			}
			id, ok := v.AttributeString("id")
			if !ok {
				return ast.WalkContinue, nil
			}
			idb, ok := id.([]byte)
			if !ok {
				return ast.WalkContinue, nil
			}
			link := ast.NewLink()
			link.Destination = append([]byte("#"), idb...)
			link.SetAttributeString("class", []byte("anchor"))
			link.SetAttributeString("title", []byte("Link to this section"))
			link.AppendChild(link, ast.NewString([]byte("#")))
			v.AppendChild(v, link)
			return ast.WalkSkipChildren, nil
		}
		return ast.WalkContinue, nil
	})
}

// Render converts Markdown source to HTML.
func Render(src string) (string, error) {
	var buf bytes.Buffer
	if err := md.Convert([]byte(src), &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// wordRe matches runs of word characters for a rough word count.
var wordRe = regexp.MustCompile(`\S+`)

// ReadingMinutes estimates reading time at ~200 wpm, floored at 1.
func ReadingMinutes(src string) int {
	n := len(wordRe.FindAllString(src, -1))
	m := n / 200
	if m < 1 {
		return 1
	}
	return m
}

var tagRe = regexp.MustCompile(`<[^>]*>`)

// Summary produces a plain-text excerpt of up to limit runes from rendered
// HTML, collapsing whitespace and appending an ellipsis when truncated.
func Summary(htmlBody string, limit int) string {
	text := tagRe.ReplaceAllString(htmlBody, " ")
	text = strings.Join(strings.Fields(text), " ")
	if len([]rune(text)) <= limit {
		return text
	}
	r := []rune(text)
	cut := string(r[:limit])
	if i := strings.LastIndex(cut, " "); i > 0 {
		cut = cut[:i]
	}
	return cut + "…"
}
