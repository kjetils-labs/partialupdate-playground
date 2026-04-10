package unset

import (
	"errors"
	"fmt"
	"maps"
	"reflect"
	"strings"
	"sync"
)

type fieldMeta struct {
	typ reflect.Type
}

type typeMeta struct {
	fields map[string]fieldMeta
}

var metaCache sync.Map // map[reflect.Type]*typeMeta

func getTypeMeta(t reflect.Type) *typeMeta {
	if v, ok := metaCache.Load(t); ok {
		return v.(*typeMeta)
	}

	meta := buildTypeMeta(t)
	metaCache.Store(t, meta)
	return meta
}

func buildTypeMeta(t reflect.Type) *typeMeta {
	fields := make(map[string]fieldMeta)

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)

		if f.PkgPath != "" {
			continue
		}

		tag := f.Tag.Get("bson")
		name, opts := parseBSONTag(tag)

		if opts["inline"] {
			ft := indirectType(f.Type)
			if ft.Kind() == reflect.Struct {
				inlineMeta := getTypeMeta(ft)
				maps.Copy(fields, inlineMeta.fields)
			}
			continue
		}

		if name == "" {
			name = strings.ToLower(f.Name)
		}

		fields[name] = fieldMeta{
			typ: f.Type,
		}
	}

	return &typeMeta{fields: fields}
}

func validateUnsetMask(resource any, paths []string) error {
	if len(paths) == 0 || resource == nil {
		return nil
	}

	t := indirectType(reflect.TypeOf(resource))

	if t.Kind() != reflect.Struct {
		return errors.New("unsetMask validation requires struct resource")
	}

	for _, path := range paths {
		if path == "" {
			return fmt.Errorf("invalid empty path")
		}

		if err := validatePath(t, path); err != nil {
			return fmt.Errorf("invalid path '%s': %w", path, err)
		}
	}

	return nil
}

func validatePath(t reflect.Type, path string) error {
	parts := strings.Split(path, ".")
	current := t

	for i := 0; i < len(parts); i++ {
		part := parts[i]
		meta := getTypeMeta(current)

		field, ok := meta.fields[part]
		if !ok {
			if checkInlineFields(meta, parts[i:]) {
				return nil
			}

			return fmt.Errorf("field '%s' not found in %s", part, current.Name())
		}

		ft := indirectType(field.typ)

		if i+1 < len(parts) && isIndex(parts[i+1]) {
			if ft.Kind() != reflect.Slice && ft.Kind() != reflect.Array {
				return fmt.Errorf("field '%s' is not an array", part)
			}

			ft = indirectType(ft.Elem())
			i++
		}

		if i < len(parts)-1 {
			if ft.Kind() != reflect.Struct {
				return fmt.Errorf("field '%s' is not a struct", part)
			}
			current = ft
		}
	}

	return nil
}

func validateUnsetMaskCached(resource any, paths []string) error {
	if len(paths) == 0 || resource == nil {
		return nil
	}

	t := indirectType(reflect.TypeOf(resource))

	if t.Kind() != reflect.Struct {
		return errors.New("unsetMask validation requires struct resource")
	}

	for _, path := range paths {
		if path == "" {
			return fmt.Errorf("invalid empty path")
		}

		if err := validatePathCached(t, path); err != nil {
			return fmt.Errorf("invalid path '%s': %w", path, err)
		}
	}

	return nil
}

func validatePathCached(t reflect.Type, path string) error {
	parts := strings.Split(path, ".")
	current := t

	for i := 0; i < len(parts); i++ {
		part := parts[i]

		meta := getTypeMeta(current)

		field, ok := meta.fields[part]
		if !ok {
			return fmt.Errorf("field '%s' not found in %s", part, current.Name())
		}

		ft := indirectType(field.typ)

		// If next part is array index
		if i+1 < len(parts) && isIndex(parts[i+1]) {
			if ft.Kind() != reflect.Slice && ft.Kind() != reflect.Array {
				return fmt.Errorf("field '%s' is not an array", part)
			}

			ft = indirectType(ft.Elem())
			i++ // skip index
		}

		// Traverse deeper if not last
		if i < len(parts)-1 {
			if ft.Kind() != reflect.Struct {
				return fmt.Errorf("field '%s' is not a struct", part)
			}
			current = ft
		}
	}

	return nil
}
