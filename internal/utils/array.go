package utils

//func SafeStringIncludes(s string, arr []string) bool {
//  for _, v := range arr {
//    sBytes := []byte(s)
//    vBytes := []byte(v)
//    if subtle.ConstantTimeCompare(sBytes, vBytes) == 1 {
//      return true
//    }
//  }
//  return false
//}

func PrependFunc(x []func(), y func()) []func() {
	z := append(x, nil)
	copy(z[1:], z)
	z[0] = y
	return z
}
