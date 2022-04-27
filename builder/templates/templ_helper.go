package {{.Package}}

import (
	"html/template"
)

{{- $printDebug := false }}
{{- $direktiv := index .Extensions "x-direktiv" }}
{{- $debug := (index $direktiv "debug") }}

{{- if ne $debug nil }}
	{{- $printDebug = $debug }}
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

func runCmd(ctx context.Context, cmdString string, envs []string,
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