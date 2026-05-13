package v2

import (
	"fmt"
	"reflect"
	"strings"
)

type (
	fieldInfoOption func(*FieldInfo)
	FieldInfo       struct {
		Path           string
		ReflectionType reflect.Type
		ElemType       reflect.Type
		Name           string
		Value          any
	}
)

func NewFieldInfo(options ...fieldInfoOption) *FieldInfo {
	fi := &FieldInfo{
		Path:           "",
		ReflectionType: nil,
		ElemType:       nil,
		Name:           "",
		Value:          nil,
	}

	for _, option := range options {
		option(fi)
	}

	return fi
}

func FieldInfoOptionPath(path string) fieldInfoOption {
	return func(fi *FieldInfo) {
		fi.Path = path
	}
}

func FieldInfoOptionReflectionType(rtype reflect.Type) fieldInfoOption {
	return func(fi *FieldInfo) {
		fi.ReflectionType = rtype
	}
}

func FieldInfoOptionElemType(elemType reflect.Type) fieldInfoOption {
	return func(fi *FieldInfo) {
		fi.ElemType = elemType
	}
}

func FieldInfoOptionName(name string) fieldInfoOption {
	return func(fi *FieldInfo) {
		fi.Name = name
	}
}

func FieldInfoOptionValue(value any) fieldInfoOption {
	return func(fi *FieldInfo) {
		fi.Value = value
	}
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
			return nil, err
		}
		output = append(output, out...)
	}

	return output, nil
}

func getFieldPath(k, v any, rtype reflect.Type, mpath string, depth int) ([]*FieldInfo, error) {

	output := make([]*FieldInfo, 0)

	for field := range rtype.Fields() {

		name := getFieldName(field)

		// If the field is inline, we need to continue searching for the path in the underlying type without adding the field name to the path.
		// If a match is found in the inline struct, we can return immediately since there are no further paths to traverse.
		// If no match is found in the inline struct, we continue searching for the path in the current struct.
		if isInline(field) {
			out, err := getFieldPath(k, v, field.Type, mpath, depth+1)
			if err != nil {
				continue
			}
			output = append(output, out...)
			return output, nil
		}

		if name != k {
			continue
		}

		// TODO: Can probably just check v for nil and it doesn't matter
		if isFieldPointer(field) {
			if v == nil {
				addFieldInfo(&output,
					FieldInfoOptionPath(mpath+"."+name),
					FieldInfoOptionReflectionType(field.Type),
					FieldInfoOptionName(name),
				)
				return output, nil
			}

			// If the pointer is to a base type, we can add the field info and return immediately since there are no further paths to traverse.
			if field.Type.Elem().Kind() != reflect.Struct && field.Type.Elem().Kind() != reflect.Map {
				addFieldInfo(&output,
					FieldInfoOptionPath(mpath+"."+name),
					FieldInfoOptionReflectionType(field.Type.Elem()),
					FieldInfoOptionName(name),
					FieldInfoOptionValue(v),
				)
				return output, nil
			}

			if value, ok := v.(map[string]any); ok {
				if mpath == "" {
					mpath += name
				} else {
					mpath += "." + name
				}
				for mk, mv := range value {
					if mv == nil {
						addFieldInfo(&output,
							FieldInfoOptionPath(mpath+"."+mk),
							FieldInfoOptionReflectionType(field.Type),
							FieldInfoOptionName(name),
						)
						continue
					}
					out, err := getFieldPath(mk, mv, field.Type.Elem(), mpath, depth+1)
					if err != nil {
						return nil, err
					}
					output = append(output, out...)
				}
				return output, nil
			}

			out, err := getFieldPath(k, v, field.Type.Elem(), mpath, depth+1)
			if err != nil {
				return nil, err
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
			if v == nil {
				addFieldInfo(&output,
					FieldInfoOptionPath(mpath),
					FieldInfoOptionReflectionType(field.Type),
					FieldInfoOptionName(name),
				)
				return output, nil
			}

			addFieldInfo(&output,
				FieldInfoOptionPath(mpath),
				FieldInfoOptionReflectionType(field.Type),
				FieldInfoOptionElemType(field.Type.Elem()),
				FieldInfoOptionName(name),
				FieldInfoOptionValue(v),
			)
			return output, nil
		}

		if isMap(field) {
			if value, ok := v.(map[string]any); ok {
				for mk, mv := range value {
					if mv == nil {
						addFieldInfo(&output,
							FieldInfoOptionPath(mpath),
							FieldInfoOptionReflectionType(field.Type),
						)
						continue
					}

					if field.Type.Elem().Kind() != reflect.Struct && field.Type.Elem().Kind() != reflect.Map {
						addFieldInfo(&output,
							FieldInfoOptionPath(mpath+"."+mk),
							FieldInfoOptionReflectionType(field.Type.Elem()),
							FieldInfoOptionName(name),
							FieldInfoOptionValue(mv),
						)
						continue
					}
					out, err := getFieldPath(mk, mv, field.Type, mpath, depth+1)
					if err != nil {
						continue
					}
					output = append(output, out...)
				}
			} else {
				return nil, fmt.Errorf("expected map[string]any for struct field '%s', got %T", name, v)
			}
			return output, nil
		}

		// if the field is a map, we need to key-value pair and continue searching for the path in the value type.
		if isStruct(field) {

			// If the value is nil, we assume the struct field is being set to nil, so we return the field info with a nil value.
			if v == nil {
				addFieldInfo(&output,
					FieldInfoOptionPath(mpath),
					FieldInfoOptionReflectionType(field.Type),
					FieldInfoOptionName(name),
				)
				return output, nil
			}

			if value, ok := v.(map[string]any); ok {
				for mk, mv := range value {
					if mv == nil {
						addFieldInfo(&output,
							FieldInfoOptionPath(mpath),
							FieldInfoOptionReflectionType(field.Type),
						)
						return output, nil
					}

					out, err := getFieldPath(mk, mv, field.Type, mpath, depth+1)
					if err != nil {
						continue
					}
					output = append(output, out...)
				}
			} else {
				return nil, fmt.Errorf("expected map[string]any for struct field '%s', got %T", name, v)
			}
			return output, nil
		}

		// if the field is a base type, we need to check if the path is empty (i.e. we have reached the end of the path) and return the field info if so.
		if isBaseType(field) {
			addFieldInfo(&output,
				FieldInfoOptionPath(mpath),
				FieldInfoOptionReflectionType(field.Type),
				FieldInfoOptionName(name),
				FieldInfoOptionValue(v),
			)
			return output, nil
		}

	}

	return output, fmt.Errorf("field '%v' not found in struct", k)
}

// addFieldInfo is a helper function to create a FieldInfo struct and append it to the output slice.
func addFieldInfo(output *[]*FieldInfo, options ...fieldInfoOption) {

	out := NewFieldInfo(options...)
	*output = append(*output, out)
}

func pointerHandler(output *[]*FieldInfo, mpath, name string, field reflect.StructField, depth int, k, v any) error {
	if v == nil {
		addFieldInfo(output,
			FieldInfoOptionPath(mpath+"."+name),
			FieldInfoOptionReflectionType(field.Type),
			FieldInfoOptionName(name),
		)
		return nil
	}

	// If the pointer is to a base type, we can add the field info and return immediately since there are no further paths to traverse.
	if field.Type.Elem().Kind() != reflect.Struct && field.Type.Elem().Kind() != reflect.Map {
		addFieldInfo(output,
			FieldInfoOptionPath(mpath+"."+name),
			FieldInfoOptionReflectionType(field.Type.Elem()),
			FieldInfoOptionName(name),
			FieldInfoOptionValue(v),
		)
		return nil
	}

	if value, ok := v.(map[string]any); ok {
		if mpath == "" {
			mpath += name
		} else {
			mpath += "." + name
		}
		for mk, mv := range value {
			if mv == nil {
				addFieldInfo(output,
					FieldInfoOptionPath(mpath+"."+mk),
					FieldInfoOptionReflectionType(field.Type),
					FieldInfoOptionName(name),
				)
				continue
			}
			out, err := getFieldPath(mk, mv, field.Type.Elem(), mpath, depth+1)
			if err != nil {
				return err
			}
			*output = append(*output, out...)
		}
		return nil
	}

	out, err := getFieldPath(k, v, field.Type.Elem(), mpath, depth+1)
	if err != nil {
		return err
	}
	*output = append(*output, out...)
	return nil
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
