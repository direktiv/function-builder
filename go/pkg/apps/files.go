package apps

import (
	"context"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/go-openapi/strfmt"
)

type DirektivFile struct {
	Name string `json:"name"`
	Data string `json:"data"`
	Mode string `json:"mode"`
}

func (m *DirektivFile) Validate(formats strfmt.Registry) error {
	return nil
}

func (m *DirektivFile) ContextValidate(ctx context.Context, formats strfmt.Registry) error {

	if m.Name == "" {
		return fmt.Errorf("direktiv file name")
	}

	req, ok := ctx.Value("req").(*http.Request)
	if !ok {
		return fmt.Errorf("no request in context")
	}

	dir := req.Header.Get("Direktiv-TempDir")
	if dir == "" {
		dir = "/tmp"
	}

	f, err := os.Create(filepath.Join(dir, m.Name))
	if err != nil {
		return err
	}
	defer f.Close()

	mode := os.FileMode(0644)
	if m.Mode != "" {
		m, err := strconv.ParseUint(m.Mode, 8, 32)
		if err == nil {
			mode = fs.FileMode(m)
		}
	}

	err = os.Chmod(f.Name(), mode)
	if err != nil {
		return err
	}

	_, err = f.WriteString(m.Data)
	return err

}
