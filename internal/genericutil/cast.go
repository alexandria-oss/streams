package genericutil

func SafeCast[T any](arg any) T {
	var zeroVal T
	val, ok := arg.(T)
	if !ok {
		return zeroVal
	}
	return val
}
