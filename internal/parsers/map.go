package parsers

import (
	"fmt"
	"reflect"
	"strings"
	"unicode"
	"unicode/utf8"
)

type FieldInfo struct {
	Path           string
	ReflectionType reflect.Type
	Name           string
}

func VerifyPath(pathMap map[string]FieldInfo, ptr string) (FieldInfo, error) {
	if fi, ok := pathMap[ptr]; ok {
		return fi, nil
	}
	return FieldInfo{}, fmt.Errorf("path %q does not exist on the target model", ptr)
}

// pathName extracts the last component of a dot‑path.
// Used when we need a “field name” for synthetic nodes like a slice.
func pathName(p string) string {
	if p == "" {
		return ""
	}
	// Find the last '.' and return everything after it.
	if i := lastIndex(p, '.'); i >= 0 {
		return p[i+1:]
	}
	return p
}

// lastIndex is a helper that avoids importing strings just for one call.
func lastIndex(s string, sep byte) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == sep {
			return i
		}
	}
	return -1
}

func getStructName(t any) string {
	var name string
	typ := reflect.TypeOf(t)
	if typ.Kind() == reflect.Pointer {
		name = typ.Elem().Name()
	}

	name = reflect.TypeOf(t).Name()
	return name
}

func firstToLower(s string) string {
	r, size := utf8.DecodeRuneInString(s)
	if r == utf8.RuneError && size <= 1 {
		return s
	}
	lc := unicode.ToLower(r)
	if r == lc {
		return s
	}
	return string(lc) + s[size:]
}

func WalkStruct(str any) []FieldInfo {
	initialPath := "/"
	t := reflect.TypeOf(str)
	// v := reflect.ValueOf(str)
	return walkStruct(initialPath, t)
}

func walkStruct(path string, str reflect.Type) []FieldInfo {
	paths := make([]FieldInfo, 0)
	for field := range str.Fields() {
		_paths := getFieldPath(path, field)
		paths = append(paths, _paths...)
	}

	return paths
}

// A field can return multiple paths if the field is a struct
func getFieldPath(path string, field reflect.StructField) []FieldInfo {
	paths := make([]FieldInfo, 0)
	kind := field.Type.Kind()

	if kind == reflect.Pointer {
		kind = reflect.TypeOf(field.Type.Elem()).Kind()
		// kind = field.Type.Elem().Kind()
	}
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

	// we burrow down until we find all base types
	if kind == reflect.Struct {

		// make sure to add this struct as a part of the path
		path += fmt.Sprintf("%s/", name)

		// walk the nested struct
		_fields := walkStruct(path, field.Type)
		paths = append(paths, _fields...)
		return paths
	}

	path += name

	finfo := FieldInfo{
		Path:           path,
		ReflectionType: field.Type,
		Name:           name,
	}
	paths = append(paths, finfo)

	return paths
}
