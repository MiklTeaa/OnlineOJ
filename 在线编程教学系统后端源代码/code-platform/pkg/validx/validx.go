package validx

import (
	"strconv"
	"strings"
)

func CheckUserData(number, name, organization string) bool {
	if strings.TrimSpace(number) == "" || strings.TrimSpace(organization) == "" || strings.TrimSpace(name) == "" {
		return false
	}
	if _, err := strconv.Atoi(number); err != nil {
		return false
	}
	return true
}
