package v2

import (
	"encoding/json"
	"fmt"
	"log/slog"
	v1 "partialupdate/internal/parsers/v1"
)

// ParsePatch takes a struct and converts it to a map of JSON raw messages, which can then be used to construct a MongoDB update document. The generic type T allows for flexibility in the input struct type.
func ParsePatch[T any](patch []byte) (*v1.MongoUpdate, error) {

	var input map[string]any
	err := json.Unmarshal(patch, &input)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal patch JSON: %w", err)
	}

	fields, err := WalkStruct(input, new(T))
	if err != nil {
		slog.Error("failed to walk struct for fields", "error", err)
	}

	for _, field := range fields {
		slog.Info("field info", "path", field.Path, "name", field.Name, "type", field.ReflectionType.String(), "value", field.Value)
	}

	output := v1.NewMongoUpdate()
	for _, field := range fields {

		if field.Value == nil {
			output.Unset[field.Path] = ""
			continue
		}

		output.Set[field.Path] = field.Value
	}

	return output, nil
}
