package utils

func Must[T any](v T, err error) T {
	Check(err)
	return v
}

func Check(err error) {
	if err != nil {
		panic(err)
	}
}
