package v1

import (
	"fmt"
	"log/slog"
	"reflect"
	"strconv"
	"strings"
)

type FieldInfo struct {
	Path           string
	ReflectionType reflect.Type
	Name           string
}

func WalkStruct(path string, str any) (*FieldInfo, error) {
	if path == "" {
		return nil, fmt.Errorf("path cannot be empty")
	}

	t := reflect.TypeOf(str)
	return walkStruct(path, t)
}

func walkStruct(path string, str reflect.Type) (*FieldInfo, error) {

	// Check if the path is just the root ("/") and return the struct type if so.
	root := getOnlyRootField(path, str)
	if root != nil {
		return root, nil
	}

	paths := splitPath(path)

	if len(paths) == 0 {
		return nil, fmt.Errorf("invalid path: %q", path)
	}

	return getFieldPath(paths, str, "/")
}
func getFieldPath(paths []string, rtype reflect.Type, jpath string) (*FieldInfo, error) {

	if len(paths) == 0 {
		return nil, fmt.Errorf("path cannot be empty")
	}
	path := paths[0]
	for field := range rtype.Fields() {

		name := getFieldName(field)

		if name != path {
			continue
		}

		slog.Debug("field status", "field_name", name, "is_pointer", isFieldPointer(field), "is_struct", isStruct(field), "is_inline", isInline(field), "is_base_type", isBaseType(field), "paths_left", len(paths)-1)

		// if the field is a pointer, we need to dereference it and continue searching for the path in the underlying type.
		if isFieldPointer(field) {
			ptr := field.Type.Elem()

			if len(paths)-1 == 0 {
				jpath += fmt.Sprintf("%s", path)
				out := &FieldInfo{
					Path:           jpath,
					ReflectionType: ptr,
					Name:           path,
				}
				return out, nil
			}

			jpath += fmt.Sprintf("%s/", path)
			return getFieldPath(paths[1:], ptr, jpath)
		}

		// if the field is a struct, we need to continue searching for the path in the struct type.
		if isStruct(field) {

			// if the field is inline, we need to continue searching for the path in the struct type without adding the field name to the path.
			if isInline(field) {
				return getFieldPath(paths[1:], field.Type, jpath)
			}

			jpath += fmt.Sprintf("%s/", path)
			return getFieldPath(paths[1:], field.Type, jpath)
		}

		if isSlice(field) {

			// if the slice field is the end of the path, we can return the field info with the slice type.
			if len(paths) == 1 {
				jpath += fmt.Sprintf("%s", path)
				out := &FieldInfo{
					Path:           jpath,
					ReflectionType: field.Type,
					Name:           path,
				}

				return out, nil
			}

			// if the next path component is "-", we need to check if it's the end of the path and return the field info with the slice element type.
			// If it's not the end of the path, we return an error since - is only allowed at the end of the path for appending to a slice.
			if paths[1] == "-" {
				if len(paths) > 2 {
					return nil, fmt.Errorf("- operation is only allowed at the end of the path")
				}
				jpath += fmt.Sprintf("%s/%v", path, paths[1])
				value := field.Type.Elem()
				out := &FieldInfo{
					Path:           jpath,
					ReflectionType: value,
					Name:           path,
				}
				return out, nil
			} else {
				index, err := strconv.ParseInt(paths[1], 10, 64)
				if err != nil {
					return nil, fmt.Errorf("component %q is not a valid index", paths[1])
				}
				jpath += fmt.Sprintf("%s/%d/", path, index)
			}

			value := field.Type.Elem()

			if !isBaseType(field) {
				return getFieldPath(paths[1:], value, jpath)
			}

			out := &FieldInfo{
				Path:           jpath,
				ReflectionType: value,
				Name:           path,
			}

			return out, nil
		}

		// if the field is a map, we need to key-value pair and continue searching for the path in the value type.
		if isMap(field) {

			// If the map field is the end of the path, we can return the field info with the map type.
			if len(paths) == 1 {
				jpath += fmt.Sprintf("%s", name)
				out := &FieldInfo{
					Path:           jpath,
					ReflectionType: field.Type,
					Name:           name,
				}
				return out, nil
			}

			value := field.Type.Elem()
			if value.Kind() == reflect.Pointer {
				value = value.Elem()
			}

			if len(paths) > 2 {
				// TODO: check if the key is valid for the map type (i.e. if the key is a string and the map key type is string)
				jpath += fmt.Sprintf("%s/%s/", name, paths[1])
				return getFieldPath(paths[2:], value, jpath)
			}

			jpath += fmt.Sprintf("%s/%s", name, paths[1])
			out := &FieldInfo{
				Path:           jpath,
				ReflectionType: value,
				Name:           name,
			}

			return out, nil
		}

		// if the field is a base type, we need to check if the path is empty (i.e. we have reached the end of the path) and return the field info if so.
		if isBaseType(field) {

			jpath += fmt.Sprintf("%s", name)
			out := &FieldInfo{
				Path:           jpath,
				ReflectionType: field.Type,
				Name:           name,
			}

			return out, nil
		}

	}

	return nil, fmt.Errorf("field %q does not exist in type %q", paths[0], rtype.Name())
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

func getOnlyRootField(path string, rtype reflect.Type) *FieldInfo {
	if path == "/" {
		return &FieldInfo{
			Path:           path,
			ReflectionType: rtype,
			Name:           rtype.Name(),
		}
	}

	return nil
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
