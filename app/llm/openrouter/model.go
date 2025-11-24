package openrouter

type Model string

type ChatCompletionRequest struct {
	Model     Model               `json:"model"`
	Reasoning Reasoning           `json:"reasoning"`
	Messages  []CompletionMessage `json:"messages"`
}
type Reasoning struct {
	Effort  string `json:"effort,omitempty"`
	Enabled bool   `json:"enabled,omitempty"`
}
type CompletionMessage struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"`
}

type ChoiceContent struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type CompletionContent struct {
	Type     string    `json:"type"`
	Text     *string   `json:"text,omitempty"`
	ImageUrl *ImageURL `json:"image_url,omitempty"`
}

type ImageURL struct {
	Url string `json:"url"`
}

type ChatCompletionResponse struct {
	ID                *string  `json:"id"`
	Provider          string   `json:"provider"`
	Model             string   `json:"model"`
	Object            string   `json:"object"`
	Created           int64    `json:"created"`
	Choices           []Choice `json:"choices"`
	SystemFingerprint string   `json:"system_fingerprint"`
	Usage             Usage    `json:"usage"`
}

type Choice struct {
	LogProbs           interface{}   `json:"logprobs"`
	FinishReason       string        `json:"finish_reason"`
	NativeFinishReason string        `json:"native_finish_reason"`
	Index              int           `json:"index"`
	Message            ChoiceContent `json:"message"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}
