package openresponses

// TruncationEnum controls how the service truncates input when it exceeds context window
type TruncationEnum string

const (
	TruncationAuto     TruncationEnum = "auto"
	TruncationDisabled TruncationEnum = "disabled"
)

// MessageRoleEnum represents the role of a message author
type MessageRoleEnum string

const (
	MessageRoleUser      MessageRoleEnum = "user"
	MessageRoleAssistant MessageRoleEnum = "assistant"
	MessageRoleSystem    MessageRoleEnum = "system"
	MessageRoleDeveloper MessageRoleEnum = "developer"
)

// MessageStatusEnum represents the status of a message item
type MessageStatusEnum string

const (
	MessageStatusInProgress MessageStatusEnum = "in_progress"
	MessageStatusCompleted  MessageStatusEnum = "completed"
	MessageStatusIncomplete MessageStatusEnum = "incomplete"
)

// ResponseStatusEnum represents the status of a response
type ResponseStatusEnum string

const (
	ResponseStatusInProgress ResponseStatusEnum = "in_progress"
	ResponseStatusCompleted ResponseStatusEnum = "completed"
	ResponseStatusFailed    ResponseStatusEnum = "failed"
	ResponseStatusIncomplete ResponseStatusEnum = "incomplete"
)

// FunctionCallStatusEnum represents the status of a function call
type FunctionCallStatusEnum string

const (
	FunctionCallStatusInProgress FunctionCallStatusEnum = "in_progress"
	FunctionCallStatusCompleted  FunctionCallStatusEnum = "completed"
	FunctionCallStatusIncomplete FunctionCallStatusEnum = "incomplete"
)

// ToolChoiceValueEnum controls which tool the model should use
type ToolChoiceValueEnum string

const (
	ToolChoiceNone    ToolChoiceValueEnum = "none"
	ToolChoiceAuto    ToolChoiceValueEnum = "auto"
	ToolChoiceRequired ToolChoiceValueEnum = "required"
)

// ImageDetailEnum represents the detail level for image input
type ImageDetailEnum string

const (
	ImageDetailLow  ImageDetailEnum = "low"
	ImageDetailHigh ImageDetailEnum = "high"
	ImageDetailAuto ImageDetailEnum = "auto"
)

// ServiceTierEnum represents the service tier for a request
type ServiceTierEnum string

const (
	ServiceTierAuto     ServiceTierEnum = "auto"
	ServiceTierDefault  ServiceTierEnum = "default"
	ServiceTierFlex     ServiceTierEnum = "flex"
	ServiceTierPriority ServiceTierEnum = "priority"
)

// IncludeEnum represents what to include in the response
type IncludeEnum string

const (
	IncludeReasoningEncryptedContent IncludeEnum = "reasoning.encrypted_content"
	IncludeMessageOutputTextLogprobs IncludeEnum = "message.output_text.logprobs"
)
