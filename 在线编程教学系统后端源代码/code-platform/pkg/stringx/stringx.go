package stringx

import (
	"strings"
	"unicode/utf8"
)

func IsLowerEqualThan(s string, maxLength int) bool {
	return len(s) <= 4*maxLength && utf8.RuneCountInString(s) <= maxLength
}

func SliceContains(all []string, s string) bool {
	for _, v := range all {
		if v == s {
			return true
		}
	}
	return false
}

func SliceContainsFold(all []string, s string) bool {
	for _, v := range all {
		if strings.EqualFold(v, s) {
			return true
		}
	}
	return false
}
