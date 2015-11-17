package http_json_api

import (
	"github.com/gogap/spirit"
)

type HttpJsonApiPayload struct {
	id       string
	data     interface{}
	err      error
	metadata spirit.Metadata
	context  spirit.Contexts
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
	p.err = err
	return
}

func (p *HttpJsonApiPayload) AppendMetadata(name string, values ...interface{}) (err error) {
	return
}

func (p *HttpJsonApiPayload) GetMetadata(name string) (values []interface{}, exist bool) {
	return
}

func (p *HttpJsonApiPayload) Metadata() (metadata spirit.Metadata) {
	return
}

func (p *HttpJsonApiPayload) GetContext(name string) (v interface{}, exist bool) {
	return
}

func (p *HttpJsonApiPayload) SetContext(name string, v interface{}) (err error) {
	return
}

func (p *HttpJsonApiPayload) Contexts() (contexts spirit.Contexts) {
	return
}

func (p *HttpJsonApiPayload) DeleteContext(name string) (err error) {
	return
}
