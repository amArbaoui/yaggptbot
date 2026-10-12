package openrouter

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

var ErrOpenrouterApiCallFail = errors.New("failed to call openrouter api")

const maxLoggedBodyLen = 1024

func truncate(body []byte, maxLen int) []byte {
	if len(body) > maxLen {
		return body[:maxLen]
	}
	return body
}

type Client struct {
	client  *http.Client
	baseUrl string
	apiKey  string
}

func (c *Client) GetChatCompletion(ctx context.Context, request ChatCompletionRequest) (*ChatCompletionResponse, error) {
	var completion ChatCompletionResponse
	if err := c.post(ctx, "/chat/completions", request, &completion); err != nil {
		return nil, err
	}
	if completion.Error != nil {
		return nil, fmt.Errorf("%w: %d %s", ErrOpenrouterApiCallFail, completion.Error.Code, completion.Error.Message)
	}
	return &completion, nil
}

func (c *Client) Transcribe(ctx context.Context, request TranscriptionRequest) (*TranscriptionResponse, error) {
	var transcription TranscriptionResponse
	if err := c.post(ctx, "/audio/transcriptions", request, &transcription); err != nil {
		return nil, err
	}
	if transcription.Error != nil {
		return nil, fmt.Errorf("%w: %d %s", ErrOpenrouterApiCallFail, transcription.Error.Code, transcription.Error.Message)
	}
	return &transcription, nil
}

func (c *Client) post(ctx context.Context, path string, request any, response any) error {
	payload, err := json.Marshal(request)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseUrl+path, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if code := resp.StatusCode; code != http.StatusOK {
		log.Printf("failed to call openrouter %s, received %d: %s", path, code, truncate(body, maxLoggedBodyLen))
		return fmt.Errorf("%w: status %d", ErrOpenrouterApiCallFail, code)
	}
	return json.Unmarshal(body, response)
}

func NewOpenrouterClient(baseUrl string, apiKey string) *Client {
	return &Client{
		client: &http.Client{
			Timeout: time.Second * 300,
		},
		baseUrl: baseUrl,
		apiKey:  apiKey,
	}
}
