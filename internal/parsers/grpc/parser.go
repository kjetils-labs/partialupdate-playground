package grpc

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

type Request struct {
	Resource   any      `json:"resource" bson:"resource,omitempty"`
	UpdateMask []string `json:"updateMask" bson:"updateMask,omitempty"`
}

func NewRequest(resource any, updateMask []string) *Request {
	return &Request{
		Resource:   resource,
		UpdateMask: updateMask,
	}
}

func (r *Request) Parse() ([]byte, error) {
	if r.Resource == nil {
		return nil, errors.New("resource is nil")
	}

	val := reflect.ValueOf(r.Resource)
	if val.Kind() == reflect.Pointer {
		if val.IsNil() {
			return nil, fmt.Errorf("nil pointer")
		}
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected struct or pointer to struct, got %T", r.Resource)
	}

	// Build a tree that groups mask entries by their first component.
	// Example:
	//   ["name", "primaryContact.phone", "primaryContact.email"]
	// becomes
	//   {"name": nil,
	//    "primaryContact": {"phone": nil, "email": nil}}
	maskTree := buildMaskTree(r.UpdateMask)

	// Recursively walk the struct and construct a map[string]interface{}
	// that contains only the requested fields.
	mapped, err := structToMap(val, maskTree)
	if err != nil {
		return nil, err
	}
	return json.Marshal(mapped)
}

type maskNode map[string]maskNode // nil leaf means “stop here”

func buildMaskTree(mask []string) maskNode {
	root := maskNode{}
	for _, p := range mask {
		parts := strings.Split(p, ".")
		cur := root
		for i, part := range parts {
			if cur[part] == nil {
				// If we are at the last component, store a nil leaf.
				if i == len(parts)-1 {
					cur[part] = nil
				} else {
					cur[part] = maskNode{}
				}
			}
			// descend (if not leaf)
			if cur[part] != nil {
				cur = cur[part]
			}
		}
	}
	return root
}

func structToMap(v reflect.Value, tree maskNode) (map[string]any, error) {
	typ := v.Type()
	out := make(map[string]any)

	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)

		// Skip unexported fields – they cannot be accessed via reflection.
		if f.PkgPath != "" && !f.Anonymous {
			continue
		}

		// Resolve the JSON name for this field.
		jsonName := jsonFieldName(f)
		if jsonName == "" || jsonName == "-" {
			// No JSON representation – ignore.
			continue
		}

		node, ok := tree[jsonName]
		if !ok {
			// Not requested in the mask → skip.
			continue
		}

		fieldVal := v.Field(i)

		// If the field is a pointer, dereference it (nil stays nil).
		if fieldVal.Kind() == reflect.Pointer {
			if fieldVal.IsNil() {
				continue
			}
			fieldVal = fieldVal.Elem()
		}

		// Leaf node → copy the value as‑is (using json Marshal for complex types).
		if node == nil {
			// Use the standard library to turn the value into something that
			// json.Marshal understands (e.g. time.Time, slices, maps, etc.).
			b, err := json.Marshal(fieldVal.Interface())
			if err != nil {
				return nil, fmt.Errorf("cannot marshal field %s: %w", jsonName, err)
			}
			var generic any
			if err := json.Unmarshal(b, &generic); err != nil {
				return nil, fmt.Errorf("cannot unmarshal field %s: %w", jsonName, err)
			}
			out[jsonName] = generic
			continue
		}

		// Non‑leaf → must be a struct (or pointer to struct) that we can dive into.
		if fieldVal.Kind() != reflect.Struct {
			return nil, fmt.Errorf("field %s is not a struct but mask expects sub‑fields", jsonName)
		}
		subMap, err := structToMap(fieldVal, node)
		if err != nil {
			return nil, err
		}
		// Omit empty sub‑objects (they happen when all nested fields were nil).
		if len(subMap) > 0 {
			out[jsonName] = subMap
		}
	}
	return out, nil
}

func jsonFieldName(f reflect.StructField) string {
	tag := f.Tag.Get("json")
	if tag == "" {
		// default rule – lower‑camel‑case of the field name.
		return lowerCamel(f.Name)
	}
	// tag may be "name,omitempty" – strip everything after the first comma.
	if idx := strings.Index(tag, ","); idx != -1 {
		tag = tag[:idx]
	}
	return tag
}

// lowerCamel converts "PrimaryContact" → "primaryContact".
func lowerCamel(s string) string {
	if s == "" {
		return ""
	}
	runes := []rune(s)
	runes[0] = []rune(strings.ToLower(string(runes[0])))[0]
	return string(runes)
}
