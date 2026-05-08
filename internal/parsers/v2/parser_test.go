package v2_test

import (
	"partialupdate/internal/models"
	v1 "partialupdate/internal/parsers/v1"
	v2 "partialupdate/internal/parsers/v2"
	"testing"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
)

type testCaseParser struct {
	name        string // descriptive name for the sub‑test
	patchJSON   string // JSON representation of the patch document
	Expected    v1.MongoUpdate
	wantErr     bool   // true → an error is expected
	errContains string // optional substring that must appear in the error message
}

func TestParsePatch(t *testing.T) {
	cases := []testCaseParser{
		{
			name: "simple patch with set and unset",
			patchJSON: `{
				"id": "123",
			"carResource": {
				"id": "car123",
				"make": "Toyota",
				"model": "Corolla"
			},
			"personResource": {
				"name": "Alice",
				"age": null
			}
			
			}`,
			Expected: v1.MongoUpdate{
				Set: bson.M{
					"id":                  "123",
					"carResource.id":      "car123",
					"carResource.make":    "Toyota",
					"carResource.model":   "Corolla",
					"personResource.name": "Alice",
				},
				Unset: bson.M{
					"personResource.age": "",
				},
				Push: bson.M{},
				Pull: bson.M{},
			},
			wantErr: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			update, err := v2.ParsePatch[models.Resource]([]byte(tc.patchJSON))

			if tc.wantErr {
				require.Error(t, err, "ParsePatch() expected an error but got nil")
				if tc.errContains != "" {
					require.Contains(t, err.Error(), tc.errContains, "ParsePatch() error = %v, want substring %q", err, tc.errContains)
				}
				return
			}

			require.NoError(t, err, "ParsePatch() unexpected error: %v", err)

			require.Equal(t, &tc.Expected, update, "ParsePatch() = %+v, want %+v", update, &tc.Expected)
		})
	}
}
