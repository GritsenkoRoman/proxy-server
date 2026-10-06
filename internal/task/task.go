package task

import "net/http"

type Status string

const (
	StatusNew        Status = "new"
	StatusInProgress Status = "in_process"
	StatusDone       Status = "done"
	StatusError      Status = "error"
)

type CreateRequest struct {
	Method  string            `json:"method" validate:"required"`
	URL     string            `json:"url" validate:"required"`
	Headers map[string]string `json:"headers,omitempty"`
}

type Response struct {
	ID             string              `json:"id"`
	Status         Status              `json:"status"`
	HTTPStatusCode int                 `json:"httpStatusCode,omitempty"`
	Headers        map[string][]string `json:"headers,omitempty"`
	Length         int64               `json:"length,omitempty"`
	Error          string              `json:"error,omitempty"`
}

type Task struct {
	ID      string
	Method  string
	URL     string
	Headers http.Header
}
