package gitnotify

import "strings"

// StringIn finds a key in a list of strings
func StringIn(list []string, key string) bool {
	for _, k := range list {
		if key == k {
			return true
		}
	}
	return false
}

func isValidEmail(email string) bool {
	if email == "" || strings.Contains(email, "@users.noreply.github.com") {
		return false
	}
	return true
}
