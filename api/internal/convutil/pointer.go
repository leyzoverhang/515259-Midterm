package convutil

func ToSafeValue[T any](ptr *T) T {
	var zero T

	if ptr == nil {
		return zero
	}

	return *ptr
}

func ToPointer[T comparable](value T) *T {
	var zero T

	if value == zero {
		return nil
	}

	return &value
}
