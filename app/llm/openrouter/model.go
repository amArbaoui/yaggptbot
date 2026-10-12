package openrouter

type Model string

type ChatCompletionRequest struct {
	Model     Model               `json:"model"`
	Reasoning Reasoning           `json:"reasoning"`
	Messages  []CompletionMessage `json:"messages"`
	Tools     []Tool              `json:"tools,omitempty"`
	Plugins   []Plugin            `json:"plugins,omitempty"`
	MaxTokens int                 `json:"max_tokens,omitempty"`
}

// Plugin configures OpenRouter request plugins, e.g. the file parser for PDFs
type Plugin struct {
	ID  string            `json:"id"`
	PDF *PDFPluginOptions `json:"pdf,omitempty"`
}

const PluginFileParser = "file-parser"

type PDFPluginOptions struct {
	Engine string `json:"engine"`
}

// Tool is an OpenRouter server tool, executed by OpenRouter when the model calls it
type Tool struct {
	Type       string `json:"type"`
	Parameters any    `json:"parameters,omitempty"`
}

const (
	ToolWebSearch = "openrouter:web_search"
	ToolWebFetch  = "openrouter:web_fetch"
)

type WebSearchParameters struct {
	Engine     string `json:"engine,omitempty"`
	MaxResults int    `json:"max_results,omitempty"`
	MaxUses    int    `json:"max_uses,omitempty"`
}

type WebFetchParameters struct {
	Engine           string `json:"engine,omitempty"`
	MaxUses          int    `json:"max_uses,omitempty"`
	MaxContentTokens int    `json:"max_content_tokens,omitempty"`
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
	Role        string       `json:"role"`
	Content     string       `json:"content"`
	Annotations []Annotation `json:"annotations,omitempty"`
}

type Annotation struct {
	Type        string          `json:"type"`
	URLCitation *URLCitation    `json:"url_citation,omitempty"`
	File        *FileAnnotation `json:"file,omitempty"`
}

// FileAnnotation holds the parsed content of a file sent in the request
type FileAnnotation struct {
	Hash    string              `json:"hash"`
	Name    string              `json:"name"`
	Content []CompletionContent `json:"content"`
}

type URLCitation struct {
	URL   string `json:"url"`
	Title string `json:"title"`
}

type CompletionContent struct {
	Type     string    `json:"type"`
	Text     *string   `json:"text,omitempty"`
	ImageUrl *ImageURL `json:"image_url,omitempty"`
	File     *File     `json:"file,omitempty"`
}

type File struct {
	Filename string `json:"filename"`
	FileData string `json:"file_data"`
}

type ImageURL struct {
	Url string `json:"url"`
}

type ChatCompletionResponse struct {
	ID                *string   `json:"id"`
	Provider          string    `json:"provider"`
	Model             string    `json:"model"`
	Object            string    `json:"object"`
	Created           int64     `json:"created"`
	Choices           []Choice  `json:"choices"`
	SystemFingerprint string    `json:"system_fingerprint"`
	Usage             Usage     `json:"usage"`
	Error             *ApiError `json:"error,omitempty"`
}

type ApiError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type Choice struct {
	LogProbs           interface{}   `json:"logprobs"`
	FinishReason       string        `json:"finish_reason"`
	NativeFinishReason string        `json:"native_finish_reason"`
	Index              int           `json:"index"`
	Message            ChoiceContent `json:"message"`
}

type Usage struct {
	PromptTokens     int            `json:"prompt_tokens"`
	CompletionTokens int            `json:"completion_tokens"`
	TotalTokens      int            `json:"total_tokens"`
	ServerToolUse    *ServerToolUse `json:"server_tool_use,omitempty"`
}

type ServerToolUse struct {
	WebSearchRequests int `json:"web_search_requests"`
}

type TranscriptionRequest struct {
	Model      Model      `json:"model"`
	InputAudio InputAudio `json:"input_audio"`
}

type InputAudio struct {
	Data   string `json:"data"`
	Format string `json:"format"`
}

type TranscriptionResponse struct {
	Text  string    `json:"text"`
	Error *ApiError `json:"error,omitempty"`
}
