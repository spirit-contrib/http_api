package http_json_api

import (
	"github.com/spirit-contrib/http"
	"net/url"
	"strings"

	"github.com/gogap/spirit"
)

type RendererConfig struct {
	DefaultTemplate string              `json:"default_template"`
	Templates       []string            `json:"templates"`
	Variables       []string            `json:"variables"`
	Relation        map[string][]string `json:"relation"`
}

type AccessControl struct {
	Credentials bool     `json:"credentials"`
	Methods     []string `json:"methods"`
	Headers     []string `json:"headers"`
	Origins     []string `json:"origins"`

	originsMap map[string]bool
	headers    string
	methods    string
}

func (p *AccessControl) initial() {
	p.originsMap = make(map[string]bool)

	if p.Origins != nil {
		for _, origin := range p.Origins {
			origin = strings.TrimSpace(origin)
			origin = strings.TrimRight(origin, "/")
			p.originsMap[origin] = true
		}
	}

	p.headers = strings.Join(p.Headers, ",")
	p.methods = strings.Join(p.Methods, ",")
}

func (p *AccessControl) ParseOrigin(refer string) (origin string, isAllow bool) {
	if _, err := url.Parse(refer); err == nil {
		refProtocol, refDomain := parseRefer(refer)
		origin := refProtocol + "://" + refDomain
		if p.originsMap["*"] ||
			p.originsMap[origin] {
			return origin, true
		}
		return "", false
	}

	return "", false
}

type HeaderDefines struct {
	ApiHeader       string `json:"api"`
	MultiCallHeader string `json:"multi_call"`
	TimeoutHeader   string `json:"timeout"`
}

type XDomainConfig struct {
	HtmlPath string            `json:"html_path"`
	LibPath  string            `json:"lib_path"`
	LibUrl   string            `json:"lib_url"`
	Masters  map[string]string `json:"masters"`
}

type ToContext struct {
	Cookies []string               `json:"cookies"`
	Headers []string               `json:"headers"`
	Customs map[string]interface{} `json:"customs"`
}

type JsonApiReceiverConfig struct {
	Http http.HTTPReceiverConfig `json:"http"`

	HeaderDefines HeaderDefines `json:"header_defines"`

	ApiURN    map[string]string        `json:"api_urn"`
	ApiLabels map[string]spirit.Labels `json:"api_labels"`

	Renderer RendererConfig `json:"renderer"`

	AccessControl AccessControl `json:"access_control"`

	ResponseHeaders map[string]string `json:"response_headers"`

	Path    string `json:"path"`
	Timeout int    `json:"timeout"`
	BindURN string `json:"bind_urn"`

	ToContext ToContext `json:"to_context"`

	XDomain XDomainConfig `json:"xdomain"`
}

func (p *JsonApiReceiverConfig) initial() {
	if p.Path == "" {
		p.Path = "/"
	}

	if p.Timeout == 0 {
		p.Timeout = int(DefaultTimeout)
	}

	if p.HeaderDefines.ApiHeader == "" {
		p.HeaderDefines.ApiHeader = DefaultApiHeader
	}

	if p.HeaderDefines.MultiCallHeader == "" {
		p.HeaderDefines.MultiCallHeader = DefaultApiMultiCallHeader
	}

	if p.HeaderDefines.TimeoutHeader == "" {
		p.HeaderDefines.TimeoutHeader = DefaultApiTimeoutHeader
	}

	distinctCache := map[string]string{}

	for _, header := range internalAllowHeaders {
		distinctCache[strings.ToLower(header)] = header
	}

	for _, header := range p.AccessControl.Headers {
		if _, exist := distinctCache[strings.ToLower(header)]; !exist {
			distinctCache[strings.ToLower(header)] = header
		}
	}

	distinctCache[strings.ToLower(p.HeaderDefines.ApiHeader)] = p.HeaderDefines.ApiHeader
	distinctCache[strings.ToLower(p.HeaderDefines.MultiCallHeader)] = p.HeaderDefines.MultiCallHeader
	distinctCache[strings.ToLower(p.HeaderDefines.TimeoutHeader)] = p.HeaderDefines.TimeoutHeader

	allowHeaders := []string{}

	for _, header := range distinctCache {
		allowHeaders = append(allowHeaders, header)
	}

	p.AccessControl.Headers = allowHeaders

	p.AccessControl.initial()
}
