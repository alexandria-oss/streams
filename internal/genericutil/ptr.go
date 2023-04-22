package genericutil

// SafeDerefPtr safely dereferences `ptr`. Returns T zero-value if nil.
func SafeDerefPtr[T any](ptr *T) T {
	if ptr != nil {
		return *ptr
	}

	var zeroVal T
	return zeroVal
}
