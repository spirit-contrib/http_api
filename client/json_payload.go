package api_client

import (
	"github.com/gogap/spirit"
)

type JsonPayload struct {
	Id      string          `json:"id"`
	Data    interface{}     `json:"data"`
	Errors  []*spirit.Error `json:"errors"`
	Context spirit.Map      `json:"context"`
}
