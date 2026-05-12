package v2

import (
	"fmt"
	"log/slog"
	"reflect"
	"strings"
)

type FieldInfo struct {
	Path           string
	ReflectionType reflect.Type
	Name           string
	Value          any
}

func WalkStruct(path map[string]any, str any) ([]*FieldInfo, error) {
	if path == nil {
		return nil, fmt.Errorf("path cannot be empty")
	}

	rtype := reflect.TypeOf(str)
	if rtype.Kind() == reflect.Pointer {
		rtype = rtype.Elem()
	}

	return traversePath(path, rtype)

}

func traversePath(pathmap map[string]any, rtype reflect.Type) ([]*FieldInfo, error) {

	output := make([]*FieldInfo, 0)

	if rtype.Kind() == reflect.Pointer {
		rtype = rtype.Elem()
	}

	for k, v := range pathmap {
		out, err := getFieldPath(k, v, rtype, "", 0)
		if err != nil {
			slog.Error("failed to get field path for key", "key", k, "error", err)
			continue
		}
		output = append(output, out...)
	}

	return output, nil
}

func getFieldPath(k, v any, rtype reflect.Type, mpath string, depth int) ([]*FieldInfo, error) {

	output := make([]*FieldInfo, 0)

	for field := range rtype.Fields() {

		name := getFieldName(field)
		if name != k {
			continue
		}

		if isFieldPointer(field) {
			// If the value is nil, we assume the pointer field is being set to nil, so we return the field info with a nil value.
			if v == nil {
				addFieldInfo(&output, mpath, field.Type, name, nil)
				continue
			}
			if value, ok := v.(map[string]any); ok {
				if mpath == "" {
					mpath += name
				} else {
					mpath += "." + name
				}
				for mk, mv := range value {
					if mv != nil {
						out, err := getFieldPath(mk, mv, field.Type.Elem(), mpath, depth+1)
						if err != nil {
							slog.Error("failed to get field path for pointer field", "field_name", name, "error", err)
							continue
						}
						output = append(output, out...)
					} else {
						addFieldInfo(&output, mpath+"."+mk, field.Type, name, nil)
						continue
					}
				}
			}

			out, err := getFieldPath(k, v, field.Type.Elem(), mpath, depth+1)
			if err != nil {
				slog.Error("failed to get field path for pointer field", "field_name", name, "error", err)
				continue
			}
			output = append(output, out...)
			continue
		}

		if mpath == "" {
			mpath += name
		} else {
			mpath += "." + name
		}

		// if the field is a pointer, we need to dereference it and continue searching for the path in the underlying type.
		if isSlice(field) {

		}

		// if the field is a map, we need to key-value pair and continue searching for the path in the value type.
		if isMap(field) || isStruct(field) {

			// If the value is nil, we assume the struct field is being set to nil, so we return the field info with a nil value.
			if v == nil {
				addFieldInfo(&output, mpath, field.Type, name, nil)
				continue
			}

			if value, ok := v.(map[string]any); ok {
				for mk, mv := range value {
					if mv == nil {
						addFieldInfo(&output, mpath, field.Type, mk, nil)
						continue
					}

					out, err := getFieldPath(mk, mv, field.Type, mpath, depth+1)
					if err != nil {
						slog.Error("failed to get field path for map field", "field_name", name, "error", err)
						continue
					}
					output = append(output, out...)
				}
			} else {
				slog.Error("expected a map for map field", "field_name", name, "value_type", fmt.Sprintf("%T", v))
			}
			continue
		}

		// if the field is a base type, we need to check if the path is empty (i.e. we have reached the end of the path) and return the field info if so.
		if isBaseType(field) {

			if v == nil {
				addFieldInfo(&output, mpath, field.Type, name, nil)
				continue
			}

			addFieldInfo(&output, mpath, field.Type, name, v)
		}

	}

	return output, nil
}

// addFieldInfo is a helper function to create a FieldInfo struct and append it to the output slice.
func addFieldInfo(output *[]*FieldInfo, mpath string, field reflect.Type, name string, value any) {

	out := &FieldInfo{
		Path:           mpath,
		ReflectionType: field,
		Name:           name,
		Value:          value,
	}
	*output = append(*output, out)
}

func splitPath(p string) []string {
	return strings.FieldsFunc(p, func(r rune) bool {
		return r == '/'
	})
}

// getFieldName creates the name from the struct field.
// This is either the json tag or the struct field name if no json tag is provided.
func getFieldName(field reflect.StructField) string {
	name := ""
	// quick and dirty way to get the first part of the json tag, which is
	// the name
	tag := strings.Split(field.Tag.Get("json"), ",")[0]

	// use the structfield name as the path
	if tag == "" {
		name = field.Name
	} else {
		name = tag
	}

	return name
}

func isFieldPointer(rtype reflect.StructField) bool {
	return rtype.Type.Kind() == reflect.Pointer
}

func isValuePointer(rtype reflect.Value) bool {
	return rtype.Kind() == reflect.Pointer
}

func isStruct(rtype reflect.StructField) bool {
	return rtype.Type.Kind() == reflect.Struct
}

func isInline(field reflect.StructField) bool {
	return field.Anonymous
}

func isBaseType(rtype reflect.StructField) bool {
	kind := rtype.Type.Kind()
	return kind != reflect.Struct && kind != reflect.Pointer
}

func isMap(rtype reflect.StructField) bool {
	return rtype.Type.Kind() == reflect.Map
}

func isSlice(rtype reflect.StructField) bool {
	return rtype.Type.Kind() == reflect.Slice
}
