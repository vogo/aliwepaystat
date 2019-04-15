package main

import "strings"

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

func BothContainsAny(s1, s2 string, f ...string) bool {
	for _, a := range f {
		if strings.Contains(s1, a) || strings.Contains(s2, a) {
			return true
		}
	}
	return false
}
