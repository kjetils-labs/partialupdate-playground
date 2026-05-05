package v1

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
)

func applyAdd(update *MongoUpdate, op Operation, field *FieldInfo, path string) error {

	t := field.ReflectionType
	if t == nil {
		return fmt.Errorf("ReflectionType is nil")
	}

	var val any
	var err error
	if op.Value != nil {
		val, err = unmarshalValue(op.Value, field.ReflectionType)
		if err != nil {
			return fmt.Errorf("unmarshal value for add '%s': %w", op.Path, err)
		}
	}

	setField(&update.Set, path, val)
	return nil
}

// applyPush handles "/path/-" with $push
func applyPush(update *MongoUpdate, parentPath string, raw json.RawMessage) error {
	var val any
	if raw != nil {
		if err := json.Unmarshal(raw, &val); err != nil {
			return fmt.Errorf("unmarshal value for push: %w", err)
		}
	}

	if update.Push == nil {
		update.Push = make(bson.M)
	}
	// MongoDB $push expects: { field: { $each: [ ... ] } } or { field: value }
	// For backward compatibility & simplicity, we store as array if possible.
	if _, ok := update.Push[parentPath]; ok {
		// Already exists — append to existing $each if possible,
		// or fallback to overwrite (not ideal, but safe).
		// Better: detect current type and handle carefully.
		return fmt.Errorf("multiple push to same field '%s' not supported without more context", parentPath)
	}
	update.Push[parentPath] = val
	return nil
}

func applyRemove(update *MongoUpdate, path string) error {

	if update.Unset == nil {
		update.Unset = make(bson.M)
	}

	update.Unset[path] = ""
	return nil
}

func applyReplace(update *MongoUpdate, op Operation, field *FieldInfo, path string) error {
	return applyAdd(update, op, field, path)
}

func setField(m *bson.M, path string, val any) {

	(*m)[path] = val
}

// convertJSONPathMongo converts a JSON path like /data/to/something to a mongo path.
// Example:
//
// JSON: /test/data
// Mongo: test.data
func convertJSONPathMongo(path string) (string, error) {

	if !strings.HasPrefix(path, "/") {
		return "", fmt.Errorf("invalid JSON pointer: %s", path)
	}

	output := unescapeRFC6902(path)
	output = strings.TrimPrefix(output, "/")
	output = strings.ReplaceAll(output, "/", ".")

	return output, nil
}

var (
	errNotAppendAction = errors.New("not a numeric index or '-'")
)

// parseIndex determines if `last` is a numeric index or "-" (append)
// Returns (isIndex, isNumericIndex, isAppend, error)
func parseIndex(s string) (isIndex bool, isNumeric bool, isAppend bool, err error) {
	if s == "-" {
		return true, false, true, nil
	}
	if _, err := strconv.Atoi(s); err == nil {
		return true, true, false, nil
	}
	return false, false, false, errNotAppendAction
}

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
