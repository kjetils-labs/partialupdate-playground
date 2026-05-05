package v1

import (
	"fmt"
	"strconv"
	"strings"
)

// ToMongoPath converts a RFC6902 JSON pointer path to MongoDB dot-notation path.
// Special handling for "-" (append to array) is needed — not direct MongoDB equivalent.
func toMongoPath(pointer string) (string, error) {
	if pointer == "" {
		return "", nil // root
	}
	if !strings.HasPrefix(pointer, "/") {
		return "", fmt.Errorf("invalid JSON pointer: %s", pointer)
	}

	segments := strings.Split(pointer[1:], "/")
	for i, seg := range segments {
		// Handle ~0 → ~, ~1 → /
		seg = strings.ReplaceAll(seg, "~1", "/")
		seg = strings.ReplaceAll(seg, "~0", "~")

		// '-' in MongoDB is not supported for add — handle separately
		if seg == "-" {
			if i != len(segments)-1 {
				return "", fmt.Errorf("'-' can only appear as final segment")
			}
			// We'll signal "append" with special marker in path or separately
			segments[i] = "-"
		}
		segments[i] = seg
	}

	path := strings.Join(segments[:len(segments)-1], ".")
	if len(segments) > 1 {
		return path, nil
	}
	return "", nil
}

// ExtractArrayIndex returns the index for append or numeric index; returns ok=false for "-".
func extractArrayIndex(lastSeg string) (index int, isArray bool, isAppend bool, err error) {
	if lastSeg == "-" {
		return 0, true, true, nil // append
	}
	if idx, err := strconv.Atoi(lastSeg); err == nil {
		return idx, true, false, nil
	}
	// Not an array index → object key
	return 0, false, false, nil
}

// unescapeRFC6902 decodes ~0 → ~, ~1 → /
func unescapeRFC6902(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, "~1", "/"), "~0", "~")
}
