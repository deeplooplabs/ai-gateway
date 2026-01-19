package gemini

// GenerateContentRequest represents a Gemini generate content request
type GenerateContentRequest struct {
	Contents         []Content        `json:"contents"`
	Tools            []Tool           `json:"tools,omitempty"`
	GenerationConfig GenerationConfig `json:"generationConfig,omitempty"`
}

// Content represents a single content item with role and parts
type Content struct {
	Role  string `json:"role,omitempty"` // "user", "model", "function"
	Parts []Part `json:"parts"`
}

// Part represents a part of content
type Part struct {
	Text         string                 `json:"text,omitempty"`
	FunctionCall map[string]interface{} `json:"functionCall,omitempty"`
	InlineData   *InlineData            `json:"inlineData,omitempty"`
}

// InlineData represents inline data (e.g., base64 encoded images)
type InlineData struct {
	MIMEType string `json:"mimeType"`
	Data     string `json:"data"`
}

// Tool represents a tool (function) declaration
type Tool struct {
	FunctionDeclarations []FunctionDeclaration `json:"functionDeclarations,omitempty"`
}

// FunctionDeclaration represents a function declaration
type FunctionDeclaration struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
}

// GenerationConfig represents generation configuration
type GenerationConfig struct {
	Temperature     float64  `json:"temperature,omitempty"`
	TopP            float64  `json:"topP,omitempty"`
	TopK            int      `json:"topK,omitempty"`
	MaxOutputTokens int      `json:"maxOutputTokens,omitempty"`
	StopSequences   []string `json:"stopSequences,omitempty"`
}

// GenerateContentResponse represents a Gemini generate content response
type GenerateContentResponse struct {
	Candidates    []Candidate   `json:"candidates"`
	UsageMetadata UsageMetadata `json:"usageMetadata"`
	ModelVersion  string        `json:"modelVersion,omitempty"`
}

// Candidate represents a response candidate
type Candidate struct {
	Content       Content        `json:"content"`
	FinishReason  string         `json:"finishReason"`
	Index         int            `json:"index"`
	SafetyRatings []SafetyRating `json:"safetyRatings,omitempty"`
}

// SafetyRating represents a safety rating
type SafetyRating struct {
	Category    string `json:"category"`
	Probability string `json:"probability"`
}

// UsageMetadata represents usage metadata
type UsageMetadata struct {
	PromptTokenCount     int `json:"promptTokenCount"`
	CandidatesTokenCount int `json:"candidatesTokenCount"`
	TotalTokenCount      int `json:"totalTokenCount"`
}

// EmbedContentRequest represents a Gemini embed content request
type EmbedContentRequest struct {
	Content  Content `json:"content"`
	TaskType string  `json:"taskType,omitempty"`
}

// EmbedContentResponse represents a Gemini embed content response
type EmbedContentResponse struct {
	Embedding EmbeddingValue `json:"embedding"`
}

// EmbeddingValue represents an embedding value
type EmbeddingValue struct {
	Values []float32 `json:"values"`
}
