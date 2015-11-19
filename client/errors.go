package api_client

import (
	"github.com/gogap/errors"
)

const (
	HttpJsonApiErrNamespace = "JSON_APIClient"
)

var JsonApiClientErrorNamespace = "APIClient"

var (
	ErrAPINameIsEmpty                   = errors.TN(JsonApiClientErrorNamespace, 1, "api name is empty")
	ErrAPIClientSendFailed              = errors.TN(JsonApiClientErrorNamespace, 2, "api client send failed, api: {{.api}}, url: {{.url}}")
	ErrAPIClientResponseUnmarshalFailed = errors.TN(JsonApiClientErrorNamespace, 3, "api response unmarshal failed, api: {{.api}}, url: {{.url}}")

	ErrAPIClientReadResponseBodyFailed = errors.TN(JsonApiClientErrorNamespace, 4, "read api response body failed, api is: {{.api}}")
	ErrAPIClientBadStatusCode          = errors.TN(JsonApiClientErrorNamespace, 5, "bad response status code, api is: {{.api}}, code is: {{.code}}")
	ErrAPIClientCreateNewRequestFailed = errors.TN(JsonApiClientErrorNamespace, 6, "create new request failed")

	ErrUnknownPayloadError = errors.TN(JsonApiClientErrorNamespace, 7, "")
)
