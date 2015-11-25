package http_json_api

import (
	"sync"
	"time"

	"github.com/gogap/spirit"
)

type HttpJsonApiDelivery struct {
	id        string
	urn       string
	sessionId string
	payload   *HttpJsonApiPayload
	metadata  spirit.Map
	labels    spirit.Labels
	timestamp time.Time

	labelsLocker sync.Mutex
}

func (p *HttpJsonApiDelivery) Id() string {
	return p.id
}

func (p *HttpJsonApiDelivery) URN() string {
	return p.urn
}

func (p *HttpJsonApiDelivery) SetURN(urn string) (err error) {
	p.urn = urn
	return
}

func (p *HttpJsonApiDelivery) SessionId() string {
	return p.sessionId
}

func (p *HttpJsonApiDelivery) Labels() spirit.Labels {
	return p.labels
}

func (p *HttpJsonApiDelivery) SetLabel(label string, value string) (err error) {
	p.labelsLocker.Lock()
	p.labelsLocker.Unlock()

	if p.labels == nil {
		p.labels = make(spirit.Labels)
		return
	}

	p.labels[label] = value

	return
}
func (p *HttpJsonApiDelivery) SetLabels(labels spirit.Labels) (err error) {
	p.labelsLocker.Lock()
	p.labelsLocker.Unlock()

	p.labels = labels
	return
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

func (p *HttpJsonApiDelivery) GetMetadata(name string) (v interface{}, exist bool) {
	if p.metadata == nil {
		return
	}

	v, exist = p.metadata[name]

	return
}

func (p *HttpJsonApiDelivery) SetMetadata(name string, v interface{}) (err error) {
	p.metadata[name] = v
	return
}

func (p *HttpJsonApiDelivery) Metadata() (metadata spirit.Map) {
	metadata = p.metadata
	return
}

func (p *HttpJsonApiDelivery) DeleteMetadata(name string) (err error) {
	if p.metadata != nil {
		delete(p.metadata, name)
	}
	return
}
