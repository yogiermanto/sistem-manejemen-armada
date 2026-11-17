package model

type WebResponse[T any] struct {
	Data T             `json:"data"`
	Meta *PageMetadata `json:"meta,omitempty"`
}

type PageMetadata struct {
	Page    int  `json:"page"`
	Limit   int  `json:"limit"`
	Count   int  `json:"count"`
	HasMore bool `json:"has_more"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
