package main

var cache = make(map[string]any)

func fetch[T any](key string, fn func() (T, error)) (T, error) {
	res, ok := cache[key]
	if ok {
		return res.(T), nil
	}

	res, err := fn()
	if err != nil {
		var z T
		return z, err
	}

	return res.(T), nil
}
