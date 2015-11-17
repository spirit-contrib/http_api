package http_json_api

import (
	gohttp "net/http"
	"time"

	"github.com/go-martini/martini"
	"github.com/gogap/spirit"
	"github.com/spirit-contrib/http"
)

var (
	receiverURN = "urn:spirit-contrib:receiver:http:json_api"
)

type JsonApiReceiverConfig struct {
	Http http.HTTPReceiverConfig `json:"http"`

	Timeout int `json:"timeout"`
}

type JsonApiReceiver struct {
	*http.HTTPReceiver

	conf JsonApiReceiverConfig
}

var (
	_ spirit.Receiver = new(JsonApiReceiver)
)

func init() {
	spirit.RegisterReceiver(receiverURN, NewJsonApiReceiver)
}

func NewJsonApiReceiver(config spirit.Config) (receiver spirit.Receiver, err error) {
	conf := JsonApiReceiverConfig{}

	if err = config.ToObject(&conf); err != nil {
		return
	}

	jsonApiReceiver := &JsonApiReceiver{
		conf: conf,
	}

	if jsonApiReceiver.HTTPReceiver, err = http.NewHTTPReceiver(conf.Http, jsonApiReceiver.requestHandler); err != nil {
		return
	}

	jsonApiReceiver.Group(conf.Http.Path, func(r martini.Router) {
		r.Post("", jsonApiReceiver.HTTPReceiver.Handler)
		r.Post("/:apiName", jsonApiReceiver.HTTPReceiver.Handler)
		r.Options("", optionHandle)
		r.Options("/:apiName", optionHandle)
	})

	receiver = jsonApiReceiver
	return
}

func optionHandle(w gohttp.ResponseWriter, r *gohttp.Request) {
	return
}

func (p *JsonApiReceiver) requestHandler(
	res gohttp.ResponseWriter,
	req *gohttp.Request,
	deliveryChan <-chan spirit.Delivery,
	done chan<- bool,
) (deliveries []spirit.Delivery, err error) {

	// request to deliveries

	go func(
		count int,
		res gohttp.ResponseWriter,
		req *gohttp.Request,
		deliveryChan <-chan spirit.Delivery,
		done chan<- bool) {

		timeout := time.Duration(p.conf.Timeout) * time.Millisecond

		received := []spirit.Delivery{}

		// get deliveries
		for {
			select {
			case delivery := <-deliveryChan:
				{
					received = append(received, delivery)
					if len(received) == count {
						break
					}
				}
			case <-time.After(timeout):
				{
					break
				}
			}
		}

		// render deliveries to json response
		// normal response: {"code": 0, "message": "", "result": null}
		// error response: {"code": 212, "error_namespace": "xxxx", "message": "something wrong", "result": null}

		// notify the main handler finished
		select {
		case done <- true:
			{
			}
		case <-time.After(time.Second * 3):
			{
			}
		}

	}(len(deliveries), res, req, deliveryChan, done)

	return
}
