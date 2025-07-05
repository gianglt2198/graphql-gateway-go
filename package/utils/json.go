package utils

import (
	"encoding/json"
	"io"
)

func JsonDecode(r io.Reader, val any) error {
	dec := json.NewDecoder(r)
	dec.UseNumber()
	return dec.Decode(val)
}
