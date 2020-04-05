// Copyright 2019 wongoo. All rights reserved.

package aliwepaystat

import (
	"fmt"
	"regexp"
	"strconv"
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

func RoundFloat(f float64) float64 {
	v, _ := strconv.ParseFloat(fmt.Sprintf("%.4f", f), 64)
	return v
}

var (
	investmentRegex1, _ = regexp.Compile(".*财富.*买入.*")
	investmentRegex2, _ = regexp.Compile(".*基金.*买入.*")
	investmentRegex3, _ = regexp.Compile(".*股票.*买入.*")
	investmentRegex4, _ = regexp.Compile(".*余额宝.*转入.*")
)

func IsInvestment(s string) bool {
	return investmentRegex1.MatchString(s) ||
		investmentRegex2.MatchString(s) ||
		investmentRegex3.MatchString(s) ||
		investmentRegex4.MatchString(s)
}
