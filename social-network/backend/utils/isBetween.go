package utils

func IsBetween(str string, x int, y int) bool {
	if len(str) >= x && len(str) <= y {
		return true
	}
	return false
}
