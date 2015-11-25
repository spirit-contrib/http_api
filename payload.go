package http_json_api

import (
	"encoding/json"
	"github.com/gogap/errors"
	"github.com/gogap/spirit"
	"github.com/rs/xid"
	"strings"
)

var (
	ErrHttpResponseGenericError = errors.TN(HttpJsonApiErrNamespace, 100, "")
)

type JsonPayload struct {
	Id      string          `json:"id"`
	Data    interface{}     `json:"data"`
	Errors  []*spirit.Error `json:"errors"`
	Context spirit.Map      `json:"context"`
}

type HttpJsonApiPayload struct {
	id      string
	data    interface{}
	errs    []*spirit.Error
	context spirit.Map
}

func NewHttpJsonApiPayload() (payload *HttpJsonApiPayload) {
	payload = &HttpJsonApiPayload{
		id:      xid.New().String(),
		context: make(spirit.Map),
	}

	return
}

func (p *HttpJsonApiPayload) Id() (id string) {
	return p.id
}

func (p *HttpJsonApiPayload) GetData() (data interface{}, err error) {
	return p.data, nil
}

func (p *HttpJsonApiPayload) DataToObject(v interface{}) (err error) {
	switch d := p.data.(type) {
	case string:
		{

			if err = json.Unmarshal([]byte(d), v); err != nil {
				var b []byte
				if b, err = json.Marshal(d); err != nil {
					return
				}

				if err = json.Unmarshal(b, v); err != nil {
					return
				}
			}
		}
	case []byte:
		{
			if err = json.Unmarshal(d, v); err != nil {
				var b []byte
				if b, err = json.Marshal(string(d)); err != nil {
					return
				}

				if err = json.Unmarshal(b, v); err != nil {
					return
				}
			}
		}
	default:
		{
			var b []byte
			if b, err = json.Marshal(p.data); err != nil {
				return
			}

			if err = json.Unmarshal(b, v); err != nil {
				return
			}
		}
	}
	return
}

func (p *HttpJsonApiPayload) SetData(data interface{}) (err error) {
	p.data = data
	return
}

func (p *HttpJsonApiPayload) Errors() (err []*spirit.Error) {
	return p.errs
}

func (p *HttpJsonApiPayload) AppendError(err ...*spirit.Error) {
	p.errs = append(p.errs, err...)
}

func (p *HttpJsonApiPayload) LastError() (err *spirit.Error) {
	if len(p.errs) > 0 {
		err = p.errs[len(p.errs)-1]
	}
	return
}

func (p *HttpJsonApiPayload) ClearErrors() {
	p.errs = nil
	return
}

func (p *HttpJsonApiPayload) GetContext(name string) (v interface{}, exist bool) {
	v, exist = p.context[strings.ToUpper(name)]
	return
}

func (p *HttpJsonApiPayload) SetContext(name string, v interface{}) (err error) {
	p.context[strings.ToUpper(name)] = v
	return
}

func (p *HttpJsonApiPayload) Context() (context spirit.Map) {
	return p.context
}

func (p *HttpJsonApiPayload) DeleteContext(name string) (err error) {
	delete(p.context, strings.ToUpper(name))
	return
}
