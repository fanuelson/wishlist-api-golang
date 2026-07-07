package stringutils

import "strings"

func IsBlank(s string) bool {
	return strings.TrimSpace(s) == ""
}

func IsNotBlank(s string) bool {
	return !IsBlank(s)
}

func RequireNonBlank(s string, err error) error {
	if IsBlank(s) {
		return err
	}
	return nil
}
