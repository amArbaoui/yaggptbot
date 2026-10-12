package llm

import (
	"amArbaoui/yaggptbot/app/config"
	"amArbaoui/yaggptbot/app/llm/openrouter"
	"context"
	"encoding/base64"
	"errors"
	"log"
	"strings"
)

var ErrNoSpeech = errors.New("no speech recognized")
var ErrNoText = errors.New("no text extracted")

// Transcribe converts speech to text
func (o *OpenRouterProvider) Transcribe(ctx context.Context, audio []byte, format string) (string, error) {
	resp, err := o.client.Transcribe(ctx, openrouter.TranscriptionRequest{
		Model: config.TranscriptionModel,
		InputAudio: openrouter.InputAudio{
			Data:   base64.StdEncoding.EncodeToString(audio),
			Format: format,
		},
	})
	if err != nil {
		return "", err
	}
	text := strings.Join(strings.Fields(resp.Text), " ")
	if text == "" {
		return "", ErrNoSpeech
	}
	return text, nil
}

// ExtractPDFText converts a PDF to text with the OpenRouter file parser.
// Scanned documents have no text layer, so they are parsed again with OCR.
func (o *OpenRouterProvider) ExtractPDFText(ctx context.Context, filename string, pdf []byte) (string, error) {
	text, err := o.parsePDF(ctx, filename, pdf, config.PDFEngine)
	if err == nil && text != "" {
		return text, nil
	}
	if err != nil {
		log.Printf("failed to parse pdf with %s, trying %s: %v", config.PDFEngine, config.PDFOCREngine, err)
	}
	text, err = o.parsePDF(ctx, filename, pdf, config.PDFOCREngine)
	if err != nil {
		return "", err
	}
	if text == "" {
		return "", ErrNoText
	}
	return text, nil
}

func (o *OpenRouterProvider) parsePDF(ctx context.Context, filename string, pdf []byte, engine string) (string, error) {
	prompt := "."
	req := openrouter.ChatCompletionRequest{
		Model:     config.OpenRouterModelMapping[config.DocumentParserModel],
		Reasoning: openrouter.Reasoning{Effort: "none"},
		// only the parsed file annotation is needed, not the answer
		MaxTokens: 1,
		Messages: []openrouter.CompletionMessage{{
			Role: "user",
			Content: []openrouter.CompletionContent{
				{Type: "text", Text: &prompt},
				{Type: "file", File: &openrouter.File{
					Filename: filename,
					FileData: "data:application/pdf;base64," + base64.StdEncoding.EncodeToString(pdf),
				}},
			},
		}},
		Plugins: []openrouter.Plugin{{ID: openrouter.PluginFileParser, PDF: &openrouter.PDFPluginOptions{Engine: engine}}},
	}
	completion, err := o.client.GetChatCompletion(ctx, req)
	if err != nil {
		return "", err
	}
	parts := make([]string, 0)
	for _, choice := range completion.Choices {
		for _, a := range choice.Message.Annotations {
			if a.File == nil {
				continue
			}
			for _, c := range a.File.Content {
				// the parser wraps the content in <file name="..."> and </file> parts
				if c.Text == nil || strings.HasPrefix(*c.Text, "<file") || *c.Text == "</file>" {
					continue
				}
				parts = append(parts, *c.Text)
			}
		}
	}
	return cleanParsedPDF(strings.Join(parts, "\n")), nil
}

// cleanParsedPDF drops the metadata header added by the parser,
// and returns an empty string when there is no text besides page headers, e.g. for scans
func cleanParsedPDF(text string) string {
	const contentsHeader = "## Contents\n"
	if i := strings.Index(text, contentsHeader); i >= 0 {
		text = text[i+len(contentsHeader):]
	}
	text = strings.TrimSpace(text)
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "### Page ") {
			return text
		}
	}
	return ""
}
