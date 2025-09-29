//go:build go1.25 && goexperiment.jsonv2

package tmpljson

import (
	"bytes"
	"encoding/json/jsontext"
	"io"
	"strings"
)

func Normalize(b []byte) ([]byte, error) {
	var buf bytes.Buffer
	var stack jsontext.Pointer

	dec := jsontext.NewDecoder(bytes.NewReader(b))
	enc := jsontext.NewEncoder(&buf, jsontext.Multiline(true))

	for {
		tok, err := dec.ReadToken()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		if tok.Kind() == '"' {
			if ptr := dec.StackPointer(); ptr != stack {
				if s := tok.String(); len(s) > 0 && s == ptr.LastToken() {
					s = strings.ReplaceAll(s, "-", "_")
					if '0' <= s[0] && s[0] <= '9' {
						s = "_" + s
					}
					tok = jsontext.String(s)
				}
				stack = ptr
			}
		}

		if err = enc.WriteToken(tok); err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
}
