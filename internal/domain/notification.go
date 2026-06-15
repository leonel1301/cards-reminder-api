package domain

type PushNotification struct {
	Title string
	Body  string
	Data  map[string]string
}

type PushSendResult struct {
	SuccessCount  int      `json:"success_count"`
	FailureCount  int      `json:"failure_count"`
	InvalidTokens []string `json:"invalid_tokens,omitempty"`
}
