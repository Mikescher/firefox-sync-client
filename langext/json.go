package langext

import (
	"bytes"
	"encoding/json"
)

type H map[string]any

type A []any

func TryPrettyPrintJson(str string) string {
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, []byte(str), "", "    "); err != nil {
		return str
	}
	return prettyJSON.String()
}
