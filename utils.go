// Copyright 2019 wongoo. All rights reserved.

package aliwepaystat

import (
	"regexp"
	"strings"
)

func Contains(s string, f string) bool {
	return strings.Contains(s, f)
}

func ContainsAny(s string, f ...string) bool {
	for _, a := range f {
		if strings.Contains(s, a) {
			return true
		}
	}
	return false
}

func EitherContainsAny(s1, s2 string, f ...string) bool {
	for _, a := range f {
		if strings.Contains(s1, a) || strings.Contains(s2, a) {
			return true
		}
	}
	return false
}

var (
	regexCsvLineFieldsSuffixBlank, _ = regexp.Compile("[ ]+,")
)

func replaceCsvLineFieldsSuffixBlank(bytes []byte) []byte {
	return regexCsvLineFieldsSuffixBlank.ReplaceAll(bytes, []byte{','})
}
