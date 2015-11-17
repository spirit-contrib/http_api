package http_json_api

import (
	"time"

	"github.com/gogap/spirit"
)

type HttpJsonApiDelivery struct {
	id        string
	urn       string
	sessionId string
	payload   *HttpJsonApiPayload
	labels    spirit.Labels
	timestamp time.Time
}

func (p *HttpJsonApiDelivery) Id() string {
	return p.id
}

func (p *HttpJsonApiDelivery) URN() string {
	return p.urn
}

func (p *HttpJsonApiDelivery) SessionId() string {
	return p.sessionId
}

func (p *HttpJsonApiDelivery) Labels() spirit.Labels {
	return p.labels
}

func (p *HttpJsonApiDelivery) Payload() spirit.Payload {
	return p.payload
}

func (p *HttpJsonApiDelivery) Validate() (err error) {
	return
}

func (p *HttpJsonApiDelivery) Timestamp() time.Time {
	return p.timestamp
}
