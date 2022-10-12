package {{.Package}}

import (
	"html/template"
	"github.com/mattn/go-shellwords"
	"github.com/direktiv/apps/go/pkg/apps"
	"github.com/Masterminds/sprig"
	"golang.org/x/net/publicsuffix"
)

{{- $printDebug := false }}
{{- $direktiv := index .Extensions "x-direktiv" }}
{{- $debug := (index $direktiv "debug") }}

{{- if ne $debug nil }}
	{{- $printDebug = $debug }}
{{- end }}

func fileExists(file string) bool {
	_, err := os.Open(file)
	if err != nil {
		return false
	}
	return true
}

func file64(path string) string {

	b, err := os.ReadFile(path)
	if err != nil {
		return err.Error()
	}

	return  base64.StdEncoding.EncodeToString(b)

}

func deref(dd interface{}) interface{} {
	switch p := dd.(type) {
	case *string:
		return *p
	case *int:
		return *p
	default:
		return p
	}
}


func templateString(tmplIn string, data interface{}) (string, error) {

	{{- if $printDebug }}
	fmt.Printf("template to use: %+v\n", tmplIn)
	fmt.Printf("data to use: %+v\n", data)
	{{- end}}

	tmpl, err := template.New("base").Funcs(sprig.FuncMap()).Funcs(template.FuncMap{
		"fileExists": fileExists,
		"deref": deref,
		"file64": file64,
	}).Parse(tmplIn)
	
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

func convertTemplateToBool(template string, data interface{}, defaultValue bool) bool {

	out, err := templateString(template, data)
	if err != nil {
		return defaultValue
	}

	ins, err := strconv.ParseBool(out)
	if err != nil {
		return defaultValue
	}

	return ins

}

func runCmd(ctx context.Context, cmdString string, envs []string,
	output string, silent, print bool, ri *apps.RequestInfo) (map[string]interface{}, error) {

	{{- if $printDebug }}
	fmt.Printf("evironment vars: %+v\n", envs)
	{{- end}}
	

	ir := make(map[string]interface{})
	ir[successKey] = false

	a, err := shellwords.Parse(cmdString)
	if err != nil {
		ir[resultKey] = err.Error()
		return ir, err
	}

	if len(a) == 0 {
		return ir, fmt.Errorf("command '%v' parsed to empty array", cmdString)
	}

	// get the binary and args
	bin := a[0]
	argsIn := []string{}
	if len(a) > 1 {
		argsIn = a[1:]
	}

	logger := io.Discard
	stdo := io.Discard
	if !silent {
		logger = ri.LogWriter()
		stdo = os.Stdout
	}

	var o bytes.Buffer
	var oerr bytes.Buffer

	mwStdout := io.MultiWriter(stdo, &o, logger)
	mwStdErr := io.MultiWriter(os.Stdout, &oerr, logger)
	
	cmd := exec.CommandContext(ctx, bin, argsIn...)
	cmd.Stdout = mwStdout
	cmd.Stderr = mwStdErr
	cmd.Dir = ri.Dir()

	// change HOME
	curEnvs := append(os.Environ(), fmt.Sprintf("HOME=%s", ri.Dir()))
	cmd.Env = append(curEnvs, envs...)

	if print {
		ri.Logger().Infof("running command %v", cmd)
	}

	err = cmd.Run()
	if err != nil {
		ir[resultKey] = string(oerr.String())
		if oerr.String() == "" {
			ir[resultKey] = err.Error()
		} else {
			ri.Logger().Errorf(oerr.String())
			err = fmt.Errorf(oerr.String())
		}
		return ir, err
	}

	// successful here
	ir[successKey] = true

	// output check
	b := o.Bytes()
	if output != "" {
		{{- if $printDebug }}
		fmt.Printf("output set to: %s\n", output)
		{{- end }}
		b, err = os.ReadFile(output)
		if err != nil {
			ri.Logger().Infof("output file %s not used", output)
			// ir[resultKey] = err.Error()
			// return ir, err
			b = o.Bytes()
		}
	}

	var rj interface{}
	err = json.Unmarshal(b, &rj)
	if err != nil {
		rj = apps.ToJSON(string(b))
	} 
    ir[resultKey] = rj

	return ir, nil

}

func doHttpRequest(debug bool, method, u, user, pwd string, 
	headers map[string]string, insecure, errNo200 bool, 
	data []byte) (map[string]interface{}, error) {

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

	req, err := http.NewRequest(strings.ToUpper(method), urlParsed.String(), bytes.NewReader(data))
	if err != nil {
		ir[resultKey] = err.Error()
		return ir, err
	}
	req.Close = true

	if user != "" {
		req.SetBasicAuth(user, pwd)
	}

	for k, v := range headers {

		// if it it is Authorization, we do a set
		if k == "Authorization" && v != "" {
			req.Header.Set(k, v)
		} else {
			req.Header.Add(k, v)
		}

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

	if debug {
		fmt.Printf("method: %s, insecure: %v\n", method, insecure)
		fmt.Printf("url: %s\n", req.URL.String())
		fmt.Println("Headers:")
		for k, v := range req.Header {
			fmt.Printf("%v = %v\n", k, v)
		}
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
		err = fmt.Errorf("response status is not between 200-299: %v (%v)", resp.StatusCode, resp.Status)
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
