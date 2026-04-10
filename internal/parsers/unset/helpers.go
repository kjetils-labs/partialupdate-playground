package unset

import (
	"reflect"
	"strconv"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
)

func indirectType(t reflect.Type) reflect.Type {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return t
}

func parseBSONTag(tag string) (string, map[string]bool) {
	parts := strings.Split(tag, ",")
	name := parts[0]

	opts := make(map[string]bool)
	for _, p := range parts[1:] {
		opts[p] = true
	}

	return name, opts
}

func isIndex(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

func normalizePath(path string) string {
	// TEMP pragmatic fix:
	result, ok := strings.CutPrefix(path, "inline.")
	if ok {
		return result
	}
	return path
}

func containsUnsetMaskNormalized(unset bson.M, field string) bool {
	_, exists := unset[field]
	return exists
}

func isZeroValue(v any) bool {
	return v == nil || reflect.DeepEqual(v, reflect.Zero(reflect.TypeOf(v)).Interface())
}

func checkInlineFields(meta *typeMeta, remainingParts []string) bool {
	for _, f := range meta.fields {
		ft := indirectType(f.typ)
		if ft.Kind() == reflect.Struct {
			// Recursively attempt to validate the remaining path in this struct
			if err := validatePath(ft, strings.Join(remainingParts, ".")); err == nil {
				return true
			}
		}
	}
	return false
}

func flattenMap(prefix string, input map[string]any, out map[string]any) {
	for k, v := range input {
		key := k
		if prefix != "" {
			key = prefix + "." + k
		}

		switch val := v.(type) {
		case map[string]any:
			flattenMap(key, val, out)
		case bson.M:
			flattenMap(key, val, out)
		default:
			out[key] = val
		}
	}
}
