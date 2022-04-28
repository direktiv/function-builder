package operations

import (
	"context"
	"encoding/json"
	"strings"
	"sync"

	"github.com/direktiv/apps/go/pkg/apps"
	"github.com/go-openapi/runtime/middleware"

	"aws-cli/models"
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

func PostDirektivHandle(params PostParams) middleware.Responder {
	var resp interface{}

	var (
		err error
		ret map[string]interface{}
	)

	ri, err := apps.RequestinfoFromRequest(params.HTTPRequest)
	if err != nil {
		return generateError(riErr, err)
	}

	ctx, cancel := context.WithCancel(params.HTTPRequest.Context())
	sm.Store(*params.DirektivActionID, cancel)
	defer sm.Delete(params.DirektivActionID)

	responses := make(map[string]map[string]interface{}, 1)

	ret, err = runCommand0(ctx, params, ri)
	responses["cmd0"] = ret

	if err != nil && true {
		errName := cmdErr
		return generateError(errName, err)
	}

	s, err := templateString(`{
  "greeting": "{{ index (index . "cmd0") "result" }}"
}
`, responses)
	if err != nil {
		return generateError(outErr, err)
	}
	responseBytes := []byte(s)

	err = json.Unmarshal(responseBytes, &resp)
	if err != nil {
		return generateError(outErr, err)
	}

	return NewPostOK().WithPayload(&resp)
}

// start commands

// exec
func runCommand0(ctx context.Context,
	params PostParams, ri *apps.RequestInfo) (map[string]interface{}, error) {

	ir := make(map[string]interface{})
	ir[successKey] = false

	ri.Logger().Infof("executing command")

	cmd, err := templateString(`echo '{{ fileExists "jens" }}'`, params.Body)
	if err != nil {
		ir[resultKey] = err.Error()
		return ir, err
	}
	cmd = strings.Replace(cmd, "\n", "", -1)

	silent := false
	print := true
	output := ""

	envs := []string{}

	return runCmd(ctx, cmd, envs, output, silent, print, ri)

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
