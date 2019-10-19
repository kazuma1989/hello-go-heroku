package main

import (
	"io/ioutil"
	"strings"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

// ConvertEUCJP converts EUC-JP string to UTF-8 string
func ConvertEUCJP(text string) string {
	decoded, _ := ioutil.ReadAll(transform.NewReader(
		strings.NewReader(text),
		japanese.EUCJP.NewDecoder(),
	))

	return string(decoded)
}
