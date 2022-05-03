package {{.Package}}

{{- $printDebug := false }}

{{- if eq .Name "Post" }}

{{- $direktiv := index .Extensions "x-direktiv" }}
{{- $commands := (index $direktiv "cmds") }}
{{- $debug := (index $direktiv "debug") }}

{{- if ne $debug nil }}
	{{- $printDebug = $debug }}
{{- end }}


import (
	{{- if $printDebug }}
	"fmt"
	{{- end }}
	"html/template"

	"github.com/Masterminds/sprig"
	"github.com/direktiv/apps/go/pkg/apps"

{{ imports .DefaultImports }}
)

const (
	successKey = "success"
	resultKey = "result"

	// http related
	statusKey = "status"
	codeKey = "code"
	headersKey = "headers"
)

var sm sync.Map

const (
	cmdErr = "io.direktiv.command.error"
	outErr = "io.direktiv.output.error"
	riErr = "io.direktiv.ri.error"
)

type accParams struct {
	PostParams
	Commands []interface{}
}

type accParamsTemplate struct {
	{{- range $i,$e := .Params }}
	{{- if eq $e.Name "body" }}
	{{- if eq $e.Schema.GoType "interface{}"}}
	In {{ $e.Schema.GoType }}
	{{- else }}
	{{ $e.Schema.GoType }}
	{{- end }}
	{{- end }}
	{{- end }}
	Commands []interface{}
}

func PostDirektivHandle(params PostParams) middleware.Responder {

	{{- if $printDebug }}
	fmt.Printf("params in: %+v", params)
	{{- end }}

	{{- if .SuccessResponse.Schema }}
	{{- if eq .SuccessResponse.Schema.GoType "interface{}"}}
	var resp {{ .SuccessResponse.Schema.GoType }}
	{{- else }}
	resp := &{{ .SuccessResponse.Schema.GoType }}{}
	{{- end}}
	{{- end }}

	var (
		err error
		ret interface{}
	)

	ri, err := apps.RequestinfoFromRequest(params.HTTPRequest)
	if err != nil {
		return generateError(riErr, err)
	}

	ctx, cancel := context.WithCancel(params.HTTPRequest.Context())
	sm.Store(*params.DirektivActionID, cancel)
	defer sm.Delete(params.DirektivActionID)

	{{/* command section */}}
	{{- $length := len $commands }}
	var responses []interface{}

	var paramsCollector []interface{}
	accParams := accParams{
		params,
		nil,
	}

	{{- range $i,$e := $commands }}

	ret, err = runCommand{{ $i }}(ctx, accParams, ri)
	responses = append(responses, ret)

	if err != nil && {{ if ne .continue nil }}!{{ .continue }}{{ else }}true{{ end }} {
		errName := {{- if ne .error nil }}"{{ .error }}"{{- else }}cmdErr{{- end }}
		return generateError(errName, err)
	}

	paramsCollector = append(paramsCollector, ret)
	accParams.Commands = paramsCollector

	{{- end }}

	{{/* output section */}}
	{{/* if there is no response schema defined, we can skip this part */}}
	{{- if and .SuccessResponse.Schema }}

	{{/* output defined */}}
	{{- if $direktiv.output }}

	{{- if $printDebug }}
	fmt.Printf("object going in output template: %+v\n", responses)
	{{- end}}

	s, err := templateString(`{{ $direktiv.output }}`, responses)
	if err != nil {
		return generateError(outErr, err)
	}
	responseBytes := []byte(s)
	{{/* default answer */}}
	{{- else }}
	responseBytes, err := json.Marshal(responses)
	{{- end  }}

	{{- if .SuccessResponse.Schema }}
	{{- if eq .SuccessResponse.Schema.GoType "interface{}"}}
	err = json.Unmarshal(responseBytes, &resp)
	if err != nil {
		return generateError(outErr, err)
	}
	{{- else }}
	// validate
	resp.UnmarshalBinary(responseBytes)
	err = resp.Validate(strfmt.Default)
	if err != nil {
		return generateError(outErr, err)
	}
	{{- end}}
	{{- end }}

	return NewPostOK().WithPayload(resp)
	{{- else }}
	return NewPostOK()
	{{- end  }}
}

{{/* start commands */}}
{{- range $i,$e := $commands }}

{{ $action := index $e "action" }}

{{- if eq $action "exec" }}

// exec
func runCommand{{ $i }}(ctx context.Context, 
		params accParams, ri *apps.RequestInfo) (map[string]interface{}, error) {

	ir := make(map[string]interface{})
	ir[successKey] = false

	ri.Logger().Infof("executing command")

	at := accParamsTemplate{
		params.Body,
		params.Commands,
	}

	{{- if $printDebug }}
	fmt.Printf("object going in command template: %+v\n", at)
	{{- end}}

	cmd, err := templateString(`{{ .exec }}`, at)
	if err != nil {
		ir[resultKey] = err.Error()
		return ir, err
	}
	cmd = strings.Replace(cmd, "\n", "", -1)
	
	silent := convertTemplateToBool("{{ .silent }}", at, false)
	print := convertTemplateToBool("{{ .print }}", at, true)
	output := "{{ if ne .output nil }}{{.output}}{{ end }}"

	envs := []string{}
	{{- range $i,$e := .env }}
		env{{ $i }}, _ := templateString(`{{ $e }}`, at) 
		envs = append(envs, env{{ $i }})
	{{- end }} 

	return runCmd(ctx, cmd, envs, output, silent, print, ri)

}

{{- else if eq $action "foreach" }}

// foreach command
type LoopStruct{{ $i }} struct {
	accParams 
	Item interface{}
}

func runCommand{{ $i }}(ctx context.Context, 
		params accParams, ri *apps.RequestInfo) ([]map[string]interface{}, error) {

	ri.Logger().Infof("foreach command over {{ .loop }}")

	var cmds []map[string]interface{}

	for a := range params.Body{{ .loop }} {

		ls := &LoopStruct{{ $i }}{
			params,
			params.Body{{ .loop }}[a],
		}

		{{- if $printDebug }}
		fmt.Printf("object going in command template: %+v\n", ls)
		{{- end}}

		cmd, err := templateString(`{{ .exec }}`, ls)
		if err != nil {
			ir := make(map[string]interface{})
			ir[successKey] = false
			ir[resultKey] = err.Error()
			cmds = append(cmds, ir)
			continue
		}

		silent := {{ if ne .silent nil }}templateString("{{ .silent }}", ls){{ else }}false{{ end }}
		print := {{ if ne .print nil }}templateString("{{ .print }}", ls){{ else }}true{{ end }}
		output := "{{ if ne .output nil }}{{.output}}{{ end }}"

		envs := []string{}
		{{- range $i,$e := .env }}
			env{{ $i }}, _ := templateString(`{{ $e }}`, ls) 
			envs = append(envs, env{{ $i }})
		{{- end }} 

		r, _ := runCmd(ctx, cmd, envs, output, silent, print, ri)
		cmds = append(cmds, r)

	}

	return cmds, nil

}


{{- else if eq $action "foreachHttp"}}

type LoopStruct{{ $i }} struct {
	accParams 
	Item interface{}
}

func runCommand{{ $i }}(ctx context.Context, 
		params accParams, ri *apps.RequestInfo) ([]map[string]interface{}, error) {

	ri.Logger().Infof("foreach http request over {{ .loop }}")

	var cmds []map[string]interface{}

	for a := range params.Body{{ .loop }} {

		ls := &LoopStruct{{ $i }}{
			params,
			params.Body{{ .loop }}[a],
		}

		u, err := templateString(`{{ .url }}`, ls)
		if err != nil {
			return cmds, err
		}

		method, err := templateString(`{{ .method }}`, ls)
		if err != nil {
			return cmds, err
		}

		user, err := templateString(`{{ .username }}`, ls)
		if err != nil {
			return cmds, err
		}

		password, err := templateString(`{{ .password }}`, ls)
		if err != nil {
			return cmds, err
		}

		ins := convertTemplateToBool(`{{ .insecure }}`, ls, false) 

		headers := make(map[string]string)
		{{- range $i,$h := .headers }}
		{{- range $k,$v := $h }}
		Header{{ $i }}, err := templateString(`{{ $v }}`, params)
		headers["{{ $k }}"] = Header{{ $i }}
		{{- end }}
		{{- end }}
	
		err200 := {{ if ne .errorNo200 nil }}{{.errorNo200}}{{ else }}true{{ end }}
		ri.Logger().Infof("%v request to %v", method, u)

		var data []byte
		{{- if .data }}

		value, err := templateString(`{{ .data }}`, ls)
		if err != nil {
			return cmds, err
		}

		{{- if $printDebug }}
		fmt.Printf("sending data: %v\n", value)
		{{- end }}

		data = []byte(value)
		{{- else if .data64 }}

		value, err := templateString(`{{ .data64 }}`, ls)
		if err != nil {
			return cmds, err
		}
		data, err = base64.StdEncoding.DecodeString(value)
		if err != nil {
			return cmds, err
		}
		{{- end }}

		r, _ := doHttpRequest(method, u, user, password, 
			headers, ins, err200, data)

		ri.Logger().Infof("request result code %v", r["code"])
		cmds = append(cmds, r)

	}

	return cmds, nil

}


{{- else }}

// http request
func runCommand{{ $i }}(ctx context.Context,
	params accParams, ri *apps.RequestInfo) (map[string]interface{}, error) {

	ri.Logger().Infof("running http request")

	at := accParamsTemplate{
		params.Body,
		params.Commands,
	}


	ir := make(map[string]interface{})
	ir[successKey] = false

	u, err := templateString(`{{ .url }}`, at)
	if err != nil {
		ir[resultKey] = err.Error()
		return ir, err
	}

	method, err := templateString(`{{ .method }}`, at)
	if err != nil {
		ir[resultKey] = err.Error()
		return ir, err
	}

	user, err := templateString(`{{ .username }}`, at)
	if err != nil {
		ir[resultKey] = err.Error()
		return ir, err
	}

	password, err := templateString(`{{ .password }}`, at)
	if err != nil {
		ir[resultKey] = err.Error()
		return ir, err
	}

	ins := convertTemplateToBool(`{{ .insecure }}`, at, false) 

	headers := make(map[string]string)
	{{- range $i,$h := .headers }}
	{{- range $k,$v := $h }}
	Header{{ $i }}, err := templateString(`{{ $v }}`, params)
	headers["{{ $k }}"] = Header{{ $i }}
	{{- end }}
	{{- end }}

	err200 := {{ if ne .errorNo200 nil }}{{.errorNo200}}{{ else }}true{{ end }}
	ri.Logger().Infof("%v request to %v", method, u)

	var data []byte
	{{- if .data }}
	value, err := templateString(`{{ .data }}`, params)
	if err != nil {
		ir[resultKey] = err.Error()
		return ir, err
	}
	data = []byte(value)
	{{- else if .data64 }}
	value, err := templateString(`{{ .data64 }}`, params)
	if err != nil {
		ir[resultKey] = err.Error()
		return ir, err
	}
	data, err = base64.StdEncoding.DecodeString(value)
	if err != nil {
		ir[resultKey] = err.Error()
		return ir, err
	}
	{{- end }}

	return doHttpRequest(method, u, user, password, 
		headers, ins, err200, data)

}

{{- end }}

// end commands
{{- end }}


func generateError(code string, err error) *PostDefault {

	d := NewPostDefault(0).WithDirektivErrorCode(code).
		WithDirektivErrorMessage(err.Error())

	errString := err.Error()

	errResp := {{ .DefaultResponse.Schema.GoType }}{
		ErrorCode: &code,
		ErrorMessage: &errString,
	}

	d.SetPayload(&errResp)

	return d
}


func HandleShutdown() {
	// nothing for generated functions
}

{{- else }}

import (
	"fmt"
	"html/template"

	"github.com/go-openapi/runtime/middleware"
	"github.com/direktiv/apps/go/pkg/apps"


{{ imports .DefaultImports }}
)


{{- $direktiv := index .Extensions "x-direktiv" }}

func DeleteDirektivHandle(params DeleteParams) middleware.Responder {

	actionId := *params.DirektivActionID
	defer sm.Delete(actionId)

	if actionId == "" {
		return NewDeleteOK()
	}

	ri, err := apps.RequestinfoFromRequest(params.HTTPRequest)
	if err != nil {
		fmt.Println("can not create ri from request")
		return NewDeleteOK()
	}	

	cancel, ok := sm.Load(actionId)
	if !ok {
		ri.Logger().Infof("can not load context for action id: %v", err)
		return NewDeleteOK()
	}

	ri.Logger().Infof("cancelling action id %v", actionId)

	cf, ok := cancel.(context.CancelFunc)
	if !ok {
		ri.Logger().Infof("can not get cancel function for action id: %v", err)
		return NewDeleteOK()
	}

	cf()

	cmd, err := templateString("{{ $direktiv.cancel }}", params)
	if err != nil {
		ri.Logger().Infof("can not template cancel command: %v", err)
		return NewDeleteOK()
	}

	_, err = runCmd(context.Background(), cmd, []string{}, "", false, true, ri)
	if err != nil {
		ri.Logger().Infof("error running cancel function: %v", err)
	}

	return NewDeleteOK()

}

{{- end }}
