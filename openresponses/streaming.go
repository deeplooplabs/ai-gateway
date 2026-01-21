package openresponses

// StreamChunk represents a chunk of streaming data from a provider
type StreamChunk struct {
	Data []byte
	Done bool
}

// StreamingEvent is the base interface for all streaming events
type StreamingEvent interface {
	GetType() string
	GetSequenceNumber() int
	SetSequenceNumber(int)
}

// BaseStreamingEvent provides common fields for streaming events
type BaseStreamingEvent struct {
	Type          string `json:"type"`
	SequenceNumber int    `json:"sequence_number"`
}

func (e *BaseStreamingEvent) GetType() string               { return e.Type }
func (e *BaseStreamingEvent) GetSequenceNumber() int         { return e.SequenceNumber }
func (e *BaseStreamingEvent) SetSequenceNumber(seq int)      { e.SequenceNumber = seq }

// ResponseCreatedEvent is emitted when a response is created
type ResponseCreatedEvent struct {
	BaseStreamingEvent
	Response *Response `json:"response"`
}

// ResponseQueuedEvent is emitted when a response is queued
type ResponseQueuedEvent struct {
	BaseStreamingEvent
	Response *Response `json:"response"`
}

// ResponseInProgressEvent is emitted when a response is in progress
type ResponseInProgressEvent struct {
	BaseStreamingEvent
	Response *Response `json:"response"`
}

// ResponseCompletedEvent is emitted when a response completes successfully
type ResponseCompletedEvent struct {
	BaseStreamingEvent
	Response *Response `json:"response,omitempty"`
}

// ResponseFailedEvent is emitted when a response fails
type ResponseFailedEvent struct {
	BaseStreamingEvent
	ResponseID string  `json:"id,omitempty"`
	Error      *Error `json:"error,omitempty"`
}

// ResponseIncompleteEvent is emitted when a response is incomplete
type ResponseIncompleteEvent struct {
	BaseStreamingEvent
	Response         *Response         `json:"response,omitempty"`
	IncompleteDetails *IncompleteDetails `json:"incomplete_details,omitempty"`
}

// ResponseOutputItemAddedEvent is emitted when a new output item is added
type ResponseOutputItemAddedEvent struct {
	BaseStreamingEvent
	OutputIndex int       `json:"output_index"`
	Item        ItemField `json:"item"`
}

// ResponseOutputItemDoneEvent is emitted when an output item is done
type ResponseOutputItemDoneEvent struct {
	BaseStreamingEvent
	OutputIndex int       `json:"output_index"`
	Item        ItemField `json:"item"`
}

// ResponseContentPartAddedEvent is emitted when a content part is added
type ResponseContentPartAddedEvent struct {
	BaseStreamingEvent
	ItemID        string      `json:"item_id"`
	OutputIndex   int         `json:"output_index"`
	ContentIndex  int         `json:"content_index"`
	Part          ContentPart `json:"part"`
}

// ContentPart represents a part of content
type ContentPart interface{}

// ResponseContentPartDoneEvent is emitted when a content part is done
type ResponseContentPartDoneEvent struct {
	BaseStreamingEvent
	ItemID       string      `json:"item_id"`
	OutputIndex  int         `json:"output_index"`
	ContentIndex int         `json:"content_index"`
	Part         ContentPart `json:"part"`
}

// ResponseOutputTextDeltaEvent is emitted when text delta is received
type ResponseOutputTextDeltaEvent struct {
	BaseStreamingEvent
	ItemID       string   `json:"item_id"`
	OutputIndex  int      `json:"output_index"`
	ContentIndex int      `json:"content_index"`
	Delta        string   `json:"delta"`
	Logprobs     []LogProb `json:"logprobs"`
}

// ResponseOutputTextDoneEvent is emitted when text output is done
type ResponseOutputTextDoneEvent struct {
	BaseStreamingEvent
	ItemID       string   `json:"item_id"`
	OutputIndex  int      `json:"output_index"`
	ContentIndex int      `json:"content_index"`
	Text         string   `json:"text"`
	Logprobs     []LogProb `json:"logprobs"`
}

// ResponseRefusalDeltaEvent is emitted when refusal delta is received
type ResponseRefusalDeltaEvent struct {
	BaseStreamingEvent
	ItemID       string `json:"item_id"`
	OutputIndex  int    `json:"output_index"`
	ContentIndex int    `json:"content_index"`
	Delta        string `json:"delta"`
}

// ResponseRefusalDoneEvent is emitted when refusal is done
type ResponseRefusalDoneEvent struct {
	BaseStreamingEvent
	ItemID       string `json:"item_id"`
	OutputIndex  int    `json:"output_index"`
	ContentIndex int    `json:"content_index"`
	Refusal      string `json:"refusal"`
}

// ResponseReasoningDeltaEvent is emitted when reasoning delta is received
type ResponseReasoningDeltaEvent struct {
	BaseStreamingEvent
	ItemID       string `json:"item_id"`
	OutputIndex  int    `json:"output_index"`
	ContentIndex int    `json:"content_index"`
	Delta        string `json:"delta"`
}

// ResponseReasoningDoneEvent is emitted when reasoning is done
type ResponseReasoningDoneEvent struct {
	BaseStreamingEvent
	ItemID       string             `json:"item_id"`
	OutputIndex  int                `json:"output_index"`
	ContentIndex int                `json:"content_index"`
	Content      []InputTextContentParam `json:"content,omitempty"`
}

// ResponseReasoningSummaryDeltaEvent is emitted when reasoning summary delta is received
type ResponseReasoningSummaryDeltaEvent struct {
	BaseStreamingEvent
	ItemID       string `json:"item_id"`
	OutputIndex  int    `json:"output_index"`
	ContentIndex int    `json:"content_index"`
	Delta        string `json:"delta"`
}

// ResponseReasoningSummaryDoneEvent is emitted when reasoning summary is done
type ResponseReasoningSummaryDoneEvent struct {
	BaseStreamingEvent
	ItemID       string                 `json:"item_id"`
	OutputIndex  int                    `json:"output_index"`
	ContentIndex int                    `json:"content_index"`
	Content      []SummaryTextContent   `json:"content,omitempty"`
}

// ResponseOutputTextAnnotationAddedEvent is emitted when an annotation is added
type ResponseOutputTextAnnotationAddedEvent struct {
	BaseStreamingEvent
	ItemID       string      `json:"item_id"`
	OutputIndex  int         `json:"output_index"`
	ContentIndex int         `json:"content_index"`
	Annotation   Annotation  `json:"annotation"`
}

// ResponseFunctionCallArgumentsDeltaEvent is emitted when function call arguments delta is received
type ResponseFunctionCallArgumentsDeltaEvent struct {
	BaseStreamingEvent
	ItemID    string `json:"item_id"`
	OutputIndex int  `json:"output_index"`
	Delta     string `json:"delta"`
}

// ResponseFunctionCallArgumentsDoneEvent is emitted when function call arguments are done
type ResponseFunctionCallArgumentsDoneEvent struct {
	BaseStreamingEvent
	ItemID      string `json:"item_id"`
	OutputIndex int    `json:"output_index"`
	Arguments   string `json:"arguments"`
}

// ResponseFileSearchCallInProgressEvent is emitted when file search is in progress
type ResponseFileSearchCallInProgressEvent struct {
	BaseStreamingEvent
	ItemID      string `json:"item_id"`
	OutputIndex int    `json:"output_index"`
}

// ResponseFileSearchCallSearchingEvent is emitted when file search is searching
type ResponseFileSearchCallSearchingEvent struct {
	BaseStreamingEvent
	ItemID      string `json:"item_id"`
	OutputIndex int    `json:"output_index"`
	UpdatedAt   int64  `json:"updated_at,omitempty"`
}

// ResponseFileSearchCallCompletedEvent is emitted when file search is completed
type ResponseFileSearchCallCompletedEvent struct {
	BaseStreamingEvent
	ItemID      string              `json:"item_id"`
	OutputIndex int                 `json:"output_index"`
	Result      *FileSearchResult   `json:"result,omitempty"`
}

// FileSearchResult represents file search results
type FileSearchResult struct {
	// File search specific fields
}

// ResponseWebSearchCallInProgressEvent is emitted when web search is in progress
type ResponseWebSearchCallInProgressEvent struct {
	BaseStreamingEvent
	ItemID      string `json:"item_id"`
	OutputIndex int    `json:"output_index"`
}

// ResponseWebSearchCallSearchingEvent is emitted when web search is searching
type ResponseWebSearchCallSearchingEvent struct {
	BaseStreamingEvent
	ItemID      string `json:"item_id"`
	OutputIndex int    `json:"output_index"`
	UpdatedAt   int64  `json:"updated_at,omitempty"`
}

// ResponseWebSearchCallCompletedEvent is emitted when web search is completed
type ResponseWebSearchCallCompletedEvent struct {
	BaseStreamingEvent
	ItemID      string             `json:"item_id"`
	OutputIndex int                `json:"output_index"`
	Result      *WebSearchResult   `json:"result,omitempty"`
}

// WebSearchResult represents web search results
type WebSearchResult struct {
	// Web search specific fields
}

// ErrorStreamingEvent is emitted when an error occurs during streaming
type ErrorStreamingEvent struct {
	BaseStreamingEvent
	Error *Error `json:"error"`
}

// NewResponseCreatedEvent creates a new ResponseCreatedEvent
func NewResponseCreatedEvent(seq int, response *Response) *ResponseCreatedEvent {
	return &ResponseCreatedEvent{
		BaseStreamingEvent: BaseStreamingEvent{
			Type:          "response.created",
			SequenceNumber: seq,
		},
		Response: response,
	}
}

// NewResponseQueuedEvent creates a new ResponseQueuedEvent
func NewResponseQueuedEvent(seq int, response *Response) *ResponseQueuedEvent {
	return &ResponseQueuedEvent{
		BaseStreamingEvent: BaseStreamingEvent{
			Type:          "response.queued",
			SequenceNumber: seq,
		},
		Response: response,
	}
}

// NewResponseInProgressEvent creates a new ResponseInProgressEvent
func NewResponseInProgressEvent(seq int, response *Response) *ResponseInProgressEvent {
	return &ResponseInProgressEvent{
		BaseStreamingEvent: BaseStreamingEvent{
			Type:          "response.in_progress",
			SequenceNumber: seq,
		},
		Response: response,
	}
}

// NewResponseCompletedEvent creates a new ResponseCompletedEvent
func NewResponseCompletedEvent(seq int, response *Response) *ResponseCompletedEvent {
	return &ResponseCompletedEvent{
		BaseStreamingEvent: BaseStreamingEvent{
			Type:          "response.completed",
			SequenceNumber: seq,
		},
		Response: response,
	}
}

// NewResponseFailedEvent creates a new ResponseFailedEvent
func NewResponseFailedEvent(seq int, responseID string, err *Error) *ResponseFailedEvent {
	return &ResponseFailedEvent{
		BaseStreamingEvent: BaseStreamingEvent{
			Type:          "response.failed",
			SequenceNumber: seq,
		},
		ResponseID: responseID,
		Error:      err,
	}
}

// NewResponseOutputItemAddedEvent creates a new ResponseOutputItemAddedEvent
func NewResponseOutputItemAddedEvent(seq int, outputIndex int, item ItemField) *ResponseOutputItemAddedEvent {
	return &ResponseOutputItemAddedEvent{
		BaseStreamingEvent: BaseStreamingEvent{
			Type:          "response.output_item.added",
			SequenceNumber: seq,
		},
		OutputIndex: outputIndex,
		Item:        item,
	}
}

// NewResponseOutputItemDoneEvent creates a new ResponseOutputItemDoneEvent
func NewResponseOutputItemDoneEvent(seq int, outputIndex int, item ItemField) *ResponseOutputItemDoneEvent {
	return &ResponseOutputItemDoneEvent{
		BaseStreamingEvent: BaseStreamingEvent{
			Type:          "response.output_item.done",
			SequenceNumber: seq,
		},
		OutputIndex: outputIndex,
		Item:        item,
	}
}

// NewResponseContentPartAddedEvent creates a new ResponseContentPartAddedEvent
func NewResponseContentPartAddedEvent(seq int, itemID string, outputIndex, contentIndex int, part ContentPart) *ResponseContentPartAddedEvent {
	return &ResponseContentPartAddedEvent{
		BaseStreamingEvent: BaseStreamingEvent{
			Type:          "response.content_part.added",
			SequenceNumber: seq,
		},
		ItemID:       itemID,
		OutputIndex:  outputIndex,
		ContentIndex: contentIndex,
		Part:         part,
	}
}

// NewResponseOutputTextDeltaEvent creates a new ResponseOutputTextDeltaEvent
func NewResponseOutputTextDeltaEvent(seq int, itemID string, outputIndex, contentIndex int, delta string) *ResponseOutputTextDeltaEvent {
	return &ResponseOutputTextDeltaEvent{
		BaseStreamingEvent: BaseStreamingEvent{
			Type:          "response.output_text.delta",
			SequenceNumber: seq,
		},
		ItemID:       itemID,
		OutputIndex:  outputIndex,
		ContentIndex: contentIndex,
		Delta:        delta,
		Logprobs:     []LogProb{},
	}
}

// NewResponseOutputTextDoneEvent creates a new ResponseOutputTextDoneEvent
func NewResponseOutputTextDoneEvent(seq int, itemID string, outputIndex, contentIndex int, text string) *ResponseOutputTextDoneEvent {
	return &ResponseOutputTextDoneEvent{
		BaseStreamingEvent: BaseStreamingEvent{
			Type:          "response.output_text.done",
			SequenceNumber: seq,
		},
		ItemID:       itemID,
		OutputIndex:  outputIndex,
		ContentIndex: contentIndex,
		Text:         text,
		Logprobs:     []LogProb{},
	}
}

// NewErrorStreamingEvent creates a new ErrorStreamingEvent
func NewErrorStreamingEvent(seq int, err *Error) *ErrorStreamingEvent {
	return &ErrorStreamingEvent{
		BaseStreamingEvent: BaseStreamingEvent{
			Type:          "error",
			SequenceNumber: seq,
		},
		Error: err,
	}
}
