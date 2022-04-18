package {{.Package}}

import (
	"fmt"
	"html/template"

	"github.com/Masterminds/sprig"
	"github.com/direktiv/apps/go/pkg/apps"
)

const (
	successKey = "success"
	resultKey = "result"

	// http related
	statusKey = "status"
	codeKey = "code"
	headersKey = "headers"
)

{{ $direktiv := index .Extensions "x-direktiv" }}
{{ $commands := (index $direktiv "cmds") }}

func DirektivHandle(params PostParams) *PostOK {
	fmt.Printf("LOADRsssAMS %v\n", *params.Limit)

	ri := &apps.RequestInfo{}

	{{- $length := len $commands }}
	responses := make(map[string]map[string]interface{}, {{ $length }})
	fmt.Printf("CMDS %v", responses)

	{{- range $i,$e := $commands }}

	{{- end }}

	return NewPostOK()
}

// start commands
{{- range $i,$e := $commands }}
// {{ $i }} {{ $e }}

{{ $action := index $e "action" }}

// http
{{ $url := index $e "url" }}


{{- if eq $action "cmd" }}
// ACTION
{{- else }}

// http request
func runCommand{{ $i }}(params *PostParams, ri *reusable.RequestInfo) (map[string]interface{}, error) {

	ir := make(map[string]interface{})
	ir[successKey] = false

	u, err := templateString("{{ .url }}", params)
	if err != nil {
		return ir, err
	}

	fmt.Printf("U %v", u)

	return nil, nil

}

{{- end }}

// end commands
{{- end }}

func templateString(tmplIn string, data interface{}) (string, error) {

	tmpl, err :=template.New("base").Funcs(sprig.FuncMap()).Parse(tmplIn)
	if err != nil {
		return "", err
	}

	var b bytes.Buffer
	err = tmpl.Execute(&b, data)
	if err != nil {
		return "", err
	}

	return html.UnescapeString(b.String()), nil
	
}

func doHttpRequest(method, u, user, pwd string, 
	headers map[string]string, 
	insecure, errNo200 bool) (map[string]interface{}, error) {

	ir := make(map[string]interface{})
	ir[successKey] = false

	urlParsed, err := url.Parse(u)
	if err != nil {
		return ir, err
	}

	v, err := url.ParseQuery(urlParsed.RawQuery)
	if err != nil {
		return ir, err
	}

	urlParsed.RawQuery = v.Encode()

	req, err := http.NewRequest(method, urlParsed.String(), nil)
	if err != nil {
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
		return ir, err
	}
	defer resp.Body.Close()
	
	b, err := io.ReadAll(resp.Body)
	if err != nil {
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

	if err != nil {
		ir[resultKey] = base64.StdEncoding.EncodeToString(b)
	} 

	return ir, nil
	
}
