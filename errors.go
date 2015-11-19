package http_json_api

import (
	"github.com/gogap/errors"
)

const (
	HttpJsonApiErrNamespace = "JSON_API"
)

var (
	ErrApiNameIsEmpty = errors.New("api name is empty")
)

var (
	ErrTmplVarAlreadyExist   = errors.TN(HttpJsonApiErrNamespace, 400, "template var already exist, key: {{.key}}, value: {{.value}}, original value: {{.originalValue}}")
	ErrApiAlreadyRelatedTmpl = errors.TN(HttpJsonApiErrNamespace, 401, "api already related template, api: {{.apiName}}, template: {{.tmplName}}")
	ErrTmplNotExit           = errors.TN(HttpJsonApiErrNamespace, 402, "template of {{.tmplName}} not exist")
	ErrRequestTimeout        = errors.TN(HttpJsonApiErrNamespace, 408, "request timeout")

	ErrApiGenericError            = errors.TN(HttpJsonApiErrNamespace, 500, "")
	ErrNotSupportMultiCallForward = errors.TN(HttpJsonApiErrNamespace, 501, "not support multi call forward")
	ErrRenderApiDataFailed        = errors.TN(HttpJsonApiErrNamespace, 502, "render api data failed")
)
