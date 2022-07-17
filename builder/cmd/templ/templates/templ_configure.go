// This file is safe to edit. Once it exists it will not be overwritten

{{ if .Copyright -}}// {{ comment .Copyright -}}{{ end }}

package {{ .APIPackage }}

import (
  "context"
  "crypto/tls"
  "io"
  "log"
  "net/http"

  "github.com/go-openapi/errors"
  "github.com/go-openapi/runtime"
  "github.com/go-openapi/runtime/middleware"
  "github.com/go-openapi/runtime/security"

  {{ imports .DefaultImports }}
  {{ imports .Imports }}
)

{{ with .GenOpts }}
//go:generate swagger generate server --target {{ .TargetPath }} --name {{ .Name }} --spec {{ .SpecPath }}
{{- if .APIPackage }}{{ if ne .APIPackage "operations" }} --api-package {{ .APIPackage }}{{ end }}{{ end }}
{{- if .ModelPackage }}{{ if ne .ModelPackage "models" }} --model-package {{ .ModelPackage }}{{ end }}{{ end }}
{{- if .ServerPackage }}{{ if ne .ServerPackage "restapi"}} --server-package {{ .ServerPackage }}{{ end }}{{ end }}
{{- if .ClientPackage }}{{ if ne .ClientPackage "client" }} --client-package {{ .ClientPackage }}{{ end }}{{ end }}
{{- if .TemplateDir }} --template-dir {{ .TemplateDir }}{{ end }}
{{- range .Operations }} --operation {{ . }}{{ end }}
{{- range .Tags }} --tags {{ . }}{{ end }}
{{- if .Principal }} --principal {{ .Principal }}{{ end }}
{{- if .DefaultScheme }}{{ if ne .DefaultScheme "http" }} --default-scheme {{ .DefaultScheme }}{{ end }}{{ end }}
{{- range .Models }} --model {{ . }}{{ end }}
{{- if or (not .IncludeModel) (not .IncludeValidator) }} --skip-models{{ end }}
{{- if or (not .IncludeHandler) (not .IncludeParameters ) (not .IncludeResponses) }} --skip-operations{{ end }}
{{- if not .IncludeSupport }} --skip-support{{ end }}
{{- if not .IncludeMain }} --exclude-main{{ end }}
{{- if .ExcludeSpec }} --exclude-spec{{ end }}
{{- if .DumpData }} --dump-data{{ end }}
{{- if .StrictResponders }} --strict-responders{{ end }}
{{ end }}
func configureFlags(api *{{.APIPackageAlias}}.{{ pascalize .Name }}API) {
  // api.CommandLineOptionsGroups = []swag.CommandLineOptionsGroup{ ... }
}

func errorAsJSON(err errors.Error) []byte {
	b, _ := json.Marshal(struct {
		Code    int32  `json:"errorCode"`
		Message string `json:"errorMessage"`
	}{err.Code(), err.Error()})
	return b
}

func asHTTPCode(input int) int {
	if input >= 600 {
		return errors.DefaultHTTPCode
	}
	return input
}

func flattenComposite(errs *errors.CompositeError) *errors.CompositeError {
	var res []error
	for _, er := range errs.Errors {
		switch e := er.(type) {
		case *errors.CompositeError:
			if len(e.Errors) > 0 {
				flat := flattenComposite(e)
				if len(flat.Errors) > 0 {
					res = append(res, flat.Errors...)
				}
			}
		default:
			if e != nil {
				res = append(res, e)
			}
		}
	}
	return errors.CompositeValidationError(res...)
}

func addErrorHeaders(rw http.ResponseWriter, code, message string) {

  rw.Header().Add("Direktiv-ErrorCode", code)
  rw.Header().Add("Direktiv-ErrorMessage", message)

}

// modified version of the openapi error function
// https://github.com/go-openapi/errors/blob/8b5b7790aa74a2148bf29be0d24e1f455fdbc706/api.go
func serveError(rw http.ResponseWriter, r *http.Request, err error) {

  rw.Header().Set("Content-Type", "application/json")
	switch e := err.(type) {
	case *errors.CompositeError:
		er := flattenComposite(e)
		// strips composite errors to first element only
		if len(er.Errors) > 0 {
			serveError(rw, r, er.Errors[0])
		} else {
			// guard against empty CompositeError (invalid construct)
			serveError(rw, r, nil)
		}
	case *errors.MethodNotAllowedError:
		rw.Header().Add("Allow", strings.Join(err.(*errors.MethodNotAllowedError).Allowed, ","))
		rw.WriteHeader(asHTTPCode(int(e.Code())))
    addErrorHeaders(rw, "io.direktiv.method", err.Error())
		if r == nil || r.Method != http.MethodHead {
			_, _ = rw.Write(errorAsJSON(e))
		}
	case errors.Error:
		value := reflect.ValueOf(e)
		if value.Kind() == reflect.Ptr && value.IsNil() {
			rw.WriteHeader(http.StatusInternalServerError)
			_, _ = rw.Write(errorAsJSON(errors.New(http.StatusInternalServerError, "Unknown error")))
			return
		}
    addErrorHeaders(rw, "io.direktiv.unfit", fmt.Sprintf("%s (%v)", e.Error(), e.Code()))
		rw.WriteHeader(asHTTPCode(int(e.Code())))
		if r == nil || r.Method != http.MethodHead {
			_, _ = rw.Write(errorAsJSON(e))
		}
	case nil:
    addErrorHeaders(rw, "io.direktiv.unknown", "Unknown error")
		rw.WriteHeader(http.StatusInternalServerError)
		_, _ = rw.Write(errorAsJSON(errors.New(http.StatusInternalServerError, "Unknown error")))
	default:
    addErrorHeaders(rw, "io.direktiv.error", err.Error())
		rw.WriteHeader(http.StatusInternalServerError)
		if r == nil || r.Method != http.MethodHead {
			_, _ = rw.Write(errorAsJSON(errors.New(http.StatusInternalServerError, err.Error())))
		}
	}

}

func configureAPI(api *{{.APIPackageAlias}}.{{ pascalize .Name }}API) http.Handler {
  // configure the api here
  api.ServeError = serveError

  // Set your custom logger if needed. Default one is log.Printf
  // Expected interface func(string, ...interface{})
  //
  // Example:
  // api.Logger = log.Printf

  api.UseSwaggerUI()
  // To continue using redoc as your UI, uncomment the following line
  // api.UseRedoc()

  {{ range .Consumes }}
    {{- if .Implementation }}
  api.{{ pascalize .Name }}Consumer = {{ .Implementation }}
    {{- else }}
  api.{{ pascalize .Name }}Consumer = runtime.ConsumerFunc(func(r io.Reader, target interface{}) error {
    return errors.NotImplemented("{{.Name}} consumer has not yet been implemented")
  })
    {{- end }}
  {{- end }}
  {{ range .Produces }}
    {{- if .Implementation }}
  api.{{ pascalize .Name }}Producer = {{ .Implementation }}
    {{- else }}
  api.{{ pascalize .Name }}Producer = runtime.ProducerFunc(func(w io.Writer, data interface{}) error {
    return errors.NotImplemented("{{.Name}} producer has not yet been implemented")
  })
    {{- end }}
  {{- end}}
  {{ range .SecurityDefinitions }}
    {{- if .IsBasicAuth }}
  // Applies when the Authorization header is set with the Basic scheme
  if api.{{ pascalize .ID }}Auth == nil {
  api.{{ pascalize .ID }}Auth = func(user string, pass string) ({{ if .PrincipalIsNullable }}*{{ end }}{{.Principal}}, error) {
      return nil, errors.NotImplemented("basic auth  ({{ .ID }}) has not yet been implemented")
    }
  }
    {{- else if .IsAPIKeyAuth }}
  // Applies when the "{{ .Name }}" {{ .Source }} is set
  if api.{{ pascalize .ID }}Auth == nil {
  api.{{ pascalize .ID }}Auth = func(token string) ({{ if .PrincipalIsNullable }}*{{ end }}{{.Principal}}, error) {
      return nil, errors.NotImplemented("api key auth ({{ .ID }}) {{.Name}} from {{.Source}} param [{{ .Name }}] has not yet been implemented")
    }
  }
    {{- else if .IsOAuth2 }}
    if api.{{ pascalize .ID }}Auth == nil {
    api.{{ pascalize .ID }}Auth = func(token string, scopes []string) ({{ if .PrincipalIsNullable }}*{{ end }}{{.Principal}}, error) {
      return nil, errors.NotImplemented("oauth2 bearer auth ({{ .ID }}) has not yet been implemented")
    }
  }
    {{- end }}
  {{- end }}
  {{- if .SecurityDefinitions }}

  // Set your custom authorizer if needed. Default one is security.Authorized()
  // Expected interface runtime.Authorizer
  //
  // Example:
  // api.APIAuthorizer = security.Authorized()
  {{- end }}
  {{- $package := .Package }}
  {{- $apipackagealias := .APIPackageAlias }}
  {{- range .Operations }}
    {{- if .HasFormParams }}
  // You may change here the memory limit for this multipart form parser. Below is the default (32 MB).
  // {{ if ne .Package $package }}{{ .PackageAlias }}{{ else }}{{ $apipackagealias }}{{ end }}.{{ pascalize .Name }}MaxParseMemory = 32 << 20
    {{- end }}
  {{- end }}
  {{ range .Operations }}
    api.{{ if ne .Package $package }}{{pascalize .Package}}{{ end }}{{ pascalize .Name }}Handler =
    {{- if ne .Package $package }}
      {{- .PackageAlias }}.{{- pascalize .Name }}HandlerFunc(func(params {{ .PackageAlias }}.{{- pascalize .Name }}Params
    {{- else }}
      {{- $apipackagealias }}.{{- pascalize .Name }}HandlerFunc(func(params {{ $apipackagealias }}.{{- pascalize .Name }}Params
    {{- end }}
  {{- if $.GenOpts.StrictResponders }}
    {{- if .Authorized}}, principal {{ if .PrincipalIsNullable }}*{{ end }}{{.Principal}}{{end}}) {{.Package}}.{{ pascalize .Name }}Responder {
     return {{.Package}}.{{ pascalize .Name }}NotImplemented()
  {{ else }}
    {{- if .Authorized}}, principal {{if .PrincipalIsNullable }}*{{ end }}{{.Principal}}{{end}}) middleware.Responder {
      return operations.{{- pascalize .Name }}DirektivHandle(params)
  {{ end -}}
    })
  {{- end }}

  api.PreServerShutdown = func() {  }

  api.ServerShutdown = func() {  }

  return setupGlobalMiddleware(api.Serve(setupMiddlewares))
}

// The TLS configuration before HTTPS server starts.
func configureTLS(tlsConfig *tls.Config) {
  // Make all necessary changes to the TLS configuration here.
}

// As soon as server is initialized but not run yet, this function will be called.
// If you need to modify a config, store server instance to stop it individually later, this is the place.
// This function can be called multiple times, depending on the number of serving schemes.
// scheme value will be set accordingly: "http", "https" or "unix".
func configureServer(s *http.Server, scheme, addr string) {
}

// The middleware configuration is for the handler executors. These do not apply to the swagger.json document.
// The middleware executes after routing but before authentication, binding and validation.
func setupMiddlewares(handler http.Handler) http.Handler {
	return handler
}

// The middleware configuration happens before anything, this middleware also applies to serving the swagger.json document.
// So this is a good place to plug in a panic handling middleware, logging and metrics.
func setupGlobalMiddleware(handler http.Handler) http.Handler {
	return handler
}
