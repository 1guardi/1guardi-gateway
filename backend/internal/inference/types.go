package inference

import "encoding/json"

type analyzeRequest struct {
	Text string `json:"text"`
}

type analyzeResponse struct {
	Analyzer string          `json:"analyzer"`
	Result   json.RawMessage `json:"result"`
}

// Result holds the raw response from a named analyzer.
// The Raw field shape is analyzer-specific; callers parse it themselves.
type Result struct {
	Analyzer string
	Raw      json.RawMessage
	Cached   bool
}
