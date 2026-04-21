package parsers

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
)

func applyAdd(update *MongoUpdate, op Operation, path string) error {
	// _, _, isAppend, err := parseIndex(last)
	// switch {
	// case errors.Is(err, errNotAppendAction):
	// default:
	// 	return fmt.Errorf("invalid array index '%s': %w", last, err)
	// }

	// Root-level path (e.g., "/" → root)
	// if len(segments) == 1 && segments[0] == "" {
	// 	return applyRootAdd(update, op)
	// }

	// Build MongoDB dot-path: join parent segments
	// parentPath := strings.Join(segments[:len(segments)-1], ".")
	// _ = parentPath + "." + last

	// Special: append to array? → use $push (only if target is an array field)
	// if isAppend {
	// 	if parentPath == "" {
	// 		return fmt.Errorf("cannot append to root: '-' only allowed on array fields")
	// 	}
	// 	return applyPush(update, parentPath, op.Value)
	// }

	// Normal add: use $set (create/replace)
	var val any
	if op.Value != nil {
		if err := json.Unmarshal(op.Value, &val); err != nil {
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

func applyReplace(update *MongoUpdate, op Operation, path string) error {
	return applyAdd(update, op, path)
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
