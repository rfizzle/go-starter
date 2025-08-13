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

func PrependFunc(x []func(), y func()) []func() {
	x = append(x, nil)
	copy(x[1:], x)
	x[0] = y
	return x
}
