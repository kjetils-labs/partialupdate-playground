package v2

import (
	"encoding/json"
	"errors"
	"reflect"
)

var (
	ErrNilValue = errors.New("nil value provided")
)

func unmarshalValue(raw json.RawMessage, targetType reflect.Type) (any, error) {

	// If raw is nil or empty, we treat it as a nil value and return the zero value of the target type.
	if len(raw) == 0 {
		return nil, ErrNilValue
	}

	val := reflect.New(targetType).Interface() // Create a pointer to the target type
	if err := json.Unmarshal(raw, &val); err != nil {
		return nil, err
	}

	rval := dereferenceValue(val)
	return rval, nil
}

func dereferenceValue(val any) any {

	rval := reflect.ValueOf(val)
	if rval.Kind() == reflect.Pointer {
		rval = rval.Elem() // val now represents the actual data being pointed to
	}
	if rval.Kind() == reflect.Invalid {
		return nil
	}
	return rval.Interface()
}
