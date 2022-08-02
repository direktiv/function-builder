// This file is safe to edit. Once it exists it will not be overwritten

package restapi

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"

	"app/restapi/operations"
)

//go:generate swagger generate server --target ../../app --name Crypto --spec ../../../swagger.yaml --template-dir /home/jensg/go/src/github.com/direktiv/service-builder/docs/examples/crypto/crypto/build/templates --principal interface{}

func configureFlags(api *operations.CryptoAPI) {
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

func configureAPI(api *operations.CryptoAPI) http.Handler {
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

	api.JSONConsumer = runtime.JSONConsumer()

	api.JSONProducer = runtime.JSONProducer()

	api.DeleteHandler = operations.DeleteHandlerFunc(func(params operations.DeleteParams) middleware.Responder {
		return operations.DeleteDirektivHandle(params)
	})
	api.PostHandler = operations.PostHandlerFunc(func(params operations.PostParams) middleware.Responder {
		return operations.PostDirektivHandle(params)
	})

	api.PreServerShutdown = func() {}

	api.ServerShutdown = func() {}

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
