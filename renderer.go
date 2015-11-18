package http_json_api

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/gogap/errors"
)

type APIResponse struct {
	Code           uint64      `json:"code"`
	ErrorId        string      `json:"error_id,omitempty"`
	ErrorNamespace string      `json:"error_namespace,omitempty"`
	Message        string      `json:"message"`
	Result         interface{} `json:"result"`
}

type APIRenderData struct {
	IsMulti  bool
	Name     string
	Response APIResponse
}

type RenderData struct {
	API  APIRenderData
	Vars map[string]interface{}
}

type APIResponseRenderer struct {
	apiTemplate map[string]string
	template.Template
	Variables       map[string]interface{}
	defaultTemplate string
}

func NewAPIResponseRenderer() *APIResponseRenderer {
	renderer := &APIResponseRenderer{
		Template:        *template.New(""),
		apiTemplate:     make(map[string]string),
		defaultTemplate: "_internal/default",
		Variables:       make(map[string]interface{}),
	}

	renderer.Funcs(funcMap)

	if e := renderer.AddInternalTemplate(defaultAPITemplate()); e != nil {
		panic(e)
	}

	return renderer
}

func (p *APIResponseRenderer) LoadTemplates(paths ...string) (err error) {
	if paths == nil {
		return
	}

	addToTemplateFunc := func(base string, file string) (err error) {
		var tmplData []byte
		if base[0] != '.' && base[0] != '~' {
			if tmplData, err = ioutil.ReadFile(file); err != nil {
				return
			}

			if err = p.AddTemplate(base, string(tmplData)); err != nil {
				return
			}
		}
		return
	}

	for _, path := range paths {
		var fi os.FileInfo

		if fi, err = os.Stat(path); err != nil {
			return
		}

		if !fi.IsDir() {
			base := filepath.Base(path)
			if err = addToTemplateFunc(base, path); err != nil {
				return
			}
		} else {

			var matches []string
			if matches, err = filepath.Glob(filepath.Join(path, "*.tmpl")); err != nil {
				return
			}

			for _, file := range matches {
				relPath := ""
				if relPath, err = filepath.Rel(path, file); err != nil {
					return
				}

				if err = addToTemplateFunc(relPath, file); err != nil {
					return
				}
			}
		}
	}

	return
}

func (p *APIResponseRenderer) LoadVariables(paths ...string) (err error) {
	if paths == nil {
		return
	}

	appendVarsFunc := func(base string, file string) (err error) {
		if base[0] != '.' && base[0] != '~' {
			var jsonData []byte
			if jsonData, err = ioutil.ReadFile(file); err != nil {
				return
			} else {
				decodder := json.NewDecoder(bytes.NewReader(jsonData))
				decodder.UseNumber()
				vars := map[string]interface{}{}
				if err = decodder.Decode(&vars); err != nil {
					return
				}
				if err = p.appendVars(vars); err != nil {
					return
				}
			}
		}
		return
	}

	for _, path := range paths {
		var fi os.FileInfo

		if fi, err = os.Stat(path); err != nil {
			return
		}

		if !fi.IsDir() {
			base := filepath.Base(path)
			if err = appendVarsFunc(base, path); err != nil {
				return
			}
		} else {
			var matches []string
			if matches, err = filepath.Glob("*.tmpl"); err != nil {
				return
			}

			for _, file := range matches {
				relPath := ""
				if relPath, err = filepath.Rel(path, file); err != nil {
					return
				}

				if err = appendVarsFunc(relPath, filepath.Join(path, file)); err != nil {
					return
				}
			}
		}
	}

	return
}

func (p *APIResponseRenderer) SetDefaultTemplate(name string) (err error) {
	name = strings.TrimSpace(name)
	if name == "" {
		p.defaultTemplate = "_internal/default"
		return
	}

	if p.Lookup(name) != nil {
		p.defaultTemplate = name
	} else {
		err = ErrTmplNotExit.New(errors.Params{"tmplName": name})
		return
	}
	return
}

func (p *APIResponseRenderer) ResetAPITemplate(apiName string) {
	if _, exist := p.apiTemplate[apiName]; exist {
		delete(p.apiTemplate, apiName)
	}
}

func (p *APIResponseRenderer) SetAPITemplate(apiName, tplName string) (err error) {
	if originalName, exist := p.apiTemplate[apiName]; exist {
		if originalName != tplName {
			err = ErrApiAlreadyRelatedTmpl.New(errors.Params{"apiName": apiName, "tmplName": originalName})
			return
		}
		return
	}

	if p.Lookup(tplName) != nil {
		p.apiTemplate[apiName] = tplName
		return
	} else {
		err = ErrTmplNotExit.New(errors.Params{"tmplName": tplName})
		return
	}

	return
}

func (p *APIResponseRenderer) AddInternalTemplate(name, tpl string) error {
	return p.AddTemplate("_internal/"+name, tpl)
}

func (p *APIResponseRenderer) AddTemplate(name, tpl string) (err error) {
	tpl = strings.Replace(tpl, "\n", "", -1)
	tpl = strings.Replace(tpl, "\t", "", -1)
	_, err = p.New(name).Parse(tpl)
	return
}

func (p *APIResponseRenderer) Render(isMulti bool, response map[string]APIResponse) (renderedData []byte, err error) {
	output := map[string]string{}

	for api, response := range response {

		renderData := RenderData{
			API: APIRenderData{
				false,
				api,
				response,
			},
			Vars: p.Variables,
		}

		tmplName := p.defaultTemplate
		if name, exist := p.apiTemplate[api]; exist {
			tmplName = name
		}

		var buf bytes.Buffer
		if err = p.ExecuteTemplate(&buf, tmplName, renderData); err != nil {
			return
		}

		if !isMulti {
			renderedData = buf.Bytes()
			return
		}

		output[api] = buf.String()
	}

	var buf bytes.Buffer

	multiResponse := APIResponse{
		Code:    0,
		Message: "",
		Result:  output,
	}

	multiRenderData := RenderData{
		API: APIRenderData{
			true,
			"",
			multiResponse,
		},
		Vars: p.Variables,
	}

	if err = p.ExecuteTemplate(&buf, p.defaultTemplate, multiRenderData); err != nil {
		return
	}

	renderedData = buf.Bytes()

	return
}

func (p *APIResponseRenderer) appendVars(vars map[string]interface{}) (err error) {
	for k, v := range vars {
		if original, exist := p.Variables[k]; exist {
			if original != v {
				err = ErrTmplVarAlreadyExist.New(errors.Params{"key": k, "value": v, "originalValue": original})
				return
			}
		} else {
			p.Variables[k] = v
		}
	}
	return
}
