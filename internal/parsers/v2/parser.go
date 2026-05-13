package v2

import (
	"encoding/json"
	"fmt"
	"log/slog"
	v1 "partialupdate/internal/parsers/v1"
	"reflect"
	"strconv"
)

// ParsePatch takes a struct and converts it to a map of JSON raw messages, which can then be used to construct a MongoDB update document. The generic type T allows for flexibility in the input struct type.
// It walks through the input struct, identifies which fields are being updated (non-nil values) and which are being unset (nil values), and constructs a MongoUpdate with the appropriate $set and $unset operations.
// The function also handles type conversion for numeric and string fields, ensuring that the values in the MongoUpdate match the expected types defined in the struct.
//
// It does not support appending to arrays (e.g. $push) or removing from arrays (e.g. $pull) in this implementation, but it can be extended to do so in the future if needed.
func ParsePatch[T any](patch []byte) (*v1.MongoUpdate, error) {

	var input map[string]any
	err := json.Unmarshal(patch, &input)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal patch JSON: %w", err)
	}

	fields, err := WalkStruct(input, new(T))
	if err != nil {
		return nil, fmt.Errorf("failed to walk struct for fields: %w", err)
	}

	output := v1.NewMongoUpdate()
	for _, field := range fields {

		if field.Value == nil {
			output.Unset[field.Path] = ""
			continue
		}

		if reflect.TypeOf(field.Value) == field.ReflectionType {
			slog.Info("type matches, no conversion needed",
				"field", field.Path,
				"value", field.Value,
				"expected_type", field.ReflectionType,
				"actual_type", reflect.TypeOf(field.Value),
			)
			output.Set[field.Path] = field.Value
			continue
		}

		if field.ReflectionType.Kind() == reflect.Slice {
			val := make([]any, 0)
			slice, ok := field.Value.([]any)
			if !ok {
				return nil, fmt.Errorf("expected slice for field '%s', got %T", field.Path, field.Value)
			}
			for i, elem := range slice {
				convertedElem, err := convertToType(elem, field.ElemType)
				if err != nil {
					return nil, fmt.Errorf("failed to convert element %d for field '%s': %w", i, field.Path, err)
				}
				val = append(val, convertedElem)
			}
			output.Set[field.Path] = val
			continue
		}

		if field.ReflectionType.Kind() == reflect.Struct {

		}

		val, err := convertToType(field.Value, field.ReflectionType)
		if err != nil {
			return nil, fmt.Errorf("failed to convert value for field '%s': %w", field.Path, err)
		}
		output.Set[field.Path] = val
	}

	return output, nil
}

// convertToType attempts to convert a value of any type to a specified target type using reflection.
// It handles common conversions for numeric types and strings, and falls back to standard Go conversions when possible.
// TODO: This is AI generated and should be thoroughly tested and reviewed for edge cases and potential issues with type conversion, especially for complex types or when dealing with pointers and interfaces.
func convertToType(val any, target reflect.Type) (any, error) {
	if target == nil {
		return nil, fmt.Errorf("target type cannot be nil")
	}
	if val == nil {
		return nil, fmt.Errorf("value cannot be nil")
	}

	rv := reflect.ValueOf(val)
	if !rv.IsValid() {
		return nil, fmt.Errorf("invalid reflect.Value")
	}

	// If target is a pointer, unwrap it to get the underlying type for conversion logic
	if target.Kind() == reflect.Pointer {
		target = target.Elem()
	}

	convertTo := func(v any, t reflect.Type) (any, error) {
		valRef := reflect.ValueOf(v)
		if !valRef.Type().ConvertibleTo(t) {
			return nil, fmt.Errorf("cannot convert %T to %v", v, t)
		}
		return valRef.Convert(t).Interface(), nil
	}

	switch target.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		var i64 int64
		switch kind := rv.Kind(); kind {
		case reflect.Float64:
			i64 = int64(rv.Float())
		case reflect.String:
			v, err := strconv.ParseInt(rv.String(), 10, 64)
			if err != nil {
				return nil, fmt.Errorf("parsing string '%s' as int: %w", rv.String(), err)
			}
			i64 = v
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			i64 = rv.Int()
		default:
			return nil, fmt.Errorf("unsupported source kind %v for target int", kind)
		}
		return convertTo(i64, target)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		var u64 uint64
		switch kind := rv.Kind(); kind {
		case reflect.Float64:
			u64 = uint64(rv.Float())
		case reflect.String:
			v, err := strconv.ParseUint(rv.String(), 10, 64)
			if err != nil {
				return nil, fmt.Errorf("parsing string '%s' as uint: %w", rv.String(), err)
			}
			u64 = v
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			u64 = rv.Uint()
		default:
			return nil, fmt.Errorf("unsupported source kind %v for target uint", kind)
		}
		return convertTo(u64, target)

	case reflect.Float32, reflect.Float64:
		var f64 float64
		switch kind := rv.Kind(); kind {
		case reflect.Float64:
			f64 = rv.Float()
		case reflect.String:
			v, err := strconv.ParseFloat(rv.String(), 64)
			if err != nil {
				return nil, fmt.Errorf("parsing string '%s' as float: %w", rv.String(), err)
			}
			f64 = v
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			f64 = float64(rv.Int())
		default:
			return nil, fmt.Errorf("unsupported source kind %v for target float", kind)
		}
		return convertTo(f64, target)

	case reflect.String:
		switch kind := rv.Kind(); kind {
		case reflect.String:
			return rv.Interface(), nil
		case reflect.Float64:
			// Safe to format any float as string
			return strconv.FormatFloat(rv.Float(), 'f', -1, 64), nil
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return fmt.Sprintf("%d", rv.Int()), nil
		default:
			return nil, fmt.Errorf("unsupported source kind %v for target string", kind)
		}

	default:
		if rv.Type().ConvertibleTo(target) {
			return rv.Convert(target).Interface(), nil
		}
		return nil, fmt.Errorf("cannot convert %v to %v", rv.Type().Kind(), target.Kind())
	}
}
