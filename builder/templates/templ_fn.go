package {{.Package}}

import (
	"fmt"
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

{{- $direktiv := index .Extensions "x-direktiv" }}
{{- $commands := (index $direktiv "cmds") }}
{{- $debug := (index $direktiv "debug") }}

{{- $printDebug := false }}
{{- if ne $debug nil }}
	{{- $printDebug = $debug }}
{{- end }}

const (
	cmdErr = "io.direktiv.command.error"
	outErr = "io.direktiv.output.error"
	riErr = "io.direktiv.ri.error"
)

func DirektivHandle(params PostParams) middleware.Responder {

	{{- if and .SuccessResponse.Schema }}
	var resp {{ .SuccessResponse.Schema.GoType }}
	{{- end }}

	var (
		err error
		ret map[string]interface{}
	)

	ri, err := apps.RequestinfoFromRequest(params.HTTPRequest)
	if err != nil {
		return generateError(riErr, err)
	}
	{{/* command section */}}
	{{- $length := len $commands }}
	responses := make(map[string]map[string]interface{}, {{ $length }})

	{{- range $i,$e := $commands }}

	ret, err = runCommand{{ $i }}(params, ri)
	responses["cmd{{ $i }}"] = ret

	if err != nil && {{ if ne .continue nil }}!{{ .continue }}{{ else }}true{{ end }} {
		errName := {{- if ne .error nil }}"{{ .error }}"{{- else }}cmdErr{{- end }}
		return generateError(errName, err)
	}

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

	err = json.Unmarshal(responseBytes, &resp)
	if err != nil {
		return generateError(outErr, err)
	}

	return NewPostOK().WithPayload(&resp)
	{{- else }}
	return NewPostOK()
	{{- end  }}
}

// start commands
{{- range $i,$e := $commands }}

{{ $action := index $e "action" }}

{{- if eq $action "exec" }}

// exec
func runCommand{{ $i }}(params PostParams, ri *apps.RequestInfo) (map[string]interface{}, error) {

	ir := make(map[string]interface{})
	ir[successKey] = false

	ri.Logger().Infof("executing command")

	{{- if $printDebug }}
	fmt.Printf("object going in command template: %+v\n", params.Body)
	{{- end}}

	cmd, err := templateString("{{ .exec }}", params.Body)
	if err != nil {
		ir[resultKey] = err.Error()
		return ir, err
	}

	silent := {{ if ne .silent nil }}{{.silent}}{{ else }}false{{ end }}
	print := {{ if ne .print nil }}{{.print}}{{ else }}true{{ end }}
	output := "{{ if ne .output nil }}{{.output}}{{ end }}"

	envs := []string{}
	{{- range $i,$e := .env }}
		env{{ $i }}, _ := templateString("{{ $e }}", ls) 
		envs = append(envs, env{{ $i }})
	{{- end }} 

	return runCmd(cmd, envs, output, silent, print, ri)

}

{{- else if eq $action "foreach" }}

// foreach command
type LoopStruct{{ $i }} struct {
	PostParams 
	Item interface{}
}

func runCommand{{ $i }}(params PostParams, ri *apps.RequestInfo) (map[string]interface{}, error) {

	ri.Logger().Infof("foreach command over {{ .loop }}")

	cmds := make(map[string]interface{}, len(params.Body{{ .loop }}))

	for a := range params.Body{{ .loop }} {

		ls := &LoopStruct{{ $i }}{
			params,
			params.Body{{ .loop }}[a],
		}

		{{- if $printDebug }}
		fmt.Printf("object going in command template: %+v\n", ls)
		{{- end}}

		cmd, err := templateString("{{ .exec }}", ls)
		if err != nil {
			ir := make(map[string]interface{})
			ir[successKey] = false
			ir[resultKey] = err.Error()
			cmds[fmt.Sprintf("Foreach%d", a)] = ir
			continue
		}

		silent := {{ if ne .silent nil }}templateString("{{ .silent }}", ls){{ else }}false{{ end }}
		print := {{ if ne .print nil }}templateString("{{ .print }}", ls){{ else }}true{{ end }}
		output := "{{ if ne .output nil }}{{.output}}{{ end }}"

		envs := []string{}
		{{- range $i,$e := .env }}
			env{{ $i }}, _ := templateString("{{ $e }}", ls) 
			envs = append(envs, env{{ $i }})
		{{- end }} 

		r, _ := runCmd(cmd, envs, output, silent, print, ri)
		cmds[fmt.Sprintf("foreach%d", a)] = r

	}

	return cmds, nil

}


{{- else if eq $action "foreachHttp"}}

// foreach http request
type LoopStruct{{ $i }} struct {
	PostParams 
	Item interface{}
}

func runCommand{{ $i }}(params PostParams, ri *apps.RequestInfo) (map[string]interface{}, error) {

	ri.Logger().Infof("foreach http request over {{ .loop }}")

	cmds := make(map[string]interface{}, len(params.Body{{ .loop }}))

	for a := range params.Body{{ .loop }} {

		ls := &LoopStruct{{ $i }}{
			params,
			params.Body{{ .loop }}[a],
		}

		u, err := templateString("{{ .url }}", ls)
		if err != nil {
			return cmds, err
		}

		method, err := templateString("{{ .method }}", ls)
		if err != nil {
			return cmds, err
		}

		user, err := templateString("{{ .username }}", ls)
		if err != nil {
			return cmds, err
		}

		password, err := templateString("{{ .password }}", ls)
		if err != nil {
			return cmds, err
		}

		insecure, err := templateString("{{ .insecure }}", ls)
		if err != nil {
			return cmds, err
		}

		ins, err := strconv.ParseBool(insecure)
		if err != nil {
			return cmds, err
		}

		headers := make(map[string]string)
		{{- range $i,$h := .headers }}
		{{- range $k,$v := $h }}
		{{ $k }}Header, err := templateString("{{ $v }}", params)
		headers["{{ $k }}"] = {{ $k }}Header
		{{- end }}
		{{- end }}
	
		err200 := {{ if ne .errorNo200 nil }}{{.errorNo200}}{{ else }}true{{ end }}
		ri.Logger().Infof("%v request to %v", method, u)

		r, _ := doHttpRequest(method, u, user, password, 
			headers, ins, err200)

		ri.Logger().Infof("request result code %v", r["code"])
		
		cmds[fmt.Sprintf("foreachHttp%d", a)] = r

	}

	return nil, nil

}


{{- else }}

// http request
func runCommand{{ $i }}(params PostParams, ri *apps.RequestInfo) (map[string]interface{}, error) {

	ri.Logger().Infof("running http request")

	ir := make(map[string]interface{})
	ir[successKey] = false

	u, err := templateString("{{ .url }}", params)
	if err != nil {
		ir[resultKey] = err.Error()
		return ir, err
	}

	method, err := templateString("{{ .method }}", params)
	if err != nil {
		ir[resultKey] = err.Error()
		return ir, err
	}

	user, err := templateString("{{ .username }}", params)
	if err != nil {
		ir[resultKey] = err.Error()
		return ir, err
	}

	password, err := templateString("{{ .password }}", params)
	if err != nil {
		ir[resultKey] = err.Error()
		return ir, err
	}

	insecure, err := templateString("{{ .insecure }}", params)
	if err != nil {
		ir[resultKey] = err.Error()
		return ir, err
	}

	ins, err := strconv.ParseBool(insecure)
	if err != nil {
		ins = false
	}

	headers := make(map[string]string)
	{{- range $i,$h := .headers }}
	{{- range $k,$v := $h }}
	{{ $k }}Header, err := templateString("{{ $v }}", params)
	headers["{{ $k }}"] = {{ $k }}Header
	{{- end }}
	{{- end }}

	err200 := {{ if ne .errorNo200 nil }}{{.errorNo200}}{{ else }}true{{ end }}
	ri.Logger().Infof("%v request to %v", method, u)

	return doHttpRequest(method, u, user, password, 
		headers, ins, err200)

}

{{- end }}

// end commands
{{- end }}

func templateString(tmplIn string, data interface{}) (string, error) {


	{{- if $printDebug }}
	fmt.Printf("template to use: %+v\n", tmplIn)
	{{- end}}

	tmpl, err :=template.New("base").Funcs(sprig.FuncMap()).Parse(tmplIn)
	if err != nil {
		{{- if $printDebug }}
		fmt.Printf("template failed: %+v\n", err)
		{{- end}}
		return "", err
	}

	var b bytes.Buffer
	err = tmpl.Execute(&b, data)
	if err != nil {
		{{- if $printDebug }}
		fmt.Printf("template failed: %+v\n", err)
		{{- end}}
		return "", err
	}

	{{- if $printDebug }}
	fmt.Printf("template output: %+v\n", html.UnescapeString(b.String()))
	{{- end}}

	v := b.String()
	if (v == "<no value>") {
		v = ""
	}

	return html.UnescapeString(v), nil
	
}

func doHttpRequest(method, u, user, pwd string, 
	headers map[string]string, 
	insecure, errNo200 bool) (map[string]interface{}, error) {

	ir := make(map[string]interface{})
	ir[successKey] = false

	urlParsed, err := url.Parse(u)
	if err != nil {
		ir[resultKey] = err.Error()
		return ir, err
	}

	v, err := url.ParseQuery(urlParsed.RawQuery)
	if err != nil {
		ir[resultKey] = err.Error()
		return ir, err
	}

	urlParsed.RawQuery = v.Encode()

	req, err := http.NewRequest(strings.ToUpper(method), urlParsed.String(), nil)
	if err != nil {
		ir[resultKey] = err.Error()
		return ir, err
	}

	for k, v := range headers {
		req.Header.Add(k, v)
	}

	if user != "" {
		req.SetBasicAuth(user, pwd)
	}

	jar, err := cookiejar.New(&cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	})
	cr := http.DefaultTransport.(*http.Transport).Clone()
	cr.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: insecure,
	}

	client := &http.Client{
		Jar:       jar,
		Transport: cr,
	}

	resp, err := client.Do(req)
	if err != nil {
		ir[resultKey] = err.Error()
		return ir, err
	}
	defer resp.Body.Close()
	
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		ir[resultKey] = err.Error()
		return ir, err
	}

	if errNo200 && (resp.StatusCode < 200 || resp.StatusCode > 299) {
		err = fmt.Errorf("response status is not between 200-299")
		ir[resultKey] = err.Error()
		return ir, err
	}

	// from here on it is successful
	ir[successKey] = true
	ir[statusKey] = resp.Status
	ir[codeKey]= resp.StatusCode
	ir[headersKey] = resp.Header

	var rj interface{}
	err = json.Unmarshal(b, &rj)
	ir[resultKey] = rj

	// if the response is not json, base64 the result
	if err != nil {
		ir[resultKey] = base64.StdEncoding.EncodeToString(b)
	}

	return ir, nil
	
}

func runCmd(cmdString string, envs []string,
	output string, silent, print bool, ri *apps.RequestInfo) (map[string]interface{}, error) {

	ir := make(map[string]interface{})
	ir[successKey] = false

	a, err := shellwords.Parse(cmdString)
	if err != nil {
		ir[resultKey] = err.Error()
		return ir, err
	}

	// get the binary and args
	bin := a[0]
	argsIn := []string{}
	if len(a) > 1 {
		argsIn = a[1:]
	}

	logger := io.Discard
	if !silent {
		logger = ri.LogWriter()
	}

	var o bytes.Buffer
	mw := io.MultiWriter(os.Stdout, &o, logger)
	
	cmd := exec.Command(bin, argsIn...)
	cmd.Stdout = mw
	cmd.Stderr = mw
	cmd.Dir = ri.Dir()
	cmd.Env = append(os.Environ(), envs...)

	if print {
		ri.Logger().Infof("running command %v", cmd)
	}

	err = cmd.Run()
	if err != nil {
		ir[resultKey] = err.Error()
		return ir, err
	}

	// successful here
	ir[successKey] = true

	// output check
	b := o.Bytes()
	if output != "" {
		b, err = os.ReadFile(output)
		if err != nil {
			ir[resultKey] = err.Error()
			return ir, err
		}
	}

	var rj interface{}
	err = json.Unmarshal(b, &rj)
	if err != nil {
		ri.Logger().Errorf("can not use output %s as response: %v", output, err)
		rj = apps.ToJSON(o.String())
	} 
    ir[resultKey] = rj

	return ir, nil

}

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
