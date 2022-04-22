package {{.Package}}

import (
	"fmt"
	"html/template"

	"github.com/Masterminds/sprig"
	"github.com/direktiv/apps/go/pkg/apps"

{{ imports .DefaultImports }}
)

func DirektivHandle(params PostParams) middleware.Responder {
	return NewPostOK()
}
