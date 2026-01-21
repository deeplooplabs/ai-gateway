package openresponses

import "time"

// CreateRequest is the request body for creating a response
type CreateRequest struct {
	Model              string           `json:"model"`
	Input              InputParam       `json:"input"`
	PreviousResponseID string           `json:"previous_response_id,omitempty"`
	Include            []IncludeEnum    `json:"include,omitempty"`
	Tools              []Tool           `json:"tools,omitempty"`
	ToolChoice         ToolChoiceParam  `json:"tool_choice,omitempty"`
	Metadata           *MetadataParam   `json:"metadata,omitempty"`
	Text               *TextParam       `json:"text,omitempty"`
	Temperature        *float64          `json:"temperature,omitempty"`
	TopP               *float64          `json:"top_p,omitempty"`
	PresencePenalty    *float64          `json:"presence_penalty,omitempty"`
	FrequencyPenalty   *float64          `json:"frequency_penalty,omitempty"`
	ParallelToolCalls  *bool             `json:"parallel_tool_calls,omitempty"`
	Stream             *bool             `json:"stream,omitempty"`
	StreamOptions      *StreamOptionsParam `json:"stream_options,omitempty"`
	Background         *bool             `json:"background,omitempty"`
	MaxOutputTokens    *int              `json:"max_output_tokens,omitempty"`
	MaxToolCalls       *int              `json:"max_tool_calls,omitempty"`
	Reasoning          *ReasoningParam   `json:"reasoning,omitempty"`
	SafetyIdentifier   string           `json:"safety_identifier,omitempty"`
	PromptCacheKey     string           `json:"prompt_cache_key,omitempty"`
	Truncation         TruncationEnum   `json:"truncation,omitempty"`
	Instructions       string           `json:"instructions,omitempty"`
	Store              *bool             `json:"store,omitempty"`
	ServiceTier        ServiceTierEnum  `json:"service_tier,omitempty"`
	TopLogprobs        *int              `json:"top_logprobs,omitempty"`
}

// InputParam represents the input which can be a string or array of items
type InputParam interface{}

// ItemParam represents an input item
type ItemParam interface{}

// UserMessageItemParam represents a user message item
type UserMessageItemParam struct {
	ID      string          `json:"id,omitempty"`
	Type    string          `json:"type"` // "message"
	Role    MessageRoleEnum `json:"role"` // "user"
	Content ContentParam    `json:"content"` // string or []ContentParam
	Status  MessageStatusEnum `json:"status,omitempty"`
}

// AssistantMessageItemParam represents an assistant message item
type AssistantMessageItemParam struct {
	ID      string          `json:"id,omitempty"`
	Type    string          `json:"type"` // "message"
	Role    MessageRoleEnum `json:"role"` // "assistant"
	Content []OutputTextContentParam `json:"content,omitempty"`
	Status  MessageStatusEnum `json:"status,omitempty"`
}

// SystemMessageItemParam represents a system message item
type SystemMessageItemParam struct {
	ID      string          `json:"id,omitempty"`
	Type    string          `json:"type"` // "message"
	Role    MessageRoleEnum `json:"role"` // "system"
	Content ContentParam    `json:"content"` // string or []ContentParam
	Status  MessageStatusEnum `json:"status,omitempty"`
}

// DeveloperMessageItemParam represents a developer message item
type DeveloperMessageItemParam struct {
	ID      string          `json:"id,omitempty"`
	Type    string          `json:"type"` // "message"
	Role    MessageRoleEnum `json:"role"` // "developer"
	Content ContentParam    `json:"content"` // string or []ContentParam
	Status  MessageStatusEnum `json:"status,omitempty"`
}

// FunctionCallItemParam represents a function call item
type FunctionCallItemParam struct {
	ID       string                `json:"id,omitempty"`
	Type     string                `json:"type"` // "function_call"
	CallID   string                `json:"call_id"`
	Name     string                `json:"name"`
	Arguments string               `json:"arguments"`
	Status   FunctionCallStatusEnum `json:"status,omitempty"`
}

// FunctionCallOutputItemParam represents a function call output item
type FunctionCallOutputItemParam struct {
	ID     string `json:"id,omitempty"`
	Type   string `json:"type"` // "function_call_output"
	CallID string `json:"call_id"`
	Output ContentParam `json:"output"` // string or []ContentParam
	Status FunctionCallStatusEnum `json:"status,omitempty"`
}

// ReasoningItemParam represents a reasoning item
type ReasoningItemParam struct {
	ID             string                 `json:"id,omitempty"`
	Type           string                 `json:"type"` // "reasoning"`
	Summary        []ReasoningSummaryContentParam `json:"summary,omitempty"`
	Content        ContentParam           `json:"content,omitempty"`
	EncryptedContent string               `json:"encrypted_content,omitempty"`
}

// ItemReferenceParam represents a reference to an existing item
type ItemReferenceParam struct {
	Type string `json:"type"` // "item_reference"
	ID   string `json:"id"`
}

// OutputTextContentParam represents output text content parameter
type OutputTextContentParam struct {
	Type        string             `json:"type"` // "output_text"
	Text        string             `json:"text,omitempty"`
	Annotations []UrlCitationParam `json:"annotations,omitempty"`
}

// UrlCitationParam represents a URL citation parameter
type UrlCitationParam struct {
	Type      string `json:"type"` // "url_citation"
	StartIndex int    `json:"start_index"`
	EndIndex   int    `json:"end_index"`
	URL        string `json:"url"`
	Title      string `json:"title"`
}

// ReasoningSummaryContentParam represents reasoning summary content
type ReasoningSummaryContentParam struct {
	Type string `json:"type"` // "summary_text"
	Text string `json:"text"`
}

// Tool represents a tool the model can use
type Tool interface{}

// FunctionTool represents a function tool
type FunctionTool struct {
	Type        string          `json:"type"` // "function"
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Parameters  map[string]any  `json:"parameters,omitempty"`
	Strict      *bool           `json:"strict,omitempty"`
}

// ToolChoiceParam controls tool selection
type ToolChoiceParam interface{}

// ToolChoiceValue represents a simple tool choice
type ToolChoiceValue struct {
	Type ToolChoiceValueEnum `json:"type"`
}

// SpecificFunctionParam specifies a particular function to call
type SpecificFunctionParam struct {
	Type string `json:"type"` // "function"
	Name string `json:"name"`
}

// AllowedToolsParam restricts which tools can be used
type AllowedToolsParam struct {
	Type  string               `json:"type"` // "allowed_tools"
	Tools []SpecificToolChoiceParam `json:"tools"`
	Mode  ToolChoiceValueEnum  `json:"mode,omitempty"`
}

// SpecificToolChoiceParam represents a specific tool choice
type SpecificToolChoiceParam interface{}

// MetadataParam represents metadata key-value pairs
type MetadataParam map[string]string

// TextParam controls text output format
type TextParam struct {
	Format TextFormatParam `json:"format,omitempty"`
}

// TextFormatParam represents text format options
type TextFormatParam interface{}

// TextResponseFormat represents basic text response format
type TextResponseFormat struct {
	Type string `json:"type"` // "text"
}

// JsonObjectResponseFormat represents JSON object response format
type JsonObjectResponseFormat struct {
	Type string `json:"type"` // "json_object"
}

// JsonSchemaResponseFormatParam represents JSON schema response format
type JsonSchemaResponseFormatParam struct {
	Type        string                 `json:"type"` // "json_schema"
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Schema      map[string]any         `json:"schema"`
	Strict      *bool                  `json:"strict,omitempty"`
}

// StreamOptionsParam controls streaming behavior
type StreamOptionsParam struct {
	IncludeObfuscation *bool `json:"include_obfuscation,omitempty"`
}

// ReasoningParam controls reasoning behavior
type ReasoningParam struct {
	Effort  ReasoningEffortEnum  `json:"effort,omitempty"`
	Summary ReasoningSummaryEnum `json:"summary,omitempty"`
}

// ReasoningEffortEnum represents reasoning effort levels
type ReasoningEffortEnum string

const (
	ReasoningEffortNone   ReasoningEffortEnum = "none"
	ReasoningEffortLow    ReasoningEffortEnum = "low"
	ReasoningEffortMedium ReasoningEffortEnum = "medium"
	ReasoningEffortHigh   ReasoningEffortEnum = "high"
	ReasoningEffortXHigh  ReasoningEffortEnum = "xhigh"
)

// ReasoningSummaryEnum represents reasoning summary modes
type ReasoningSummaryEnum string

const (
	ReasoningSummaryConcise ReasoningSummaryEnum = "concise"
	ReasoningSummaryDetailed ReasoningSummaryEnum = "detailed"
	ReasoningSummaryAuto    ReasoningSummaryEnum = "auto"
)

// Response represents the API response
type Response struct {
	ID                string             `json:"id"`
	Object            string             `json:"object"` // "response"
	Status            ResponseStatusEnum `json:"status"`
	CreatedAt         int64              `json:"created_at"`
	CompletedAt       *int64             `json:"completed_at,omitempty"`
	Model             string             `json:"model"`
	PreviousResponseID string            `json:"previous_response_id,omitempty"`
	Instructions      string             `json:"instructions,omitempty"`
	Output            []ItemField        `json:"output"`
	Error             *Error             `json:"error,omitempty"`
	Tools             []Tool             `json:"tools,omitempty"`
	ToolChoice        ToolChoiceParam    `json:"tool_choice,omitempty"`
	Truncation        TruncationEnum     `json:"truncation,omitempty"`
	ParallelToolCalls *bool              `json:"parallel_tool_calls,omitempty"`
	Text              *TextField         `json:"text,omitempty"`
	TopP              *float64            `json:"top_p,omitempty"`
	PresencePenalty   *float64            `json:"presence_penalty,omitempty"`
	FrequencyPenalty  *float64            `json:"frequency_penalty,omitempty"`
	TopLogprobs       *int               `json:"top_logprobs,omitempty"`
	Reasoning         *Reasoning         `json:"reasoning,omitempty"`
	Usage             *Usage             `json:"usage,omitempty"`
	MaxOutputTokens   int                `json:"max_output_tokens,omitempty"`
	MaxToolCalls      int                `json:"max_tool_calls,omitempty"`
	Store             *bool              `json:"store,omitempty"`
	Background        *bool              `json:"background,omitempty"`
	ServiceTier       string             `json:"service_tier,omitempty"`
	Metadata          *MetadataParam     `json:"metadata,omitempty"`
	SafetyIdentifier  string             `json:"safety_identifier,omitempty"`
	PromptCacheKey    string             `json:"prompt_cache_key,omitempty"`
}

// ItemField represents an item in the output
type ItemField interface{}

// MessageItem represents a message in the output
type MessageItem struct {
	ID      string             `json:"id"`
	Type    string             `json:"type"` // "message"
	Status  MessageStatusEnum  `json:"status"`
	Role    MessageRoleEnum    `json:"role"`
	Content []OutputTextContent `json:"content"`
}

// FunctionCallItem represents a function call in the output
type FunctionCallItem struct {
	ID       string                `json:"id"`
	Type     string                `json:"type"` // "function_call"
	Status   FunctionCallStatusEnum `json:"status"`
	CallID   string                `json:"call_id"`
	Name     string                `json:"name"`
	Arguments string               `json:"arguments"`
}

// FunctionCallOutputItem represents a function call output in the response
type FunctionCallOutputItem struct {
	ID     string `json:"id"`
	Type   string `json:"type"` // "function_call_output"
	CallID string `json:"call_id"`
	Output ContentParam `json:"output"`
	Status FunctionCallStatusEnum `json:"status"`
}

// ReasoningItem represents a reasoning item in the output
type ReasoningItem struct {
	ID              string              `json:"id"`
	Type            string              `json:"type"` // "reasoning"`
	Status          string              `json:"status"`
	Content         []InputTextContentParam `json:"content,omitempty"`
	Summary         []SummaryTextContent `json:"summary,omitempty"`
	EncryptedContent string             `json:"encrypted_content,omitempty"`
}

// TextField represents text output configuration
type TextField struct {
	Format TextFormat `json:"format,omitempty"`
}

// TextFormat represents text format in the response
type TextFormat interface{}

// Reasoning represents reasoning in the response
type Reasoning struct {
	Effort  ReasoningEffortEnum  `json:"effort,omitempty"`
	Summary ReasoningSummaryEnum `json:"summary,omitempty"`
}

// Usage represents token usage statistics
type Usage struct {
	InputTokens        int               `json:"input_tokens"`
	OutputTokens       int               `json:"output_tokens"`
	TotalTokens        int               `json:"total_tokens"`
	InputTokensDetails *InputTokensDetails `json:"input_tokens_details,omitempty"`
	OutputTokensDetails *OutputTokensDetails `json:"output_tokens_details,omitempty"`
}

// InputTokensDetails breaks down input token usage
type InputTokensDetails struct {
	CachedTokens int `json:"cached_tokens,omitempty"`
}

// OutputTokensDetails breaks down output token usage
type OutputTokensDetails struct {
	ReasoningTokens int `json:"reasoning_tokens,omitempty"`
}

// Error represents an error response
type Error struct {
	Type    string `json:"type"`
	Code    string `json:"code,omitempty"`
	Message string `json:"message"`
	Param   string `json:"param,omitempty"`
}

// IncompleteDetails describes why a response was incomplete
type IncompleteDetails struct {
	Reason string `json:"reason"`
}

// NewResponse creates a new Response with initialized fields
func NewResponse(id string, model string) *Response {
	return &Response{
		ID:        id,
		Object:    "response",
		Status:    ResponseStatusInProgress,
		CreatedAt: time.Now().Unix(),
		Model:     model,
		Output:    []ItemField{},
	}
}

// NewError creates a new Error
func NewError(errorType, code, message, param string) *Error {
	return &Error{
		Type:    errorType,
		Code:    code,
		Message: message,
		Param:   param,
	}
}
