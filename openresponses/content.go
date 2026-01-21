package openresponses

// InputTextContentParam represents a text input to the model
type InputTextContentParam struct {
	Type string `json:"type"` // "input_text"
	Text string `json:"text"`
}

// InputImageContentParam represents an image input to the model
type InputImageContentParam struct {
	Type     string         `json:"type"` // "input_image"
	ImageURL string         `json:"image_url"`
	Detail   ImageDetailEnum `json:"detail,omitempty"`
}

// InputFileContentParam represents a file input to the model
type InputFileContentParam struct {
	Type     string `json:"type"` // "input_file"
	Filename string `json:"filename,omitempty"`
	FileData string `json:"file_data,omitempty"` // base64 encoded
	FileURL  string `json:"file_url,omitempty"`
}

// InputVideoContentParam represents a video input to the model
type InputVideoContentParam struct {
	Type     string `json:"type"` // "input_video"
	VideoURL string `json:"video_url"`
}

// ContentParam represents a union of input content types
type ContentParam interface{}

// OutputTextContent represents a text output from the model
type OutputTextContent struct {
	Type        string      `json:"type"` // "output_text"
	Text        string      `json:"text"`
	Annotations []Annotation `json:"annotations"` // Required, empty array when not used
	Logprobs    []LogProb    `json:"logprobs"`    // Required, empty array when not used
}

// RefusalContent represents a refusal from the model
type RefusalContent struct {
	Type    string `json:"type"` // "refusal"
	Refusal string `json:"refusal"`
}

// SummaryTextContent represents a summary of reasoning
type SummaryTextContent struct {
	Type string `json:"type"` // "summary_text"
	Text string `json:"text"`
}

// Annotation represents an annotation on output text
type Annotation interface{}

// UrlCitationBody represents a URL citation
type UrlCitationBody struct {
	Type      string `json:"type"` // "url_citation"
	URL       string `json:"url"`
	StartIndex int    `json:"start_index"`
	EndIndex   int    `json:"end_index"`
	Title      string `json:"title"`
}

// LogProb represents log probability information
type LogProb struct {
	Token       string     `json:"token"`
	Logprob     float64    `json:"logprob"`
	Bytes       []int      `json:"bytes,omitempty"`
	TopLogprobs []TopLogProb `json:"top_logprobs,omitempty"`
}

// TopLogProb represents top log probability
type TopLogProb struct {
	Token   string  `json:"token"`
	Logprob float64 `json:"logprob"`
	Bytes   []int   `json:"bytes,omitempty"`
}
