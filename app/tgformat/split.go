package tgformat

import (
	"strings"
	"unicode/utf16"
)

// pack joins blocks into messages of at most limit characters.
// Blocks longer than the limit are split: code blocks by lines, keeping the code formatting,
// other blocks lose their formatting and are split as plain text.
func pack(blocks []block, limit int) []string {
	messages := make([]string, 0)
	var current strings.Builder
	currentLen := 0
	flush := func() {
		if current.Len() > 0 {
			messages = append(messages, current.String())
			current.Reset()
			currentLen = 0
		}
	}
	for _, b := range blocks {
		for _, piece := range splitBlock(b, limit) {
			pieceLen := length(piece)
			if currentLen > 0 && currentLen+len(blockSeparator)+pieceLen > limit {
				flush()
			}
			if currentLen > 0 {
				current.WriteString(blockSeparator)
				currentLen += len(blockSeparator)
			}
			current.WriteString(piece)
			currentLen += pieceLen
		}
	}
	flush()
	return messages
}

func splitBlock(b block, limit int) []string {
	if limit <= 0 || length(b.html) <= limit {
		return []string{b.html}
	}
	if b.code != nil {
		return splitText(b.code.code, limit, b.code.wrap)
	}
	return splitText(Plain(b.html), limit, func(s string) string { return s })
}

// splitText splits text by lines (or by characters for too long lines) into pieces,
// so that wrap(escape(piece)) is at most limit characters long.
func splitText(s string, limit int, wrapFn func(string) string) []string {
	budget := limit - (length(wrapFn("x")) - 1)
	if budget <= 0 {
		budget = 1
	}
	pieces := make([]string, 0)
	var current strings.Builder
	currentLen := 0
	flush := func() {
		if piece := strings.TrimRight(current.String(), "\n"); piece != "" {
			pieces = append(pieces, wrapFn(piece))
		}
		current.Reset()
		currentLen = 0
	}
	for _, line := range strings.SplitAfter(s, "\n") {
		lineLen := length(escape(line))
		if currentLen+lineLen > budget {
			flush()
		}
		if lineLen <= budget {
			current.WriteString(escape(line))
			currentLen += lineLen
			continue
		}
		for _, r := range line {
			e := escape(string(r))
			if currentLen+length(e) > budget {
				flush()
			}
			current.WriteString(e)
			currentLen += length(e)
		}
	}
	flush()
	return pieces
}

// length returns the length of s in UTF-16 code units, which is how Telegram counts message length.
func length(s string) int {
	n := 0
	for _, r := range s {
		n += utf16.RuneLen(r)
	}
	return n
}
