//go:build !goexperiment.jsonv2

package tmpljson

import (
	"bytes"
	"regexp"
)

var (
	reLeadingChar = regexp.MustCompile(`"([0-9][^"]*?)":`)
	reHyphen      = regexp.MustCompile(`"([^"]*?)\-([^"]*?)":`)
)

func Normalize(b []byte) ([]byte, error) {
	b = reLeadingChar.ReplaceAll(b, []byte(`"_$1":`))
	b = reHyphen.ReplaceAllFunc(b, func(b []byte) []byte {
		return bytes.ReplaceAll(b, []byte("-"), []byte("_"))
	})
	return b, nil
}
