package api_client

import (
	"github.com/gogap/spirit"
)

type ErrorCode struct {
	Namespace  string                 `json:"namespace"`
	Code       uint64                 `json:"code"`
	Id         string                 `json:"id"`
	Message    string                 `json:"message"`
	StackTrace string                 `json:"stack_trace"`
	Context    map[string]interface{} `json:"context"`
}

type JsonPayload struct {
	Id       string          `json:"id"`
	Data     interface{}     `json:"data"`
	Metadata spirit.Metadata `json:"metadata"`
	Error    *ErrorCode      `json:"error"`
	Context  spirit.Contexts `json:"context"`
}
