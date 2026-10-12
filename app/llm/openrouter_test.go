package llm

import (
	"amArbaoui/yaggptbot/app/config"
	"amArbaoui/yaggptbot/app/llm/openrouter"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestOpenrouterProvider(t *testing.T, status int, body string) *OpenRouterProvider {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return NewOpenrouterProvider(openrouter.NewOpenrouterClient(srv.URL, "test"), "")
}

func TestOpenrouterProviderErrors(t *testing.T) {
	tests := []struct {
		name        string
		status      int
		body        string
		expectedErr error
	}{
		{"non-200 status", http.StatusBadGateway, `{"error":{"code":502,"message":"upstream failed"}}`, openrouter.ErrOpenrouterApiCallFail},
		{"error in 200 body", http.StatusOK, `{"error":{"code":429,"message":"rate limited"}}`, openrouter.ErrOpenrouterApiCallFail},
		{"no choices", http.StatusOK, `{"choices":[]}`, ErrEmptyCompletion},
		{"empty content", http.StatusOK, `{"choices":[{"message":{"role":"assistant","content":""}}]}`, ErrEmptyCompletion},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := newTestOpenrouterProvider(t, tt.status, tt.body)
			resp, err := provider.GetCompletionMessage(context.Background(), nil, "", config.ClaudeSonnet5)
			if !errors.Is(err, tt.expectedErr) {
				t.Errorf("expected %v, got %v (resp %q)", tt.expectedErr, err, resp)
			}
		})
	}
}

func TestOpenrouterProviderSuccess(t *testing.T) {
	provider := newTestOpenrouterProvider(t, http.StatusOK, `{"choices":[{"message":{"role":"assistant","content":"hi"}}]}`)
	resp, err := provider.GetCompletionMessage(context.Background(), nil, "", config.ClaudeSonnet5)
	if err != nil || resp != "hi" {
		t.Errorf("expected hi, got %q, %v", resp, err)
	}
}

func TestOpenrouterProviderSendsPromptAsSystemMessage(t *testing.T) {
	var got openrouter.ChatCompletionRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}
		w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"hi"}}]}`))
	}))
	t.Cleanup(srv.Close)
	provider := NewOpenrouterProvider(openrouter.NewOpenrouterClient(srv.URL, "test"), "")
	messages := []CompletionRequestMessage{{Text: "question", Role: "user"}}

	if _, err := provider.GetCompletionMessage(context.Background(), messages, "be brief", config.ClaudeSonnet5); err != nil {
		t.Fatal(err)
	}
	if len(got.Messages) != 2 || got.Messages[0].Role != "system" || got.Messages[0].Content != "be brief" {
		t.Errorf("expected system prompt first, got %+v", got.Messages)
	}

	if _, err := provider.GetCompletionMessage(context.Background(), messages, "", config.ClaudeSonnet5); err != nil {
		t.Fatal(err)
	}
	if len(got.Messages) != 1 || got.Messages[0].Role != "user" {
		t.Errorf("expected no system message for empty prompt, got %+v", got.Messages)
	}
}

func TestOpenrouterProviderSendsTextWithImage(t *testing.T) {
	var got map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&got)
		w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"a cat"}}]}`))
	}))
	t.Cleanup(srv.Close)
	provider := NewOpenrouterProvider(openrouter.NewOpenrouterClient(srv.URL, "test"), "")
	image := "data:image/png;base64,AAAA"
	messages := []CompletionRequestMessage{{Text: "what is it?", Role: "user", ImageUrl: &image}}
	if _, err := provider.GetCompletionMessage(context.Background(), messages, "", config.ClaudeSonnet5); err != nil {
		t.Fatal(err)
	}
	content := got["messages"].([]any)[0].(map[string]any)["content"].([]any)
	if len(content) != 2 || content[0].(map[string]any)["text"] != "what is it?" || content[1].(map[string]any)["type"] != "image_url" {
		t.Errorf("expected text and image parts, got %v", content)
	}
}
