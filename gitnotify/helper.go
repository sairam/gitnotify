package gitnotify

// StringIn finds a key in a list of strings
func StringIn(list []string, key string) bool {
	for _, k := range list {
		if key == k {
			return true
		}
	}
	return false
}
