package openai

// Message represents a chat message
type Message struct {
	Role       string     `json:"role"`
	Content    string     `json:"content"`
	Refusal    string     `json:"refusal,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"` // For tool response messages
}

// Choice represents a completion choice
type Choice struct {
	Index        int            `json:"index"`
	Message      Message        `json:"message,omitempty"`
	Delta        *Delta         `json:"delta,omitempty"`
	FinishReason string         `json:"finish_reason"`
	Logprobs     *ChoiceLogprobs `json:"logprobs,omitempty"`
}

// Delta represents streaming message delta
type Delta struct {
	Role      string     `json:"role,omitempty"`
	Content   string     `json:"content,omitempty"`
	Refusal   string     `json:"refusal,omitempty"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

// ChoiceLogprobs contains log probability information for a choice
type ChoiceLogprobs struct {
	Content []LogprobsContent `json:"content,omitempty"`
}

// LogprobsContent contains log probability info for a token
type LogprobsContent struct {
	Token      string       `json:"token"`
	Logprob    float64      `json:"logprob"`
	TopLogprobs []TopLogprob `json:"top_logprobs,omitempty"`
}

// TopLogprob represents a top log probability entry
type TopLogprob struct {
	Token   string  `json:"token"`
	Logprob float64 `json:"logprob"`
}

// Usage represents token usage
type Usage struct {
	PromptTokens              int                      `json:"prompt_tokens"`
	CompletionTokens          int                      `json:"completion_tokens"`
	TotalTokens               int                      `json:"total_tokens"`
	PromptTokensDetails       *PromptTokensDetails     `json:"prompt_tokens_details,omitempty"`
	CompletionTokensDetails   *CompletionTokensDetails `json:"completion_tokens_details,omitempty"`
}

// PromptTokensDetails breaks down prompt token usage
type PromptTokensDetails struct {
	CachedTokens int `json:"cached_tokens,omitempty"`
	AudioTokens  int `json:"audio_tokens,omitempty"`
}

// CompletionTokensDetails breaks down completion token usage
type CompletionTokensDetails struct {
	ReasoningTokens             int `json:"reasoning_tokens,omitempty"`
	AudioTokens                 int `json:"audio_tokens,omitempty"`
	AcceptedPredictionTokens    int `json:"accepted_prediction_tokens,omitempty"`
	RejectedPredictionTokens    int `json:"rejected_prediction_tokens,omitempty"`
}

// StreamOptions controls streaming behavior
type StreamOptions struct {
	IncludeUsage bool `json:"include_usage,omitempty"`
}

// ResponseFormat specifies the output format
type ResponseFormat struct {
	Type       string     `json:"type"` // "text", "json_object", "json_schema"
	JSONSchema *JSONSchema `json:"json_schema,omitempty"`
}

// JSONSchema defines a JSON schema for structured output
type JSONSchema struct {
	Description string       `json:"description,omitempty"`
	Name        string       `json:"name"`
	Schema      map[string]any `json:"schema"`
	Strict      *bool        `json:"strict,omitempty"`
}

// LogProbs configures log probability output
type LogProbsOption struct {
	Enabled    bool  `json:"logprobs,omitempty"`
	TopLogprobs *int  `json:"top_logprobs,omitempty"` // 0-20
}

// ChatCompletionRequest represents a chat completion request
type ChatCompletionRequest struct {
	Model                string             `json:"model"`
	Messages             []Message          `json:"messages"`
	Temperature          *float64           `json:"temperature,omitempty"`
	TopP                 *float64           `json:"top_p,omitempty"`
	N                    *int               `json:"n,omitempty"`
	Stream               bool               `json:"stream,omitempty"`
	MaxTokens            *int               `json:"max_tokens,omitempty"`
	MaxCompletionTokens  *int               `json:"max_completion_tokens,omitempty"`
	Stop                 any                `json:"stop,omitempty"`
	PresencePenalty      *float64           `json:"presence_penalty,omitempty"`
	FrequencyPenalty     *float64           `json:"frequency_penalty,omitempty"`
	Tools                []Tool             `json:"tools,omitempty"`
	ToolChoice           any                `json:"tool_choice,omitempty"`
	StreamOptions        *StreamOptions     `json:"stream_options,omitempty"`
	ResponseFormat       *ResponseFormat    `json:"response_format,omitempty"`
	ServiceTier          string             `json:"service_tier,omitempty"`
	LogProbs             *LogProbsOption    `json:"logprobs,omitempty"`
	Seed                 *int               `json:"seed,omitempty"`
	Metadata             map[string]any     `json:"metadata,omitempty"`
	User                 string             `json:"user,omitempty"`
}

// ChatCompletionResponse represents a chat completion response
type ChatCompletionResponse struct {
	ID                string   `json:"id"`
	Object            string   `json:"object"`
	Created           int64    `json:"created"`
	Model             string   `json:"model"`
	Choices           []Choice `json:"choices"`
	Usage             Usage    `json:"usage"`
	SystemFingerprint string   `json:"system_fingerprint,omitempty"`
	ServiceTier       string   `json:"service_tier,omitempty"`
}

// ChatCompletionStreamResponse represents a streaming chunk
type ChatCompletionStreamResponse struct {
	ID                string   `json:"id"`
	Object            string   `json:"object"`
	Created           int64    `json:"created"`
	Model             string   `json:"model"`
	Choices           []Choice `json:"choices"`
	SystemFingerprint string   `json:"system_fingerprint,omitempty"`
	ServiceTier       string   `json:"service_tier,omitempty"`
	Usage             *Usage   `json:"usage,omitempty"` // Present in final chunk with include_usage
}

// EmbeddingRequest represents an embedding request
type EmbeddingRequest struct {
	Input          any    `json:"input"` // string, []string, or [][]string
	Model          string `json:"model"`
	EncodingFormat string `json:"encoding_format,omitempty"` // "float" or "base64"
	Dimensions     int    `json:"dimensions,omitempty"`      // embedding dimensions
}

// EmbeddingResponse represents an embedding response
type EmbeddingResponse struct {
	Object string      `json:"object"`
	Data   []Embedding `json:"data"`
	Model  string      `json:"model"`
	Usage  Usage       `json:"usage"`
}

// Embedding represents a single embedding vector
type Embedding struct {
	Object    string    `json:"object"`
	Embedding []float32 `json:"embedding"`
	Index     int       `json:"index"`
}

// ImageRequest represents an image generation request
type ImageRequest struct {
	Model   string `json:"model,omitempty"`
	Prompt  string `json:"prompt"`
	N       int    `json:"n,omitempty"`
	Size    string `json:"size,omitempty"`    // "256x256", "512x512", "1024x1024", "1792x1024", "1024x1792"
	Quality string `json:"quality,omitempty"` // "standard" or "hd"
	Style   string `json:"style,omitempty"`   // "vivid" or "natural"
}

// ImageResponse represents an image generation response
type ImageResponse struct {
	Created int64   `json:"created"`
	Data    []Image `json:"data"`
}

// Image represents a generated image
type Image struct {
	URL           string `json:"url,omitempty"`      // For DALL-E 2
	B64JSON       string `json:"b64_json,omitempty"` // For DALL-E 3
	RevisedPrompt string `json:"revised_prompt,omitempty"`
}

// Tool represents a tool that can be called by the model
type Tool struct {
	Type     string             `json:"type"`     // "function"
	Function FunctionDefinition `json:"function"`
}

// FunctionDefinition defines a function tool
type FunctionDefinition struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Parameters  map[string]any `json:"parameters,omitempty"`
}

// ToolCall represents a tool call in a response
type ToolCall struct {
	ID       string `json:"id,omitempty"`
	Type     string `json:"type"` // "function"
	Index    *int   `json:"index,omitempty"` // For streaming tool calls
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

// ToolCallChoice controls tool calling behavior
type ToolCallChoice any // Can be "none", "auto", or a specific object
