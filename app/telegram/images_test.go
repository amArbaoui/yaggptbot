package telegram

import (
	"amArbaoui/yaggptbot/app/config"
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const testToken = "123:secret-token"

// rewriteTransport sends all requests to the test server, telegram file urls are not configurable
type rewriteTransport struct {
	target *url.URL
}

func (t rewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	req.URL.Scheme = t.target.Scheme
	req.URL.Host = t.target.Host
	return http.DefaultTransport.RoundTrip(req)
}

func newTestBotWithFile(t *testing.T, file []byte) *GPTBot {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/getMe"):
			w.Write([]byte(`{"ok":true,"result":{"id":1,"is_bot":true,"username":"test"}}`))
		case strings.HasSuffix(r.URL.Path, "/getFile"):
			w.Write([]byte(`{"ok":true,"result":{"file_id":"f","file_path":"photos/f.jpg"}}`))
		case r.URL.Path == "/file/bot"+testToken+"/photos/f.jpg":
			w.Write(file)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)
	target, _ := url.Parse(srv.URL)
	client := &http.Client{Transport: rewriteTransport{target: target}}
	prevClient := fileClient
	fileClient = client
	t.Cleanup(func() { fileClient = prevClient })
	botAPI, err := tgbotapi.NewBotAPIWithClient(testToken, tgbotapi.APIEndpoint, client)
	if err != nil {
		t.Fatalf("failed to create bot api: %v", err)
	}
	return &GPTBot{botAPI: botAPI}
}

var pngHeader = []byte("\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR")

func TestImageDataURL(t *testing.T) {
	bot := newTestBotWithFile(t, pngHeader)
	got, err := bot.imageDataURL(context.Background(), "f")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(got, "data:image/png;base64,") {
		t.Errorf("expected png data url, got %q", got)
	}
	if strings.Contains(got, testToken) {
		t.Errorf("data url must not contain the bot token")
	}
}

func TestImageDataURLRejectsNotImages(t *testing.T) {
	bot := newTestBotWithFile(t, []byte("just text"))
	if _, err := bot.imageDataURL(context.Background(), "f"); !errors.Is(err, ErrNotAnImage) {
		t.Errorf("expected ErrNotAnImage, got %v", err)
	}
}

func TestImageDataURLRejectsLargeImages(t *testing.T) {
	bot := newTestBotWithFile(t, append(pngHeader, bytes.Repeat([]byte{0}, config.MaxImageSize)...))
	if _, err := bot.imageDataURL(context.Background(), "f"); !errors.Is(err, ErrImageTooLarge) {
		t.Errorf("expected ErrImageTooLarge, got %v", err)
	}
}

func TestWithoutURLHidesToken(t *testing.T) {
	err := &url.Error{Op: "Get", URL: "https://api.telegram.org/file/bot" + testToken + "/x", Err: errors.New("timeout")}
	if got := withoutURL(err).Error(); strings.Contains(got, testToken) {
		t.Errorf("token leaked into error: %q", got)
	}
}

func TestImageFileID(t *testing.T) {
	tests := []struct {
		name     string
		message  tgbotapi.Message
		expected string
	}{
		{"largest photo", tgbotapi.Message{Photo: []tgbotapi.PhotoSize{{FileID: "small"}, {FileID: "large"}}}, "large"},
		{"image document", tgbotapi.Message{Document: &tgbotapi.Document{FileID: "doc", MimeType: "image/png"}}, "doc"},
		{"other document", tgbotapi.Message{Document: &tgbotapi.Document{FileID: "doc", MimeType: "application/pdf"}}, ""},
		{"text", tgbotapi.Message{Text: "hi"}, ""},
	}
	for _, tt := range tests {
		if got := imageFileID(&tt.message); got != tt.expected {
			t.Errorf("%s: expected %q, got %q", tt.name, tt.expected, got)
		}
	}
}

func TestConversationChainAttachesOnlyLatestImages(t *testing.T) {
	bot := newTestBotWithFile(t, pngHeader)
	rep := newTestMessageRepository(t)
	bot.msgService = &MessageDbService{rep: rep}
	bot.botOptions = BotOptions{MaxConversationDepth: config.MaxMessageContextDepth}
	photos := []tgbotapi.PhotoSize{{FileID: "f"}}
	total := config.MaxContextImages + 2
	for i := 1; i <= total; i++ {
		m := &tgbotapi.Message{MessageID: i, Chat: &tgbotapi.Chat{ID: 1}, Photo: photos}
		if i > 1 {
			m.ReplyToMessage = &tgbotapi.Message{MessageID: i - 1}
		}
		if err := rep.SaveMessage(context.Background(), m, "user"); err != nil {
			t.Fatal(err)
		}
	}
	next := &tgbotapi.Message{MessageID: total + 1, Chat: &tgbotapi.Chat{ID: 1}, ReplyToMessage: &tgbotapi.Message{MessageID: total}}
	chain, err := bot.GetConversationChain(context.Background(), next)
	if err != nil {
		t.Fatal(err)
	}
	if len(chain) != total {
		t.Fatalf("expected %d messages, got %d", total, len(chain))
	}
	for i, m := range chain {
		attached := i >= total-config.MaxContextImages
		if attached != (m.ImageUrl != nil) {
			t.Errorf("message %d: expected image attached=%v", i, attached)
		}
		if !attached && m.Text != "[image]" {
			t.Errorf("message %d: expected placeholder text, got %q", i, m.Text)
		}
	}
}
