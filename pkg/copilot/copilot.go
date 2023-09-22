package copilot

type ChatResponse struct {
	Message string     `json:"message"`
	Steps   []Langcode `json:"steps"`
}

type Langcode struct {
	Language string `json:"language"`
	Code     string `json:"code"`
}
