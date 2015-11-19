package api_client

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/gogap/errors"
	"github.com/gogap/spirit"
	"github.com/mreiferson/go-httpclient"
)

var (
	DefaultClientTimeout = time.Second * 5
)

type HTTPAPIClient struct {
	apiHeaderName string
	url           string
	client        *http.Client
}

func NewHTTPAPIClient(url string, apiHeaderName string, timeout time.Duration) APIClient {
	url = strings.TrimSpace(url)
	apiHeaderName = strings.TrimSpace(apiHeaderName)

	if url == "" {
		panic("url could not be nil")
	}

	if apiHeaderName == "" {
		apiHeaderName = "X-Api"
	}

	if timeout <= 0 {
		timeout = DefaultClientTimeout
	}

	transport := &httpclient.Transport{
		ConnectTimeout:        timeout,
		RequestTimeout:        timeout,
		ResponseHeaderTimeout: timeout,
	}

	apiClient := HTTPAPIClient{
		apiHeaderName: apiHeaderName,
		url:           url,
		client:        &http.Client{Transport: transport},
	}
	return &apiClient
}

func (p *HTTPAPIClient) Call(apiName string, payload spirit.Payload, v interface{}) (err error) {
	apiName = strings.TrimSpace(apiName)

	if apiName == "" {
		err = ErrAPINameIsEmpty.New()
		return
	}

	var payloadData interface{}

	if payloadData, err = payload.GetData(); err != nil {
		return
	}

	jsonPayload := JsonPayload{
		Id:       payload.Id(),
		Data:     payloadData,
		Metadata: payload.Metadata(),
		Context:  payload.Contexts(),
	}

	if e := payload.GetError(); e != nil {
		if errCode, ok := e.(errors.ErrCode); ok {
			jsonPayload.Error = &ErrorCode{
				Namespace:  errCode.Namespace(),
				Code:       errCode.Code(),
				Id:         errCode.Id(),
				Message:    errCode.Error(),
				StackTrace: errCode.StackTrace(),
				Context:    errCode.Context(),
			}
		} else {
			errCode := ErrUnknownPayloadError.New().Append(e)
			jsonPayload.Error = &ErrorCode{
				Namespace: errCode.Namespace(),
				Code:      errCode.Code(),
				Id:        errCode.Id(),
				Message:   errCode.Error(),
			}
		}
	}

	var data []byte
	if data, err = json.Marshal(jsonPayload); err != nil {
		return
	}

	postBodyReader := bytes.NewReader(data)

	var req *http.Request
	if req, err = http.NewRequest("POST", p.url, postBodyReader); err != nil {
		err = ErrAPIClientCreateNewRequestFailed.New().Append(err)
		return
	}

	req.Header.Add(p.apiHeaderName, apiName)

	var resp *http.Response
	if resp, err = p.client.Do(req); err != nil {
		err = ErrAPIClientSendFailed.New(errors.Params{"api": apiName, "url": p.url})
		return
	}

	var body []byte

	if resp != nil {
		defer resp.Body.Close()

		if bBody, e := ioutil.ReadAll(resp.Body); e != nil {
			err = ErrAPIClientReadResponseBodyFailed.New(errors.Params{"api": apiName}).Append(e)
			return
		} else if resp.StatusCode != http.StatusOK {
			err = ErrAPIClientBadStatusCode.New(errors.Params{"api": apiName, "code": resp.StatusCode})
			return
		} else {
			body = bBody
		}

		if v == nil {
			return
		}
	}

	if v == nil {
		return
	}

	var tmpResp struct {
		Code           uint64      `json:"code"`
		ErrorId        string      `json:"error_id,omitempty"`
		ErrorNamespace string      `json:"error_namespace,omitempty"`
		Message        string      `json:"message"`
		Result         interface{} `json:"result"`
	}

	tmpResp.Result = v

	if e := json.Unmarshal(body, &tmpResp); e != nil {
		err = ErrAPIClientResponseUnmarshalFailed.New(errors.Params{"api": apiName, "url": p.url}).Append(e)
		return
	}

	if tmpResp.Code == 0 {
		return
	} else {
		err = errors.NewErrorCode(tmpResp.ErrorId, tmpResp.Code, tmpResp.ErrorNamespace, tmpResp.Message, "", nil)
		return
	}

	return
}

func (p *HTTPAPIClient) Cast(apiName string, payload spirit.Payload) {
	go p.Call(apiName, payload, nil)
}
