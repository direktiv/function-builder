package apps

import (
	"encoding/json"
	"strings"

	"github.com/acarl005/stripansi"
)

func ToJSON(str string) interface{} {

	str = strings.TrimSpace(str)
	str = stripansi.Strip(str)

	var js json.RawMessage
	err := json.Unmarshal([]byte(str), &js)
	if err != nil {
		return str
	}

	return json.RawMessage(str)

}
