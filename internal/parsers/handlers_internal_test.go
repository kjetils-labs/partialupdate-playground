package parsers

import (
	"encoding/json"
	"partialupdate/internal/models"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnmarshalValue(t *testing.T) {
	// This is just a placeholder to make the package "main" compile.
	// The actual logic is in the parsers package and its tests.

	fields := WalkStruct(models.Resource{})

	var val any
	var err error
	for _, field := range fields {
		if field.Path == "/personResource/ageInt" {
			data := []byte(`5`)
			rawMessage := json.RawMessage(data)

			val, err = unmarshalValue(rawMessage, field.ReflectionType)
		}
	}

	require.NoError(t, err, "unmarshal value should not error")
	require.Equal(t, 5, val, "unmarshaled value should be 5")
	require.IsType(t, 5, val, "unmarshaled value should be of type int")
}

func TestUnmarshalnilValue(t *testing.T) {
	fields := WalkStruct(models.Resource{})

	var val any
	var err error
	for _, field := range fields {
		if field.Path == "/personResource/ageInt" {
			data := make([]byte, 0) // nil slice
			rawMessage := json.RawMessage(data)

			val, err = unmarshalValue(rawMessage, field.ReflectionType)
		}
	}

	require.ErrorAs(t, err, &ErrNilValue, "unmarshal nil value should return ErrNilValue")
	require.Equal(t, nil, val, "unmarshaled nil value should be zero value (0 for int)")
	require.IsType(t, nil, val, "unmarshaled nil value should be of type int")
}
