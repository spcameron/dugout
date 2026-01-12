package testutil

import "reflect"

// Equaler is an interface for types with an Equal method
// (like time.Time or net.IP).
type Equaler[T any] interface {
	Equal(T) bool
}

// IsEqual prefers a type's Equal method when present; otherwise it falls back to DeepEqual.
func IsEqual[T any](got, want T) bool {
	if IsNil(got) && IsNil(want) {
		return true
	}

	if eq, ok := any(got).(Equaler[T]); ok {
		return eq.Equal(want)
	}
	if eq, ok := any(want).(Equaler[T]); ok {
		return eq.Equal(got)
	}

	return reflect.DeepEqual(got, want)
}

// IsNil first checks for nil equality, then uses reflection to check typed nil inside an interface.
func IsNil(v any) bool {
	if v == nil {
		return true
	}

	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice, reflect.UnsafePointer:
		return rv.IsNil()
	}

	return false
}
