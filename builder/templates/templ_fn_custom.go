package {{.Package}}

import (
	"fmt"
	"html/template"

	"github.com/Masterminds/sprig"
	"github.com/direktiv/apps/go/pkg/apps"

{{ imports .DefaultImports }}
)


{{- if eq .Name "Post" }}

func PostDirektivHandle(params PostParams) middleware.Responder {

	actionID := *params.DirektivActionID
	fmt.Printf("run action id: %s\n", actionID)

	return NewPostOK()
}

func HandleShutdown() {
	// called when the function is getting destroyed
}

{{- else }}

func DeleteDirektivHandle(params DeleteParams) middleware.Responder {

	actionID := *params.DirektivActionID
	fmt.Printf("delete action id: %s\n", actionID)

	return NewDeleteOK()
}

{{- end }}

