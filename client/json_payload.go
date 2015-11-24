package api_client

import (
	"github.com/gogap/spirit"
)

type JsonPayload struct {
	Id       string          `json:"id"`
	Data     interface{}     `json:"data"`
	Metadata spirit.Metadata `json:"metadata"`
	Errors   []*spirit.Error `json:"errors"`
	Context  spirit.Context  `json:"context"`
}
