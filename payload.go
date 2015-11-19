package http_json_api

import (
	"github.com/gogap/errors"
	"github.com/gogap/spirit"
	"github.com/rs/xid"
	"strings"
)

var (
	ErrHttpResponseGenericError = errors.TN(HttpJsonApiErrNamespace, 100, "")
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

type HttpJsonApiPayload struct {
	id       string
	data     interface{}
	err      error
	metadata spirit.Metadata
	contexts spirit.Contexts
}

func NewHttpJsonApiPayload() (payload *HttpJsonApiPayload) {
	payload = &HttpJsonApiPayload{
		id:       xid.New().String(),
		metadata: make(spirit.Metadata),
		contexts: make(spirit.Contexts),
	}

	return
}

func (p *HttpJsonApiPayload) Id() (id string) {
	return p.id
}

func (p *HttpJsonApiPayload) GetData() (data interface{}, err error) {
	return p.data, nil
}

func (p *HttpJsonApiPayload) SetData(data interface{}) (err error) {
	p.data = data
	return
}

func (p *HttpJsonApiPayload) GetError() (err error) {
	return p.err
}

func (p *HttpJsonApiPayload) SetError(err error) {
	switch e := err.(type) {
	case errors.ErrCode:
		{
			p.err = e
		}
	default:
		{
			p.err = ErrHttpResponseGenericError.New().Append(err)
		}
	}
	return
}

func (p *HttpJsonApiPayload) AppendMetadata(name string, values ...interface{}) (err error) {
	p.metadata[strings.ToUpper(name)] = append(p.metadata[strings.ToUpper(name)], values...)
	return
}

func (p *HttpJsonApiPayload) GetMetadata(name string) (values []interface{}, exist bool) {
	values, exist = p.metadata[strings.ToUpper(name)]
	return
}

func (p *HttpJsonApiPayload) Metadata() (metadata spirit.Metadata) {
	return p.metadata
}

func (p *HttpJsonApiPayload) GetContext(name string) (v interface{}, exist bool) {
	v, exist = p.contexts[strings.ToUpper(name)]
	return
}

func (p *HttpJsonApiPayload) SetContext(name string, v interface{}) (err error) {
	p.contexts[strings.ToUpper(name)] = v
	return
}

func (p *HttpJsonApiPayload) Contexts() (contexts spirit.Contexts) {
	return p.contexts
}

func (p *HttpJsonApiPayload) DeleteContext(name string) (err error) {
	delete(p.contexts, strings.ToUpper(name))
	return
}
