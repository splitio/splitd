package conf

func SetIfNotNil[T any](dst *T, src *T) {
	if src != nil {
		*dst = *src
	}
}

func SetIfNotEmpty[T comparable](dst *T, src *T) {
	var t T
	if src != nil && *src != t {
		*dst = *src
	}
}

func MapIfNotNil[T any, U any](dst *T, src *U, fn func(U) T) {
	if src != nil {
		*dst = fn(*src)
	}
}

func MapIfNotEmpty[T any, U comparable](dst *T, src *U, fn func(U) T) {
	var u U
	if src != nil && *src != u {
		*dst = fn(*src)
	}
}
