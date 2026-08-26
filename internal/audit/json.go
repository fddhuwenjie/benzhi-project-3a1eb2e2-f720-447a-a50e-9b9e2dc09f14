package audit

import (
	"bytes"
	"encoding/json"
)

func jsonUnmarshalNumber(data []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	return decoder.Decode(target)
}
