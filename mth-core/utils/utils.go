package utils

import "reflect"

// IsNil checks for nil values and nil interfaces.
func IsNil(a interface{}) bool {
	// reflect.ValueOf(a).IsNil() panics if the value's Kind is
	// anything other than Chan, Func, Map, Ptr, Interface or Slice
	defer func() { recover() }()
	return a == nil || reflect.ValueOf(a).IsNil()
}
