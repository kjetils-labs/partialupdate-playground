package v1_test

import (
	"partialupdate/internal/models"
	v1 "partialupdate/internal/parsers/v1"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

type testCase struct {
	name     string
	path     string
	input    any
	expected v1.FieldInfo
	wantErr  bool
	errCheck func(error) bool
}

func TestWalkStruct(t *testing.T) {

	tests := []testCase{
		{
			name:  "walk root",
			path:  "/",
			input: models.Resource{},
			expected: v1.FieldInfo{
				Path:           "/",
				ReflectionType: reflect.TypeFor[models.Resource](),
				Name:           "Resource",
			},
		},
		{
			name:  "walk field",
			path:  "/id",
			input: models.Resource{},
			expected: v1.FieldInfo{
				Path:           "/id",
				ReflectionType: reflect.TypeFor[string](),
				Name:           "id",
			},
		},
		{
			name:  "walk nested field",
			path:  "/carResource/model",
			input: models.Resource{},
			expected: v1.FieldInfo{
				Path:           "/carResource/model",
				ReflectionType: reflect.TypeFor[string](),
				Name:           "model",
			},
		},
		{
			name:  "walk pointer field",
			path:  "/personResource",
			input: models.Resource{},
			expected: v1.FieldInfo{
				Path:           "/personResource",
				ReflectionType: reflect.TypeFor[models.Resource](),
				Name:           "personResource",
			},
		},
		{
			name:  "walk nested pointer field string",
			path:  "/personResource/name",
			input: models.Resource{},
			expected: v1.FieldInfo{
				Path:           "/personResource/name",
				ReflectionType: reflect.TypeFor[string](),
				Name:           "name",
			},
		},
		{
			name:  "walk nested pointer field int",
			path:  "/personResource/personData/personalNumber",
			input: models.Resource{},
			expected: v1.FieldInfo{
				Path:           "/personResource/personData/personalNumber",
				ReflectionType: reflect.TypeFor[int](),
				Name:           "personalNumber",
			},
		},
		{
			name:    "walk non-existent field",
			path:    "/nonExistent",
			input:   models.Resource{},
			wantErr: true,
			errCheck: func(err error) bool {
				return err.Error() == `field "nonExistent" does not exist in type "Resource"`
			},
		},
		{
			name:  "1 layer map field",
			path:  "/personResource/tags/fork",
			input: models.Resource{},
			expected: v1.FieldInfo{
				Path:           "/personResource/tags/fork",
				ReflectionType: reflect.TypeFor[string](),
				Name:           "tags",
			},
		},
		{
			name:  "struct map",
			path:  "/personResource/structTags",
			input: models.Resource{},
			expected: v1.FieldInfo{
				Path:           "/personResource/structTags",
				ReflectionType: reflect.TypeFor[map[string]models.Car](),
				Name:           "structTags",
			},
		},
		{
			name:  "struct map nested field",
			path:  "/personResource/structTags/somevalue/horses/horses",
			input: models.Resource{},
			expected: v1.FieldInfo{
				Path:           "/personResource/structTags/somevalue/horses/horses",
				ReflectionType: reflect.TypeFor[int](),
				Name:           "horses",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := v1.WalkStruct(tc.path, tc.input)
			if err != nil {

				if tc.wantErr {
					if !tc.errCheck(err) {
						t.Fatalf("error = '%v' does not satisfy error check", err)
					}
					return
				}
				t.Fatalf("unexpected error: %v", err)

			}
			require.Equal(t, tc.expected.Path, got.Path, "path does not match")
			require.Equal(t, tc.expected.ReflectionType.Kind().String(), got.ReflectionType.Kind().String(), "reflection type kind does not match")
			require.Equal(t, tc.expected.Name, got.Name, "name does not match")
		})
	}
}
