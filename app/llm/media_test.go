package llm

import (
	"amArbaoui/yaggptbot/app/config"
	"amArbaoui/yaggptbot/app/llm/openrouter"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTranscribe(t *testing.T) {
	var got openrouter.TranscriptionRequest
	var path string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		json.NewDecoder(r.Body).Decode(&got)
		w.Write([]byte(`{"text":"  Привет!  Как   дела? "}`))
	}))
	t.Cleanup(srv.Close)
	provider := NewOpenrouterProvider(openrouter.NewOpenrouterClient(srv.URL, "test"), "")

	text, err := provider.Transcribe(context.Background(), []byte("audio"), "ogg")
	if err != nil {
		t.Fatal(err)
	}
	if text != "Привет! Как дела?" {
		t.Errorf("expected normalized text, got %q", text)
	}
	if path != "/audio/transcriptions" || got.Model != config.TranscriptionModel || got.InputAudio.Format != "ogg" ||
		got.InputAudio.Data != base64.StdEncoding.EncodeToString([]byte("audio")) {
		t.Errorf("unexpected request to %s: %+v", path, got)
	}
}

func TestTranscribeNoSpeech(t *testing.T) {
	provider := newTestOpenrouterProvider(t, http.StatusOK, `{"text":"  "}`)
	if _, err := provider.Transcribe(context.Background(), []byte("audio"), "ogg"); !errors.Is(err, ErrNoSpeech) {
		t.Errorf("expected ErrNoSpeech, got %v", err)
	}
}

// pdfParserServer answers parse requests with the given text for each engine
func pdfParserServer(t *testing.T, textByEngine map[string]string) (*OpenRouterProvider, *[]string) {
	t.Helper()
	engines := make([]string, 0)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req openrouter.ChatCompletionRequest
		json.NewDecoder(r.Body).Decode(&req)
		engine := req.Plugins[0].PDF.Engine
		engines = append(engines, engine)
		if req.MaxTokens != 1 {
			t.Errorf("parse request should not generate an answer, max_tokens %d", req.MaxTokens)
		}
		content := []map[string]string{
			{"type": "text", "text": `<file name="doc.pdf">`},
			{"type": "text", "text": textByEngine[engine]},
			{"type": "text", "text": "</file>"},
		}
		resp := map[string]any{"choices": []any{map[string]any{"message": map[string]any{
			"role": "assistant", "content": "x",
			"annotations": []any{map[string]any{"type": "file", "file": map[string]any{"hash": "h", "name": "doc.pdf", "content": content}}},
		}}}}
		json.NewEncoder(w).Encode(resp)
	}))
	t.Cleanup(srv.Close)
	return NewOpenrouterProvider(openrouter.NewOpenrouterClient(srv.URL, "test"), ""), &engines
}

func TestExtractPDFText(t *testing.T) {
	provider, engines := pdfParserServer(t, map[string]string{
		config.PDFEngine: "# doc.pdf\n## Metadata\n- Producer=x\n\n## Contents\n### Page 1\nСрок аренды: 12 месяцев.",
	})
	text, err := provider.ExtractPDFText(context.Background(), "doc.pdf", []byte("%PDF-1.3"))
	if err != nil {
		t.Fatal(err)
	}
	if text != "### Page 1\nСрок аренды: 12 месяцев." {
		t.Errorf("unexpected text %q", text)
	}
	if len(*engines) != 1 {
		t.Errorf("expected a single parse request, got %v", *engines)
	}
}

func TestExtractPDFTextFallsBackToOCR(t *testing.T) {
	provider, engines := pdfParserServer(t, map[string]string{
		config.PDFEngine:    "## Contents\n### Page 1\n\n### Page 2\n",
		config.PDFOCREngine: "Scanned text",
	})
	text, err := provider.ExtractPDFText(context.Background(), "scan.pdf", []byte("%PDF-1.3"))
	if err != nil || text != "Scanned text" {
		t.Errorf("expected OCR text, got %q, %v", text, err)
	}
	if len(*engines) != 2 || (*engines)[1] != config.PDFOCREngine {
		t.Errorf("expected fallback to OCR, got %v", *engines)
	}
}

func TestExtractPDFTextNoText(t *testing.T) {
	provider, _ := pdfParserServer(t, map[string]string{})
	if _, err := provider.ExtractPDFText(context.Background(), "empty.pdf", []byte("%PDF-1.3")); !errors.Is(err, ErrNoText) {
		t.Errorf("expected ErrNoText, got %v", err)
	}
}
