package unset

import (
	"maps"
	"reflect"
	"strconv"
	"strings"
	"time"

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

// normalizePath removes "inline." prefix since inline fields are flattened in BSON.
func normalizePath(path string) string {
	result, ok := strings.CutPrefix(path, "inline.")
	if ok {
		return result
	}
	return path
}

func isZeroValue(v reflect.Value) bool {
	return !v.IsValid() || v.IsZero()
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

// buildUpdate walks a struct recursively and builds MongoDB update operations.
//
// Rules:
// - Structs are traversed recursively using dot notation.
// - Inline fields are flattened into parent scope.
// - Fields in unsetSet are added to $unset and excluded from $set.
// - Zero values are ignored (PATCH semantics).
// - Nested empty structs are skipped.
func buildUpdate(v reflect.Value, prefix string, set bson.M, unset bson.M, unsetSet map[string]struct{}) {
	v = indirectValue(v)
	if !v.IsValid() {
		return
	}

	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		fv := v.Field(i)

		bsonTag := field.Tag.Get("bson")
		name, opts := parseBSONTag(bsonTag)
		if name == "" {
			name = field.Name
		}

		if opts["inline"] {
			buildUpdate(fv, prefix, set, unset, unsetSet)
			continue
		}

		path := name
		if prefix != "" {
			path = prefix + "." + name
		}

		if path == "_id" {
			continue
		}

		// Unset takes precedence
		if _, ok := unsetSet[path]; ok {
			unset[path] = ""
			continue
		}

		fvIndirect := reflect.Indirect(fv)

		if fvIndirect.Kind() == reflect.Struct && fvIndirect.Type() != reflect.TypeFor[time.Time]() {
			childSet := bson.M{}
			buildUpdate(fvIndirect, path, childSet, unset, unsetSet)

			for k := range childSet {
				if _, ok := unsetSet[k]; ok {
					delete(childSet, k)
				}
			}

			if len(childSet) > 0 {
				maps.Copy(set, childSet)
			}

			continue
		}

		// Skip zero values
		if isZeroValue(fv) {
			continue
		}

		set[path] = fv.Interface()
	}
}

func indirectValue(v reflect.Value) reflect.Value {
	for v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return reflect.Value{}
		}
		v = v.Elem()
	}
	return v
}
func buildUpdateMeta(v reflect.Value, prefix string, set bson.M, unset bson.M, unsetSet map[string]struct{}) {
	v = indirectValue(v)
	if !v.IsValid() {
		return
	}

	t := v.Type()
	meta := getTypeMeta(t)

	for _, f := range meta.fields {
		fv := v.Field(f.index)

		// Handle inline fields
		if f.inline {
			buildUpdateMeta(fv, prefix, set, unset, unsetSet)
			continue
		}

		path := f.name
		if prefix != "" {
			path = prefix + "." + f.name
		}

		if path == "_id" {
			continue
		}

		// Unset takes precedence
		if _, ok := unsetSet[path]; ok {
			unset[path] = ""
			continue
		}

		fv = indirectValue(fv)

		// Nested struct (but not time.Time)
		if fv.Kind() == reflect.Struct && fv.Type() != reflect.TypeFor[time.Time]() {
			buildUpdateMeta(fv, path, set, unset, unsetSet)
			continue
		}

		// Skip zero values
		if isZeroValue(fv) {
			continue
		}

		set[path] = fv.Interface()
	}
}
