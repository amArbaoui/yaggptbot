// Package tgformat converts markdown produced by LLMs into Telegram HTML messages.
//
// Telegram supports only a small subset of HTML (https://core.telegram.org/bots/api#html-style),
// so headings are rendered as bold text, lists as text with markers and tables as preformatted text.
// The result is split into messages that fit the Telegram length limit without breaking tags.
package tgformat

import (
	"fmt"
	"html"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	east "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

const blockSeparator = "\n\n"

var markdown = goldmark.New(goldmark.WithExtensions(extension.GFM))

var tagRe = regexp.MustCompile(`<[^>]*>`)

// block is a rendered top level markdown block. code is set for code blocks,
// so they can be split by lines when they don't fit into one message.
type block struct {
	html string
	code *codeBlock
}

type codeBlock struct {
	code string
	lang string
}

// Render converts markdown source to Telegram HTML messages, each at most limit characters long.
func Render(source string, limit int) []string {
	src := []byte(source)
	doc := markdown.Parser().Parse(text.NewReader(src))
	r := renderer{src: src}
	blocks := make([]block, 0)
	for n := doc.FirstChild(); n != nil; n = n.NextSibling() {
		if cb, ok := r.codeBlock(n); ok {
			blocks = append(blocks, block{html: cb.html(), code: &cb})
			continue
		}
		if h := strings.TrimSpace(r.block(n, 0)); h != "" {
			blocks = append(blocks, block{html: h})
		}
	}
	return pack(blocks, limit)
}

// Plain strips Telegram HTML markup, it is used to resend a message without formatting.
func Plain(htmlText string) string {
	return html.UnescapeString(tagRe.ReplaceAllString(htmlText, ""))
}

type renderer struct {
	src []byte
}

func (r renderer) block(n ast.Node, depth int) string {
	switch n := n.(type) {
	case *ast.Heading:
		return wrap("b", r.inlines(n))
	case *ast.Paragraph, *ast.TextBlock:
		return r.inlines(n)
	case *ast.ThematicBreak:
		return "——————"
	case *ast.FencedCodeBlock, *ast.CodeBlock:
		cb, _ := r.codeBlock(n)
		return cb.html()
	case *ast.Blockquote:
		return wrap("blockquote", r.children(n, depth, "\n"))
	case *ast.List:
		return r.list(n, depth)
	case *east.Table:
		return r.table(n)
	case *ast.HTMLBlock:
		return escape(strings.TrimRight(r.lines(n), "\n"))
	default:
		if n.Type() == ast.TypeInline {
			return r.inlines(n)
		}
		return r.children(n, depth, blockSeparator)
	}
}

func (r renderer) children(n ast.Node, depth int, sep string) string {
	parts := make([]string, 0, n.ChildCount())
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		if p := strings.TrimSpace(r.block(c, depth)); p != "" {
			parts = append(parts, p)
		}
	}
	return strings.Join(parts, sep)
}

func (r renderer) list(l *ast.List, depth int) string {
	indent := strings.Repeat("   ", depth)
	num := l.Start
	items := make([]string, 0, l.ChildCount())
	for item := l.FirstChild(); item != nil; item = item.NextSibling() {
		marker := "•"
		if l.IsOrdered() {
			marker = fmt.Sprintf("%d.", num)
			num++
		}
		var b strings.Builder
		b.WriteString(indent + marker + " ")
		first := true
		for c := item.FirstChild(); c != nil; c = c.NextSibling() {
			if sub, ok := c.(*ast.List); ok {
				b.WriteString("\n" + r.list(sub, depth+1))
				continue
			}
			part := strings.TrimSpace(r.block(c, depth+1))
			if part == "" {
				continue
			}
			if !first {
				b.WriteString("\n" + indent + "   ")
			}
			b.WriteString(part)
			first = false
		}
		items = append(items, b.String())
	}
	return strings.Join(items, "\n")
}

// table renders a table as preformatted text with aligned columns
func (r renderer) table(t *east.Table) string {
	rows := make([][]string, 0, t.ChildCount())
	widths := make([]int, 0)
	for row := t.FirstChild(); row != nil; row = row.NextSibling() {
		cells := make([]string, 0, row.ChildCount())
		for cell := row.FirstChild(); cell != nil; cell = cell.NextSibling() {
			value := strings.TrimSpace(Plain(r.inlines(cell)))
			if i := len(cells); i >= len(widths) {
				widths = append(widths, 0)
			}
			widths[len(cells)] = max(widths[len(cells)], utf8.RuneCountInString(value))
			cells = append(cells, value)
		}
		rows = append(rows, cells)
	}
	lines := make([]string, 0, len(rows)+1)
	for i, cells := range rows {
		padded := make([]string, len(cells))
		for j, c := range cells {
			padded[j] = c + strings.Repeat(" ", widths[j]-utf8.RuneCountInString(c))
		}
		lines = append(lines, strings.TrimRight(strings.Join(padded, " | "), " "))
		if _, isHeader := t.FirstChild().(*east.TableHeader); i == 0 && isHeader {
			sep := make([]string, len(widths))
			for j, w := range widths {
				sep[j] = strings.Repeat("-", w)
			}
			lines = append(lines, strings.Join(sep, "-|-"))
		}
	}
	return wrap("pre", escape(strings.Join(lines, "\n")))
}

func (r renderer) codeBlock(n ast.Node) (codeBlock, bool) {
	switch n := n.(type) {
	case *ast.FencedCodeBlock:
		return codeBlock{code: strings.TrimRight(r.lines(n), "\n"), lang: string(n.Language(r.src))}, true
	case *ast.CodeBlock:
		return codeBlock{code: strings.TrimRight(r.lines(n), "\n")}, true
	}
	return codeBlock{}, false
}

func (r renderer) lines(n ast.Node) string {
	var b strings.Builder
	lines := n.Lines()
	for i := 0; i < lines.Len(); i++ {
		seg := lines.At(i)
		b.Write(seg.Value(r.src))
	}
	return b.String()
}

func (r renderer) inlines(n ast.Node) string {
	var b strings.Builder
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		r.inline(&b, c)
	}
	return b.String()
}

func (r renderer) inline(b *strings.Builder, n ast.Node) {
	switch n := n.(type) {
	case *ast.Text:
		value := n.Value(r.src)
		if !n.IsRaw() {
			value = util.ResolveEntityNames(util.ResolveNumericReferences(util.UnescapePunctuations(value)))
		}
		b.WriteString(escape(string(value)))
		if n.SoftLineBreak() || n.HardLineBreak() {
			b.WriteString("\n")
		}
	case *ast.String:
		b.WriteString(escape(string(n.Value)))
	case *ast.CodeSpan:
		var code strings.Builder
		for c := n.FirstChild(); c != nil; c = c.NextSibling() {
			switch c := c.(type) {
			case *ast.Text:
				code.Write(c.Value(r.src))
			case *ast.String:
				code.Write(c.Value)
			}
		}
		b.WriteString(wrap("code", escape(code.String())))
	case *ast.Emphasis:
		tag := "i"
		if n.Level >= 2 {
			tag = "b"
		}
		b.WriteString(wrap(tag, r.inlines(n)))
	case *east.Strikethrough:
		b.WriteString(wrap("s", r.inlines(n)))
	case *ast.Link:
		b.WriteString(link(string(n.Destination), r.inlines(n)))
	case *ast.Image:
		label := r.inlines(n)
		if label == "" {
			label = "image"
		}
		b.WriteString(link(string(n.Destination), label))
	case *ast.AutoLink:
		b.WriteString(escape(string(n.Label(r.src))))
	case *ast.RawHTML:
		for i := 0; i < n.Segments.Len(); i++ {
			seg := n.Segments.At(i)
			b.WriteString(escape(string(seg.Value(r.src))))
		}
	case *east.TaskCheckBox:
		if n.IsChecked {
			b.WriteString("☑ ")
		} else {
			b.WriteString("☐ ")
		}
	default:
		b.WriteString(r.inlines(n))
	}
}

func (cb codeBlock) html() string {
	return cb.wrap(escape(cb.code))
}

func (cb codeBlock) wrap(escapedCode string) string {
	if cb.lang == "" {
		return wrap("pre", escapedCode)
	}
	return fmt.Sprintf(`<pre><code class="language-%s">%s</code></pre>`, escape(cb.lang), escapedCode)
}

func link(dest string, label string) string {
	if !strings.HasPrefix(dest, "http://") && !strings.HasPrefix(dest, "https://") && !strings.HasPrefix(dest, "tg://") {
		return label
	}
	return fmt.Sprintf(`<a href="%s">%s</a>`, escape(dest), label)
}

func wrap(tag string, content string) string {
	if content == "" {
		return ""
	}
	return "<" + tag + ">" + content + "</" + tag + ">"
}

func escape(s string) string {
	return html.EscapeString(s)
}

// EscapeMarkdown escapes markdown syntax, so that text is rendered as is
func EscapeMarkdown(s string) string {
	var b strings.Builder
	for _, r := range s {
		if strings.ContainsRune("\\`*_{}[]()<>#+-.!|~", r) {
			b.WriteRune('\\')
		}
		b.WriteRune(r)
	}
	return b.String()
}
