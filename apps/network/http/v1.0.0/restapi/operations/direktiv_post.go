package operations

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/direktiv/apps/go/pkg/apps"
	"github.com/go-openapi/runtime/middleware"

	"http-request/models"
)

const (
	successKey = "success"
	resultKey  = "result"

	// http related
	statusKey  = "status"
	codeKey    = "code"
	headersKey = "headers"
)

var sm sync.Map

const (
	cmdErr = "io.direktiv.command.error"
	outErr = "io.direktiv.output.error"
	riErr  = "io.direktiv.ri.error"
)

type accParams struct {
	PostParams
	Commands []interface{}
}

type accParamsTemplate struct {
	PostBody
	Commands []interface{}
}

func PostDirektivHandle(params PostParams) middleware.Responder {
	fmt.Printf("params in: %+v", params)
	var resp interface{}

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

	var responses []interface{}

	var paramsCollector []interface{}
	accParams := accParams{
		params,
		nil,
	}

	ret, err = runCommand0(ctx, accParams, ri)
	responses = append(responses, ret)

	if err != nil && true {
		errName := cmdErr
		return generateError(errName, err)
	}

	paramsCollector = append(paramsCollector, ret)
	accParams.Commands = paramsCollector

	responseBytes, err := json.Marshal(responses)
	err = json.Unmarshal(responseBytes, &resp)
	if err != nil {
		return generateError(outErr, err)
	}

	return NewPostOK().WithPayload(resp)
}

// http request
func runCommand0(ctx context.Context,
	params accParams, ri *apps.RequestInfo) (map[string]interface{}, error) {

	ri.Logger().Infof("running http request")

	at := accParamsTemplate{
		params.Body,
		params.Commands,
	}

	ir := make(map[string]interface{})
	ir[successKey] = false

	// u, err := templateString(`https://webhook.site/38c796f4-bb9e-4d81-aad5-9e15e3cd5f4f`, at)
	// if err != nil {
	// 	ir[resultKey] = err.Error()
	// 	return ir, err
	// }

	// method, err := templateString(`post`, at)
	// if err != nil {
	// 	ir[resultKey] = err.Error()
	// 	return ir, err
	// }

	// user, err := templateString(`<no value>`, at)
	// if err != nil {
	// 	ir[resultKey] = err.Error()
	// 	return ir, err
	// }

	// password, err := templateString(`<no value>`, at)
	// if err != nil {
	// 	ir[resultKey] = err.Error()
	// 	return ir, err
	// }

	// ins := convertTemplateToBool(`<no value>`, at, false)

	// err200 := true
	// ri.Logger().Infof("%v request to %v", method, u)

	type baseRequest struct {
		url, method, user, password string
		insecure, err200            bool
	}

	baseInfo := func(paramsIn interface{}) (*baseRequest, error) {

		u, err := templateString(`https://webhook.site/38c796f4-bb9e-4d81-aad5-9e15e3cd5f4f`, paramsIn)
		if err != nil {
			return nil, err
		}

		method, err := templateString(`post`, paramsIn)
		if err != nil {
			return nil, err
		}

		user, err := templateString(`<no value>`, paramsIn)
		if err != nil {
			return nil, err
		}

		password, err := templateString(`<no value>`, paramsIn)
		if err != nil {
			return nil, err
		}

		return &baseRequest{
			url:      u,
			method:   method,
			user:     user,
			password: password,
			err200:   convertTemplateToBool(`<no value>`, paramsIn, true),
			insecure: convertTemplateToBool(`<no value>`, paramsIn, false),
		}, nil

	}
	br, err := baseInfo(at)
	if err != nil {
		ir[resultKey] = err.Error()
		return ir, err
	}

	headers := make(map[string]string)
	Header0, err := templateString(`gerke1`, params)
	headers["jens"] = Header0
	Header1, err := templateString(`text/plain`, params)
	headers["content-type"] = Header1

	attachData := func(paramsIn interface{}) ([]byte, error) {

		value, err := templateString(`HELLWORLD {{ .Headers }}`, paramsIn)
		if err != nil {
			return nil, err
		}
		return []byte(value), nil
	}
	data, err := attachData(at)
	if err != nil {
		ir[resultKey] = err.Error()
		return ir, err
	}

	return doHttpRequest(br.method, br.url, br.user, br.password,
		headers, br.insecure, br.err200, data)

	// handle attachment
	// var data []byte
	//
	//
	//
	// value, err = templateString(`map[kind:string value:HELLWORLD {{ .Headers }}]`, params)
	// if err != nil {
	// 	ir[resultKey] = err.Error()
	// 	return ir, err
	// }
	//

	//
	// value, err := templateString(`map[kind:string value:HELLWORLD {{ .Headers }}]`, params)
	// if err != nil {
	// 	ir[resultKey] = err.Error()
	// 	return ir, err
	// }
	// data = []byte(value)
	//

	// return &baseRequest{
	// 	url: u,
	// 	method: method,
	// 	user: user,
	// 	password: password,
	// 	err200: convertTemplateToBool(`<no value>`, paramsIn, true),
	// 	insecure: convertTemplateToBool(`<no value>`, paramsIn, false),
	// }, nil

}

// end commands

func generateError(code string, err error) *PostDefault {

	d := NewPostDefault(0).WithDirektivErrorCode(code).
		WithDirektivErrorMessage(err.Error())

	errString := err.Error()

	errResp := models.Error{
		ErrorCode:    &code,
		ErrorMessage: &errString,
	}

	d.SetPayload(&errResp)

	return d
}

func HandleShutdown() {
	// nothing for generated functions
}
