package parsers

import (
	"fmt"
	"log/slog"
	"reflect"
)

type StructMapCache struct {
	StructMap map[string]StructMap
}

type StructMap struct {
	fields []FieldInfo
}

type FieldInfo struct {
	Path           string
	ReflectionType reflect.Type
	Name           string
}

var FieldMaps = StructMapCache{make(map[string]StructMap, 0)}

func (s *StructMapCache) Get(name string) *StructMap {
	smap := s.StructMap[name]
	if smap.Length() == 0 {
		return nil
	}
	return &smap
}

func (s *StructMapCache) Add(t any) error {
	fields, err := walkStruct(t)
	if err != nil {
		return err
	}

	name := getStructName(t)
	slog.Debug("adding entry to struct map", "name", name, "fields", len(fields))
	s.StructMap[name] = StructMap{fields: fields}
	return nil
}

func (s *StructMap) Validate(path string) *FieldInfo {

	for _, field := range s.fields {
		if field.Path == path {
			return &field
		}
	}

	return nil
}

func (s *StructMap) Length() int {
	return len(s.fields)
}

// Walk returns a slice describing every reachable exported field in v.
// v may be a struct, a pointer to a struct, or an interface holding one.
// Embedded structs, pointers, slices, arrays and maps are all traversed.
// Unexported fields are silently ignored.
func walkStruct(v any) ([]FieldInfo, error) {
	start := reflect.ValueOf(v)
	if !start.IsValid() {
		return nil, fmt.Errorf("nil value")
	}
	infos := []FieldInfo{}
	seen := map[uintptr]bool{} // for cycle detection
	if err := walk("", start, &infos, seen); err != nil {
		return nil, err
	}
	return infos, nil
}

func VerifyPath(pathMap map[string]FieldInfo, ptr string) (FieldInfo, error) {
	if fi, ok := pathMap[ptr]; ok {
		return fi, nil
	}
	return FieldInfo{}, fmt.Errorf("path %q does not exist on the target model", ptr)
}

// walk is the internal depth‑first walker.
// TODO: add value type check
func walk(path string, cur reflect.Value, out *[]FieldInfo, seen map[uintptr]bool) error {

	//1️⃣ Resolve indirections (pointers, interfaces) first.
	// Keep a set of visited addresses to avoid cycles.
	for {
		switch cur.Kind() {
		case reflect.Pointer, reflect.Interface:
			if cur.IsNil() {
				return nil // nothing more to explore here
			}
			if cur.Kind() == reflect.Pointer {
				ptr := cur.Pointer()
				if seen[ptr] {
					return nil // already visited – stop recursion
				}
				seen[ptr] = true
			}
			cur = cur.Elem()
			continue
		}
		break
	}

	// 2️⃣ Dispatch on the concrete kind
	switch cur.Kind() {

	case reflect.Struct:
		typ := cur.Type()
		for i := 0; i < typ.NumField(); i++ {
			sf := typ.Field(i)

			// Skip unexported fields – you cannot Interface() them safely.
			if sf.PkgPath != "" { // non‑empty => private
				continue
			}

			fieldVal := cur.Field(i)

			// Build the dotted path for this child.
			var childPath string
			if path == "" {
				childPath = sf.Name
			} else {
				childPath = fmt.Sprintf("%s.%s", path, sf.Name)
			}

			// Record the leaf **if it is not a further struct/slice/map**.
			// Otherwise we descend and let the recursion record the leaves.
			switch fieldVal.Kind() {
			case reflect.Struct, reflect.Pointer, reflect.Interface,
				reflect.Slice, reflect.Array, reflect.Map:
				// Not a leaf – keep walking.
				if err := walk(childPath, fieldVal, out, seen); err != nil {
					return err
				}
			default:
				// Primitive / leaf node – store its description.
				*out = append(*out, FieldInfo{
					Path:           childPath,
					ReflectionType: fieldVal.Type(),
					Name:           sf.Name,
				})
			}
		}
		return nil

	case reflect.Slice, reflect.Array:
		// Treat the slice itself as a field (e.g. “Tags”).
		// Then walk each element if the element type can contain structs.
		// We do **not** add an index to the path – the slice is considered a container.
		*out = append(*out, FieldInfo{
			Path:           path,
			ReflectionType: cur.Type(),
			Name:           pathName(path),
		})

		// If the element type is a struct/pointer/etc., walk the first element
		// just to discover the nested shape (no need to iterate all elements).
		if cur.Len() > 0 {
			elem := cur.Index(0)
			if elem.Kind() == reflect.Pointer && elem.IsNil() {
				// can't walk a nil pointer – stop here.
				return nil
			}
			// Recurse into the element without adding an index to the path.
			return walk(path, elem, out, seen)
		}
		return nil

	case reflect.Map:
		// Similar treatment to slices: record the map itself, then
		// walk one representative value (if any) to capture nested fields.
		*out = append(*out, FieldInfo{
			Path:           path,
			ReflectionType: cur.Type(),
			Name:           pathName(path),
		})

		if cur.Len() > 0 {
			// Grab an arbitrary value to discover its structure.
			val := cur.MapIndex(cur.MapKeys()[0])
			return walk(path, val, out, seen)
		}
		return nil

	default:
		// Primitive leaf that is the *root* value (e.g. Walk(5)).
		// Record it only if we have a non‑empty path (otherwise it's not a field).
		if path != "" {
			*out = append(*out, FieldInfo{
				Path:           path,
				ReflectionType: cur.Type(),
				Name:           pathName(path),
			})
		}
		return nil
	}
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
