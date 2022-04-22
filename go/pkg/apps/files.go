package apps

import (
	"context"
	"io/fs"
	"os"
	"strconv"

	"github.com/go-openapi/strfmt"
)

type DirektivFile struct {
	Name string `json:"name"`
	Data string `json:"data"`
	Mode string `json:"mode"`
}

func (m *DirektivFile) Validate(formats strfmt.Registry) error {

	f, err := os.Create(m.Name)
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

func (m *DirektivFile) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}
