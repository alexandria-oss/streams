package typeutils

// Coalesce retrieves first non-empty value.
func Coalesce[T comparable](vv ...T) T {
	var zeroVal T
	for _, v := range vv {
		if v != zeroVal {
			return v
		}
	}
	return zeroVal
}
