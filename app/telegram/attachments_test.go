package telegram

import (
	"context"
	"errors"
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type stubMediaService struct {
	pdfText string
}

func (s stubMediaService) Transcribe(ctx context.Context, audio []byte, format string) (string, error) {
	return "transcript", nil
}

func (s stubMediaService) ExtractPDFText(ctx context.Context, filename string, pdf []byte) (string, error) {
	return s.pdfText, nil
}

func TestAudioFile(t *testing.T) {
	tests := []struct {
		name     string
		message  tgbotapi.Message
		format   string
		duration int
	}{
		{"voice", tgbotapi.Message{Voice: &tgbotapi.Voice{FileID: "v", Duration: 5}}, "ogg", 5},
		{"mp3 audio", tgbotapi.Message{Audio: &tgbotapi.Audio{FileID: "a", MimeType: "audio/mpeg", Duration: 60}}, "mp3", 60},
		{"audio by extension", tgbotapi.Message{Audio: &tgbotapi.Audio{FileID: "a", FileName: "memo.M4A"}}, "m4a", 0},
		{"audio document", tgbotapi.Message{Document: &tgbotapi.Document{FileID: "d", MimeType: "audio/x-wav"}}, "wav", 0},
		{"unknown audio", tgbotapi.Message{Audio: &tgbotapi.Audio{FileID: "a", MimeType: "audio/x-ms-wma", FileName: "a.wma"}}, "", 0},
		{"pdf document", tgbotapi.Message{Document: &tgbotapi.Document{FileID: "d", MimeType: "application/pdf"}}, "", 0},
		{"text", tgbotapi.Message{Text: "hi"}, "", 0},
	}
	for _, tt := range tests {
		_, duration, format := audioFile(&tt.message)
		if format != tt.format || duration != tt.duration {
			t.Errorf("%s: expected %q/%d, got %q/%d", tt.name, tt.format, tt.duration, format, duration)
		}
	}
}

func TestDocumentText(t *testing.T) {
	tests := []struct {
		name     string
		file     []byte
		doc      tgbotapi.Document
		expected string
		err      error
	}{
		{"text file", []byte("\xef\xbb\xbfpackage main\n"), tgbotapi.Document{FileID: "f", FileName: "main.go"}, "package main\n", nil},
		{"pdf by mime type", []byte("anything"), tgbotapi.Document{FileID: "f", MimeType: "application/pdf"}, "pdf text", nil},
		{"pdf by content", []byte("%PDF-1.7 ..."), tgbotapi.Document{FileID: "f", FileName: "report"}, "pdf text", nil},
		{"binary file", []byte("PK\x03\x04\x00\x00"), tgbotapi.Document{FileID: "f", FileName: "doc.docx"}, "", ErrUnsupportedDocument},
		{"invalid utf8", []byte{0xff, 0xfe, 'a'}, tgbotapi.Document{FileID: "f", FileName: "a.txt"}, "", ErrUnsupportedDocument},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bot := newTestBotWithFile(t, tt.file)
			bot.mediaService = stubMediaService{pdfText: "pdf text"}
			got, err := bot.documentText(context.Background(), &tt.doc)
			if !errors.Is(err, tt.err) || got != tt.expected {
				t.Errorf("expected %q, %v, got %q, %v", tt.expected, tt.err, got, err)
			}
		})
	}
}
