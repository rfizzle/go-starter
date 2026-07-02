package utils

import "crypto/subtle"

func SafeStringIncludes(s string, arr []string) bool {
	for _, v := range arr {
		sBytes := []byte(s)
		vBytes := []byte(v)
		if subtle.ConstantTimeCompare(sBytes, vBytes) == 1 {
			return true
		}
	}
	return false
}
