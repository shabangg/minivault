package types

// Request represents the input prompt structure
// @Description Request payload for text generation
type Request struct {
	// The prompt text to generate from
	// @Example "Tell me a joke"
	Prompt string `json:"prompt" binding:"required" example:"Tell me a joke"`
}

// Response represents the output response structure
// @Description Response payload containing generated text
type Response struct {
	// The generated response text
	// @Example "Why did the chicken cross the road? To get to the other side!"
	Response string `json:"response" example:"Why did the chicken cross the road? To get to the other side!"`
}

// LogEntry represents a single log entry
// @Description Log entry for tracking prompt-response interactions
type LogEntry struct {
	// ISO 8601 timestamp of the interaction
	Timestamp string `json:"timestamp" example:"2024-01-01T12:00:00Z"`
	// The original prompt
	Prompt string `json:"prompt" example:"Tell me a joke"`
	// The generated response
	Response string `json:"response" example:"Why did the chicken cross the road?"`
	// Whether the response was streamed
	Streaming bool `json:"streaming,omitempty" example:"false"`
}
