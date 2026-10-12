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
	url := fmt.Sprintf("%s/chat/completions", c.baseUrl)
	payload, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	response, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	if code := response.StatusCode; code != http.StatusOK {
		log.Printf("failed to call openrouter, received %d: %s", code, truncate(body, maxLoggedBodyLen))
		return nil, fmt.Errorf("%w: status %d", ErrOpenrouterApiCallFail, code)
	}

	if err := json.Unmarshal(body, &completion); err != nil {
		return nil, err
	}
	if completion.Error != nil {
		return nil, fmt.Errorf("%w: %d %s", ErrOpenrouterApiCallFail, completion.Error.Code, completion.Error.Message)
	}
	return &completion, nil

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
