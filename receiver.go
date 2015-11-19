package http_json_api

import (
	"encoding/json"
	"github.com/gogap/errors"
	"github.com/rs/xid"
	"io/ioutil"
	gohttp "net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-martini/martini"
	"github.com/gogap/spirit"
	"github.com/spirit-contrib/http"
)

var (
	receiverURN = "urn:spirit-contrib:receiver:http:json_api"
)

type JsonApiReceiver struct {
	*http.HTTPReceiver

	conf JsonApiReceiverConfig

	responseRenderer *APIResponseRenderer
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

	conf.initial()

	jsonApiReceiver := &JsonApiReceiver{
		conf: conf,
	}

	if jsonApiReceiver.HTTPReceiver, err = http.NewHTTPReceiver(conf.Http, jsonApiReceiver.requestHandler); err != nil {
		return
	}

	jsonApiReceiver.responseRenderer = NewAPIResponseRenderer()

	path := strings.TrimRight(conf.Path, "/")
	jsonApiReceiver.Group(path, func(r martini.Router) {
		r.Post("", jsonApiReceiver.HTTPReceiver.Handler)
		r.Post("/:apiName", jsonApiReceiver.HTTPReceiver.Handler)
		r.Options("", jsonApiReceiver.optionHandle)
		r.Options("/:apiName", jsonApiReceiver.optionHandle)
	})

	jsonApiReceiver.Group("/", func(r martini.Router) {
		r.Get("ping", func() string {
			return "pong"
		})
	})

	receiver = jsonApiReceiver
	return
}

func (p *JsonApiReceiver) optionHandle(w gohttp.ResponseWriter, r *gohttp.Request) {
	if r.Method == "OPTIONS" {
		p.writeAccessHeaders(w, r)
		p.writeBasicHeaders(w, r)
		w.Write([]byte(""))
	}
}

func (p *JsonApiReceiver) requestHandler(
	res gohttp.ResponseWriter,
	req *gohttp.Request,
	deliveryChan <-chan spirit.Delivery,
	done chan<- bool,
) (deliveries []spirit.Delivery, err error) {

	var apiIds map[string]string

	// request to deliveries
	if deliveries, apiIds, err = p.toDeliveries(req); err != nil {

		var apiResponse APIResponse

		switch errCode := err.(type) {
		case errors.ErrCode:
			{
				apiResponse = APIResponse{
					Code:           errCode.Code(),
					ErrorId:        errCode.Id(),
					ErrorNamespace: errCode.Namespace(),
					Message:        errCode.Error(),
					Result:         nil,
				}
			}
		default:
			e := ErrApiGenericError.New().Append(err)
			apiResponse = APIResponse{
				Code:           e.Code(),
				ErrorId:        e.Id(),
				ErrorNamespace: e.Namespace(),
				Message:        e.Error(),
				Result:         nil,
			}
		}

		if data, e := json.Marshal(apiResponse); e != nil {
			spirit.Logger().WithField("event", "to deliveries").Println(err)
		} else {
			p.writeResponse(data, res, req)
		}

		return
	}

	go func(
		count int,
		apiIds map[string]string,
		res gohttp.ResponseWriter,
		req *gohttp.Request,
		deliveryChan <-chan spirit.Delivery,
		done chan<- bool) {

		defer func() {
			// notify the main handler finished
			select {
			case done <- true:
				{
				}
			case <-time.After(time.Second * 3):
				{
				}
			}
		}()

		// get timeout duration
		timeout := time.Duration(p.conf.Timeout) * time.Millisecond

		if strTimeout := req.Header.Get(p.conf.HeaderDefines.TimeoutHeader); strTimeout != "" {
			if i, e := strconv.Atoi(strTimeout); e == nil {
				timeout = time.Duration(i) * time.Millisecond
			}
		}

		if timeout <= 0 {
			timeout = DefaultTimeout
		}

		apiResponse := map[string]APIResponse{}

		i := count
		// get deliveries
	label_timeout_or_finished:
		for i > 0 {
			select {
			case delivery := <-deliveryChan:
				{
					if api, exist := apiIds[delivery.Id()]; !exist {
						spirit.Logger().
							WithField("api", api).
							WithField("delivery_id", delivery.Id()).
							Errorln("api not exist in request while delivery response")

					} else {
						apiResponse[api] = p.deliveryToApiResponse(delivery)
					}

					i = i - 1
				}
			case <-time.After(timeout):
				{
					break label_timeout_or_finished
				}
			}
		}

		// request timeout
		for _, api := range apiIds {
			if _, exist := apiResponse[api]; !exist {
				errCode := ErrRequestTimeout.New()
				apiResponse[api] = APIResponse{
					Code:           errCode.Code(),
					ErrorId:        errCode.Id(),
					ErrorNamespace: errCode.Namespace(),
					Message:        errCode.Error(),
					Result:         nil,
				}
			}
		}

		// render deliveries to json response
		// normal response: {"code": 0, "message": "", "result": null}
		// error response: {"code": 212, "error_namespace": "xxxx", "message": "something wrong", "result": null}

		isMultiCall := req.Header.Get(p.conf.HeaderDefines.MultiCallHeader) == "1" ||
			req.Header.Get(p.conf.HeaderDefines.MultiCallHeader) == "on"

		if renderedData, e := p.responseRenderer.Render(isMultiCall, apiResponse); e != nil {
			err := ErrRenderApiDataFailed.New(errors.Params{"err": e})
			resp := APIResponse{
				Code:           err.Code(),
				ErrorId:        err.Id(),
				ErrorNamespace: err.Namespace(),
				Message:        err.Error(),
				Result:         nil,
			}

			if errRespData, e := json.Marshal(resp); e != nil {
				strInternalErr := `{"code": 500, "message": "api server internal error", "result": null}`
				p.writeResponseWithStatusCode([]byte(strInternalErr), res, req, gohttp.StatusInternalServerError)
			} else {
				p.writeResponse(errRespData, res, req)
			}
		} else {
			p.writeResponse(renderedData, res, req)
		}
		return
	}(len(deliveries), apiIds, res, req, deliveryChan, done)

	return
}

func (p *JsonApiReceiver) toDeliveries(req *gohttp.Request) (deliveries []spirit.Delivery, apiIds map[string]string, err error) {
	isMultiCall := req.Header.Get(p.conf.HeaderDefines.MultiCallHeader) == "1" ||
		req.Header.Get(p.conf.HeaderDefines.MultiCallHeader) == "on"

	isForwarded := req.Header.Get(HeaderForwardedPayload) == "1" ||
		req.Header.Get(HeaderForwardedPayload) == "on"

	if isMultiCall == true && isForwarded == true {
		err = ErrNotSupportMultiCallForward.New()
		return
	}

	idMapping := make(map[string]string)

	var body []byte
	if body, err = ioutil.ReadAll(req.Body); err != nil {
		return
	}

	var apiDatas map[string]interface{} = make(map[string]interface{})

	if isMultiCall {
		if err = json.Unmarshal(body, &apiDatas); err != nil {
			return
		}
	} else {

		apiName := req.Header.Get(p.conf.HeaderDefines.ApiHeader)

		if apiName == "" {
			if p.conf.Path != req.RequestURI {
				apiName = strings.TrimPrefix(req.RequestURI, p.conf.Path)
				apiName = strings.TrimRight(apiName, "/")
			}
		}

		if apiName == "" {
			err = ErrApiNameIsEmpty
			return
		}

		if isForwarded {
			apiData := JsonPayload{}
			if json.Unmarshal(body, &apiData); err != nil {
				return
			}
			apiDatas[apiName] = apiData
		} else {
			var apiData map[string]interface{}
			if json.Unmarshal(body, &apiData); err != nil {
				return
			}
			apiDatas[apiName] = apiData
		}

	}

	var tmpDeliveries []spirit.Delivery
	for api, apiData := range apiDatas {

		payload := NewHttpJsonApiPayload()

		if isForwarded {
			if jsonPayload, ok := apiData.(JsonPayload); ok {
				payload.id = jsonPayload.Id
				payload.data = jsonPayload.Data
				payload.metadata = jsonPayload.Metadata
				payload.contexts = jsonPayload.Context

				if jsonPayload.Error != nil {
					errCode := errors.NewErrorCode(
						jsonPayload.Error.Id,
						jsonPayload.Error.Code,
						jsonPayload.Error.Namespace,
						jsonPayload.Error.Message,
						jsonPayload.Error.StackTrace,
						jsonPayload.Error.Context)

					payload.SetError(errCode)
				}
			}
		} else {
			payload.SetData(apiData)
		}

		deliveryURN := ""
		if urn, exist := p.conf.ApiURN[api]; exist {
			deliveryURN = urn
		} else {
			deliveryURN = p.conf.BindURN
		}

		deliveryLabels := spirit.Labels{}
		if labels, exist := p.conf.ApiLabels[api]; exist {
			deliveryLabels = labels
		}

		de := &HttpJsonApiDelivery{
			id:        xid.New().String(),
			payload:   payload,
			urn:       deliveryURN,
			labels:    deliveryLabels,
			timestamp: time.Now(),
		}

		tmpDeliveries = append(tmpDeliveries, de)

		idMapping[de.id] = api
	}

	deliveries = tmpDeliveries
	apiIds = idMapping

	return
}

func (p *JsonApiReceiver) deliveryToApiResponse(delivery spirit.Delivery) (resp APIResponse) {

	var apiResp APIResponse

	toErrResponseFunc := func(err error) APIResponse {
		if errCode, ok := delivery.Payload().GetError().(errors.ErrCode); ok {
			return APIResponse{
				Code:           errCode.Code(),
				ErrorId:        errCode.Id(),
				ErrorNamespace: errCode.Namespace(),
				Message:        errCode.Error(),
				Result:         nil,
			}
		} else {
			errCode := ErrApiGenericError.New().Append(err)
			return APIResponse{
				Code:           errCode.Code(),
				ErrorId:        errCode.Id(),
				ErrorNamespace: errCode.Namespace(),
				Message:        errCode.Error(),
				Result:         nil,
			}
		}
	}

	if e := delivery.Payload().GetError(); e != nil {
		apiResp = toErrResponseFunc(e)
	} else {
		if data, e := delivery.Payload().GetData(); e != nil {
			apiResp = toErrResponseFunc(e)
		} else {
			apiResp = APIResponse{
				Code:   0,
				Result: data,
			}
		}
	}

	resp = apiResp

	return
}

func (p *JsonApiReceiver) writeResponse(data []byte, w gohttp.ResponseWriter, r *gohttp.Request) {
	p.writeResponseWithStatusCode(data, w, r, gohttp.StatusOK)
}

func (p *JsonApiReceiver) writeResponseWithStatusCode(data []byte, w gohttp.ResponseWriter, r *gohttp.Request, code int) {
	p.writeAccessHeaders(w, r)
	p.writeBasicHeaders(w, r)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(data)
}

func (p *JsonApiReceiver) writeAccessHeaders(w gohttp.ResponseWriter, r *gohttp.Request) {
	refer := r.Referer()
	if refer == "" {
		refer = r.Header.Get("Origin")
	}

	if origin, isAllowed := p.conf.AccessControl.ParseOrigin(refer); isAllowed {
		w.Header().Set("Access-Control-Allow-Origin", origin)
	}

	if p.conf.AccessControl.Credentials {
		w.Header().Set("Access-Control-Allow-Credentials", "true")
	}

	if len(p.conf.AccessControl.methods) > 0 {
		w.Header().Set("Access-Control-Allow-Methods", p.conf.AccessControl.methods)
	}

	if len(p.conf.AccessControl.headers) > 0 {
		w.Header().Set("Access-Control-Allow-Headers", p.conf.AccessControl.headers)
	}
}

func (p *JsonApiReceiver) writeBasicHeaders(w gohttp.ResponseWriter, r *gohttp.Request) {
	if p.conf.ResponseHeaders != nil {
		for key, value := range p.conf.ResponseHeaders {
			w.Header().Set(key, value)
		}
	}
}
