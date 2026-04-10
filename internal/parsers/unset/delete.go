package unset

import (
	"strconv"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
)

func deleteNested(m bson.M, path string) {
	parts := strings.Split(path, ".")
	deleteRecursive(m, parts)
}

func deleteRecursive(current any, parts []string) {
	if len(parts) == 0 {
		return
	}

	switch obj := current.(type) {

	case bson.M:
		if len(parts) == 1 {
			delete(obj, parts[0])
			return
		}

		next, ok := obj[parts[0]]
		if !ok {
			return
		}

		deleteRecursive(next, parts[1:])

	case []any:
		idx, err := strconv.Atoi(parts[0])
		if err != nil || idx < 0 || idx >= len(obj) {
			return
		}

		if len(parts) == 1 {
			obj[idx] = nil
			return
		}

		deleteRecursive(obj[idx], parts[1:])
	}
}

func removeEmptyMaps(m bson.M) {
	for k, v := range m {
		switch val := v.(type) {

		case bson.M:
			removeEmptyMaps(val)
			if len(val) == 0 {
				delete(m, k)
			}

		case map[string]any:
			sub := bson.M(val)
			removeEmptyMaps(sub)
			if len(sub) == 0 {
				delete(m, k)
			} else {
				m[k] = sub
			}

		case []any:
			for i := range val {
				if sub, ok := val[i].(bson.M); ok {
					removeEmptyMaps(sub)
					if len(sub) == 0 {
						val[i] = nil
					}
				}
			}
		}
	}
}
