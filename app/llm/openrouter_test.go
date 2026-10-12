package llm

import (
	"amArbaoui/yaggptbot/app/config"
	"amArbaoui/yaggptbot/app/llm/openrouter"
	"context"
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
