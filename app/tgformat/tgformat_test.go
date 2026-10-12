package tgformat

import (
	"strings"
	"testing"
)

func TestRender(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		expected string
	}{
		{"plain text", "hello world", "hello world"},
		{"escapes html", "a < b && c > d", "a &lt; b &amp;&amp; c &gt; d"},
		{"bold and italic", "**bold** and *italic* and __b__ _i_", "<b>bold</b> and <i>italic</i> and <b>b</b> <i>i</i>"},
		{"strikethrough", "~~old~~ new", "<s>old</s> new"},
		{"inline code", "run `a<b && c`", "run <code>a&lt;b &amp;&amp; c</code>"},
		{"backslash escapes", `not \*bold\* snake\_case`, "not *bold* snake_case"},
		{"underscores in words", "snake_case_name", "snake_case_name"},
		{"heading", "### Title\n\ntext", "<b>Title</b>\n\ntext"},
		{"soft line breaks are kept", "line one\nline two", "line one\nline two"},
		{"link", "[site](https://example.com?a=1&b=2)", `<a href="https://example.com?a=1&amp;b=2">site</a>`},
		{"unsafe link scheme", "[x](javascript:alert(1))", "x"},
		{"fenced code", "```go\nfmt.Println(\"<hi>\")\n```", `<pre><code class="language-go">fmt.Println(&#34;&lt;hi&gt;&#34;)</code></pre>`},
		{"code without language", "```\nx := 1\n```", "<pre>x := 1</pre>"},
		{"unordered list", "- one\n- **two**", "• one\n• <b>two</b>"},
		{"ordered list", "3. one\n4. two", "3. one\n4. two"},
		{"nested list", "- one\n  - inner\n- two", "• one\n   • inner\n• two"},
		{"task list", "- [x] done\n- [ ] todo", "• ☑ done\n• ☐ todo"},
		{"blockquote", "> quoted\n> text", "<blockquote>quoted\ntext</blockquote>"},
		{"thematic break", "a\n\n---\n\nb", "a\n\n——————\n\nb"},
		{"raw html is escaped", "<b>not bold</b>", "&lt;b&gt;not bold&lt;/b&gt;"},
		{"table", "| a | bb |\n|---|---|\n| ccc | d |", "<pre>a   | bb\n----|---\nccc | d</pre>"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Render(tt.source, 4096)
			if len(got) != 1 || got[0] != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestRenderEmpty(t *testing.T) {
	if got := Render("  \n\n ", 4096); len(got) != 0 {
		t.Errorf("expected no messages, got %q", got)
	}
}

func TestRenderSplitsByBlocks(t *testing.T) {
	paragraph := strings.Repeat("a", 40)
	source := strings.Repeat(paragraph+"\n\n", 5)
	got := Render(source, 100)
	if len(got) != 3 {
		t.Fatalf("expected 3 messages, got %d: %q", len(got), got)
	}
	for _, m := range got {
		if length(m) > 100 {
			t.Errorf("message exceeds limit: %d", length(m))
		}
		if strings.HasPrefix(m, "\n") || strings.HasSuffix(m, "\n") {
			t.Errorf("message has dangling separators: %q", m)
		}
	}
}

func TestRenderSplitsLongCodeBlock(t *testing.T) {
	code := strings.Repeat("x := a < b\n", 30)
	got := Render("```go\n"+code+"```", 120)
	if len(got) < 2 {
		t.Fatalf("expected the code block to be split, got %d messages", len(got))
	}
	var joined strings.Builder
	for _, m := range got {
		if length(m) > 120 {
			t.Errorf("message exceeds limit: %d, %q", length(m), m)
		}
		if !strings.HasPrefix(m, `<pre><code class="language-go">`) || !strings.HasSuffix(m, "</code></pre>") {
			t.Errorf("each piece should be a complete code block, got %q", m)
		}
		joined.WriteString(Plain(m) + "\n")
	}
	if joined.String() != code {
		t.Errorf("code content changed after split")
	}
}

func TestRenderSplitsLongParagraph(t *testing.T) {
	source := "**" + strings.Repeat("word & ", 100) + "**"
	got := Render(source, 50)
	for _, m := range got {
		if length(m) > 50 {
			t.Errorf("message exceeds limit: %d, %q", length(m), m)
		}
		if strings.Contains(m, "<") {
			t.Errorf("oversized block should be split as plain text, got %q", m)
		}
	}
	if strings.Join(got, "") == "" || !strings.Contains(got[0], "&amp;") {
		t.Errorf("expected escaped plain text, got %q", got)
	}
}

func TestLengthCountsUTF16(t *testing.T) {
	if l := length("a😀"); l != 3 {
		t.Errorf("expected 3, got %d", l)
	}
}

func TestPlain(t *testing.T) {
	if got := Plain(`<b>a &lt; b</b> <a href="x">link</a>`); got != "a < b link" {
		t.Errorf("got %q", got)
	}
}
