package deps

import (
	"fmt"
	"reflect"
)

// Type returns the reflect.Type for T.
//
// This function is particularly useful when T is an interface
// and it is impossible to get a value with concrete type T.
func Type[T any]() reflect.Type {
	return reflect.TypeOf((*T)(nil)).Elem()
}

// IfaceType returns the reflect.Type for the interface.
func IfaceType[T any]() (reflect.Type, error) {
	ifaceType := Type[T]()
	if ifaceType.Kind() != reflect.Interface {
		return nil, fmt.Errorf("%s is not an interface type", ifaceType.Name())
	}
	return ifaceType, nil
}

// StructType returns the reflect.Type for the struct.
func StructType[T any]() (reflect.Type, error) {
	implType := Type[T]()
	if implType.Kind() != reflect.Struct {
		return nil, fmt.Errorf("%s is not a struct type", implType.Name())
	}
	return implType, nil
}
