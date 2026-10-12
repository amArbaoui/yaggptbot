package telegram

import (
	"amArbaoui/yaggptbot/app/config"
	"amArbaoui/yaggptbot/app/llm"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var ErrFileTooLarge = errors.New("file is too large")
var ErrNotAnImage = errors.New("file is not an image")

var fileClient = &http.Client{Timeout: 60 * time.Second}

// imageFileID returns the file id of the image attached to the message:
// the largest photo size, or a document with an image mime type.
func imageFileID(m *tgbotapi.Message) string {
	if len(m.Photo) > 0 {
		return m.Photo[len(m.Photo)-1].FileID
	}
	if m.Document != nil && strings.HasPrefix(m.Document.MimeType, "image/") {
		return m.Document.FileID
	}
	return ""
}

func imageFileSize(m *tgbotapi.Message) int {
	if len(m.Photo) > 0 {
		return m.Photo[len(m.Photo)-1].FileSize
	}
	if m.Document != nil {
		return m.Document.FileSize
	}
	return 0
}

// imageDataURL downloads a telegram image and returns it as a base64 data url.
// Telegram file urls contain the bot token, so the file is never passed to the LLM provider by url.
func (b *GPTBot) imageDataURL(ctx context.Context, fileID string) (string, error) {
	data, err := b.downloadFile(ctx, fileID, config.MaxImageSize)
	if err != nil {
		return "", err
	}
	mimeType := http.DetectContentType(data)
	if !strings.HasPrefix(mimeType, "image/") {
		return "", fmt.Errorf("%w: %s", ErrNotAnImage, mimeType)
	}
	return fmt.Sprintf("data:%s;base64,%s", mimeType, base64.StdEncoding.EncodeToString(data)), nil
}

// downloadFile downloads a telegram file of at most maxSize bytes.
// Errors never include the file url, as it contains the bot token.
func (b *GPTBot) downloadFile(ctx context.Context, fileID string, maxSize int) ([]byte, error) {
	fileURL, err := b.botAPI.GetFileDirectURL(fileID)
	if err != nil {
		return nil, fmt.Errorf("failed to get file: %w", withoutURL(err))
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fileURL, nil)
	if err != nil {
		return nil, errors.New("failed to create file request")
	}
	resp, err := fileClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", withoutURL(err))
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download file, status %d", resp.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, int64(maxSize)+1))
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", withoutURL(err))
	}
	if len(data) > maxSize {
		return nil, ErrFileTooLarge
	}
	return data, nil
}

// withoutURL drops the request url from http errors, so the bot token doesn't end up in logs
func withoutURL(err error) error {
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		return urlErr.Err
	}
	return err
}

func hasImages(messages []llm.CompletionRequestMessage) bool {
	for _, m := range messages {
		if m.ImageUrl != nil {
			return true
		}
	}
	return false
}
