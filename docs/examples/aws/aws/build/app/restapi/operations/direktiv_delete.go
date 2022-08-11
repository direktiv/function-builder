package operations

import (
	"context"
	"fmt"

	"github.com/direktiv/apps/go/pkg/apps"
	"github.com/go-openapi/runtime/middleware"
)

func DeleteDirektivHandle(params DeleteParams) middleware.Responder {

	actionId := *params.DirektivActionID
	// defer sm.Delete(actionId)

	if actionId == "" {
		return NewDeleteOK()
	}

	ri, err := apps.RequestinfoFromRequest(params.HTTPRequest)
	if err != nil {
		fmt.Println("can not create ri from request")
		return NewDeleteOK()
	}

	// cancel, ok := sm.Load(actionId)
	ci, ok := sm.Load(actionId)
	if !ok {
		ri.Logger().Infof("can not load context for action id1", err)
		return NewDeleteOK()
	}

	cinfo, ok := ci.(*ctxInfo)
	if !ok {
		ri.Logger().Infof("can not load context for action id2")
		return NewDeleteOK()
	}

	// set to cancelled
	cinfo.cancelled = true

	ri.Logger().Infof("cancelling action id %v", actionId)
	cinfo.cf()

	cmd, err := templateString("echo 'cancel {{ .DirektivActionID }}'", params)
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
