package telegram

import (
	"amArbaoui/yaggptbot/app/config"
	"amArbaoui/yaggptbot/app/llm"
	"amArbaoui/yaggptbot/app/tgformat"
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"unicode/utf8"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var ErrUnsupportedDocument = errors.New("unsupported document")

var audioMimeTypes = map[string]string{
	"audio/ogg":    "ogg",
	"audio/opus":   "ogg",
	"audio/mpeg":   "mp3",
	"audio/mp3":    "mp3",
	"audio/mp4":    "m4a",
	"audio/m4a":    "m4a",
	"audio/x-m4a":  "m4a",
	"audio/wav":    "wav",
	"audio/x-wav":  "wav",
	"audio/flac":   "flac",
	"audio/x-flac": "flac",
	"audio/webm":   "webm",
	"audio/aac":    "aac",
}

var audioExtensions = map[string]string{
	".ogg":  "ogg",
	".oga":  "ogg",
	".opus": "ogg",
	".mp3":  "mp3",
	".m4a":  "m4a",
	".wav":  "wav",
	".flac": "flac",
	".webm": "webm",
	".aac":  "aac",
}

// UserVoiceHandler transcribes voice messages and audio files and answers them as text
func UserVoiceHandler(ctx context.Context, bot *GPTBot, update *tgbotapi.Update) {
	select {
	case <-ctx.Done():
		return
	default:
	}

	m := update.Message
	fileID, duration, format := audioFile(m)
	if format == "" {
		bot.TextReplyWithContext(ctx, config.UnsupportedAudioMessage, m)
		return
	}
	if duration > config.MaxVoiceDuration {
		bot.TextReplyWithContext(ctx, fmt.Sprintf(config.VoiceTooLongMessage, config.MaxVoiceDuration/60), m)
		return
	}
	stopTyping := bot.keepTyping(ctx, m.Chat.ID)
	transcript, err := bot.transcribe(ctx, fileID, format)
	stopTyping()
	if err != nil {
		log.Printf("failed to transcribe message %d: %v", m.MessageID, err)
		replyText := config.VoiceFailedMessage
		switch {
		case errors.Is(err, llm.ErrNoSpeech):
			replyText = config.SpeechNotRecognizedMessage
		case errors.Is(err, ErrFileTooLarge):
			replyText = config.FileTooLargeMessage
		}
		bot.TextReplyWithContext(ctx, replyText, m)
		return
	}
	bot.TextReplyWithContext(ctx, "> 🎤 "+tgformat.EscapeMarkdown(transcript), m)

	msg := *m
	msg.Text = transcript
	if m.ForwardDate != 0 {
		msg.Text = config.ForwardedVoicePrefix + transcript
	}
	bot.handleText(ctx, &msg)
}

// UserDocumentHandler answers documents: images and audio are handled as such,
// text files and PDFs are converted to text and sent to the model together with the caption
func UserDocumentHandler(ctx context.Context, bot *GPTBot, update *tgbotapi.Update) {
	select {
	case <-ctx.Done():
		return
	default:
	}

	m := update.Message
	if imageFileID(m) != "" {
		UserPhotoHandler(ctx, bot, update)
		return
	}
	if _, _, format := audioFile(m); format != "" {
		UserVoiceHandler(ctx, bot, update)
		return
	}
	if m.Document.FileSize > config.MaxFileSize {
		bot.TextReplyWithContext(ctx, config.FileTooLargeMessage, m)
		return
	}
	stopTyping := bot.keepTyping(ctx, m.Chat.ID)
	text, err := bot.documentText(ctx, m.Document)
	stopTyping()
	if err != nil {
		log.Printf("failed to read document of message %d: %v", m.MessageID, err)
		replyText := config.DocumentFailedMessage
		switch {
		case errors.Is(err, ErrUnsupportedDocument):
			replyText = config.DocumentsNotSupportedMessage
		case errors.Is(err, ErrFileTooLarge):
			replyText = config.FileTooLargeMessage
		case errors.Is(err, llm.ErrNoText):
			replyText = config.DocumentNoTextMessage
		}
		bot.TextReplyWithContext(ctx, replyText, m)
		return
	}
	if utf8.RuneCountInString(text) > config.MaxDocumentChars {
		text = string([]rune(text)[:config.MaxDocumentChars])
		bot.TextReplyWithContext(ctx, fmt.Sprintf(config.DocumentTruncatedMessage, config.MaxDocumentChars), m)
	}
	question := m.Caption
	if question == "" {
		question = config.DefaultDocumentCaption
	}
	msg := *m
	msg.Text = fmt.Sprintf("<file name=%q>\n%s\n</file>\n\n%s", m.Document.FileName, text, question)
	bot.handleText(ctx, &msg)
}

func (b *GPTBot) transcribe(ctx context.Context, fileID string, format string) (string, error) {
	data, err := b.downloadFile(ctx, fileID, config.MaxFileSize)
	if err != nil {
		return "", err
	}
	return b.mediaService.Transcribe(ctx, data, format)
}

// documentText returns the text of a PDF or a text file
func (b *GPTBot) documentText(ctx context.Context, doc *tgbotapi.Document) (string, error) {
	data, err := b.downloadFile(ctx, doc.FileID, config.MaxFileSize)
	if err != nil {
		return "", err
	}
	if doc.MimeType == "application/pdf" || bytes.HasPrefix(data, []byte("%PDF-")) {
		return b.mediaService.ExtractPDFText(ctx, doc.FileName, data)
	}
	data = bytes.TrimPrefix(data, []byte("\xef\xbb\xbf"))
	if !utf8.Valid(data) || bytes.IndexByte(data, 0) >= 0 {
		return "", fmt.Errorf("%w: %s", ErrUnsupportedDocument, doc.MimeType)
	}
	return string(data), nil
}

// audioFile returns the file id, duration in seconds and transcription format of a voice message
// or an audio file, the format is empty when the message has no supported audio
func audioFile(m *tgbotapi.Message) (fileID string, duration int, format string) {
	switch {
	case m.Voice != nil:
		return m.Voice.FileID, m.Voice.Duration, "ogg"
	case m.Audio != nil:
		return m.Audio.FileID, m.Audio.Duration, audioFormat(m.Audio.MimeType, m.Audio.FileName)
	case m.Document != nil:
		return m.Document.FileID, 0, audioFormat(m.Document.MimeType, m.Document.FileName)
	}
	return "", 0, ""
}

func audioFormat(mimeType string, fileName string) string {
	if format, ok := audioMimeTypes[strings.ToLower(mimeType)]; ok {
		return format
	}
	return audioExtensions[strings.ToLower(filepath.Ext(fileName))]
}
