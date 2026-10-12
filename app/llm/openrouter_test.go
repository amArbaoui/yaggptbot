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
			resp, err := provider.GetCompletionMessage(context.Background(), nil, "", config.ClaudeSonnet5, CompletionOptions{})
			if !errors.Is(err, tt.expectedErr) {
				t.Errorf("expected %v, got %v (resp %q)", tt.expectedErr, err, resp)
			}
		})
	}
}

func TestOpenrouterProviderSuccess(t *testing.T) {
	provider := newTestOpenrouterProvider(t, http.StatusOK, `{"choices":[{"message":{"role":"assistant","content":"hi"}}]}`)
	resp, err := provider.GetCompletionMessage(context.Background(), nil, "", config.ClaudeSonnet5, CompletionOptions{})
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

	if _, err := provider.GetCompletionMessage(context.Background(), messages, "be brief", config.ClaudeSonnet5, CompletionOptions{}); err != nil {
		t.Fatal(err)
	}
	if len(got.Messages) != 2 || got.Messages[0].Role != "system" || got.Messages[0].Content != "be brief" {
		t.Errorf("expected system prompt first, got %+v", got.Messages)
	}

	if _, err := provider.GetCompletionMessage(context.Background(), messages, "", config.ClaudeSonnet5, CompletionOptions{}); err != nil {
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
	if _, err := provider.GetCompletionMessage(context.Background(), messages, "", config.ClaudeSonnet5, CompletionOptions{}); err != nil {
		t.Fatal(err)
	}
	content := got["messages"].([]any)[0].(map[string]any)["content"].([]any)
	if len(content) != 2 || content[0].(map[string]any)["text"] != "what is it?" || content[1].(map[string]any)["type"] != "image_url" {
		t.Errorf("expected text and image parts, got %v", content)
	}
}

func TestOpenrouterProviderWebTools(t *testing.T) {
	var got map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = nil
		json.NewDecoder(r.Body).Decode(&got)
		w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"hi"}}]}`))
	}))
	t.Cleanup(srv.Close)
	provider := NewOpenrouterProvider(openrouter.NewOpenrouterClient(srv.URL, "test"), "")
	messages := []CompletionRequestMessage{{Text: "news?", Role: "user"}}

	if _, err := provider.GetCompletionMessage(context.Background(), messages, "", config.ClaudeSonnet5, CompletionOptions{WebTools: true}); err != nil {
		t.Fatal(err)
	}
	tools, _ := got["tools"].([]any)
	if len(tools) != 2 {
		t.Fatalf("expected search and fetch tools, got %v", got["tools"])
	}
	search := tools[0].(map[string]any)
	params := search["parameters"].(map[string]any)
	if search["type"] != openrouter.ToolWebSearch || params["engine"] != config.WebSearchEngine || params["max_results"] != float64(config.WebSearchMaxResults) {
		t.Errorf("unexpected search tool %v", search)
	}
	if tools[1].(map[string]any)["type"] != openrouter.ToolWebFetch {
		t.Errorf("unexpected fetch tool %v", tools[1])
	}

	if _, err := provider.GetCompletionMessage(context.Background(), messages, "", config.ClaudeSonnet5, CompletionOptions{}); err != nil {
		t.Fatal(err)
	}
	if _, ok := got["tools"]; ok {
		t.Errorf("tools should not be sent without WebTools, got %v", got["tools"])
	}
}

func TestOpenrouterProviderAppendsSources(t *testing.T) {
	provider := newTestOpenrouterProvider(t, http.StatusOK, `{"choices":[{"message":{"role":"assistant",
		"content":"Go 1.30 is out, see [notes](https://go.dev/doc/go1.30).",
		"annotations":[
			{"type":"url_citation","url_citation":{"url":"https://go.dev/doc/go1.30","title":"Release notes"}},
			{"type":"url_citation","url_citation":{"url":"https://go.dev/blog/go1.30","title":"Go 1.30 [blog]"}},
			{"type":"url_citation","url_citation":{"url":"https://go.dev/blog/go1.30","title":"duplicate"}},
			{"type":"url_citation","url_citation":{"url":"https://example.com/x"}}
		]}}],
		"usage":{"server_tool_use":{"web_search_requests":1}}}`)
	got, err := provider.GetCompletionMessage(context.Background(), nil, "", config.ClaudeSonnet5, CompletionOptions{WebTools: true})
	if err != nil {
		t.Fatal(err)
	}
	expected := "Go 1.30 is out, see [notes](https://go.dev/doc/go1.30).\n\n**Sources:**" +
		"\n1. [Go 1.30 (blog)](<https://go.dev/blog/go1.30>)" +
		"\n2. [https://example.com/x](<https://example.com/x>)"
	if got != expected {
		t.Errorf("expected\n%s\ngot\n%s", expected, got)
	}
}
