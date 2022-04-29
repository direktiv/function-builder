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
	return true
}

func templateString(tmplIn string, data interface{}) (string, error) {

	{{- if $printDebug }}
	fmt.Printf("template to use: %+v\n", tmplIn)
	fmt.Printf("data to use: %+v\n", data)
	{{- end}}

	tmpl, err := template.New("base").Funcs(sprig.FuncMap()).Funcs(template.FuncMap{
		"fileExists": fileExists,
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
	
	cmd := exec.CommandContext(ctx, bin, argsIn...)
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
		rj = apps.ToJSON(o.String())
	} 
    ir[resultKey] = rj

	return ir, nil

}


func doHttpRequest(method, u, user, pwd string, 
	headers map[string]string, 
	insecure, errNo200 bool, data []byte) (map[string]interface{}, error) {

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
