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
		{
			name: "patch with null top-level field",
			patchJSON: `{
				"id": "123",
				"carResource": null,
				"personResource": {
					"name": "Alice",
					"ageInt": 30
				}
			}`,
			Expected: v1.MongoUpdate{
				Set: bson.M{
					"id":                    "123",
					"personResource.name":   "Alice",
					"personResource.ageInt": 30,
				},
				Unset: bson.M{
					"carResource": "",
				},
				Push: bson.M{},
				Pull: bson.M{},
			},
			wantErr: false,
		},
		{
			name: "invalid value type (string instead of int)",
			patchJSON: `{
				"id": "123",
				"carResource": {
					"id": "car123",
					"make": "Toyota",
					"model": "Corolla"
				},
				"personResource": {
					"name": "Alice",
					"ageInt": "thirty"
				}
			}`,
			wantErr:     true,
			errContains: "failed to convert value for field 'personResource.ageInt'",
		},
		{
			name: "invalid json syntax",
			patchJSON: `{
				"id": "123",
				"carResource": {
					"id": "car123",
					"make": "Toyota",
					"model": "Corolla"
				},
				"personResource": {
					"name": "Alice",
					"ageInt": 30,
				}
			`, // missing closing brace and extra comma
			wantErr:     true,
			errContains: "failed to unmarshal patch JSON",
		},
		{
			name:      "empty patch",
			patchJSON: `{}`,
			Expected: v1.MongoUpdate{
				Set:   bson.M{},
				Unset: bson.M{},
				Push:  bson.M{},
				Pull:  bson.M{},
			},
			wantErr: false,
		},
		{
			name: "null value on non-pointer field",
			patchJSON: `{
				"id": "123",
				"carResource": {
					"id": "car123",
					"make": "Toyota",
					"model": "Corolla"
				},
				"personResource": {
					"name": null,
					"ageInt": 30
				}
			}`,
			Expected: v1.MongoUpdate{
				Set: bson.M{
					"id":                    "123",
					"carResource.id":        "car123",
					"carResource.make":      "Toyota",
					"carResource.model":     "Corolla",
					"personResource.ageInt": 30,
				},
				Unset: bson.M{
					"personResource.name": "",
				},
				Push: bson.M{},
				Pull: bson.M{},
			},
			wantErr: false,
		},
		{
			name: "mismatched fields in JSON (extra fields)",
			patchJSON: `{
				"id": "123",
				"carResource": {
					"id": "car123",
					"make": "Toyota",
					"model": "Corolla",
					"color": "red"
				},
				"personResource": {
					"name": "Alice",
					"ageInt": 30,
					"hobby": "painting"
				}
			}`,
			wantErr:     true,                                // extra fields should be ignored, not cause an error
			errContains: "failed to walk struct for fields:", // if you want to check for specific error message
		},
		{
			name: "struct in value field (invalid)",
			patchJSON: `{
				"id": "123",
				"carResource": {
					"id": "car123",
					"make": "Toyota",
					"model": "Corolla"
				},
				"personResource": {
					"name": {"first": "Alice", "last": "Smith"},
					"ageInt": 30
				}
			}`,
			wantErr:     true,
			errContains: "failed to convert value for field 'personResource.name'",
		},
		{
			name: "test types conversion",
			patchJSON: `{
				"personResource": {
					"name": "Alice",
					"pointerString": "pointer value",
					"pointerInt": 42,
					"pointerBool": true,
					"ageInt": 30,
					"ageInt8": 30,
					"ageInt16": 30,
					"ageInt32": 30,
					"ageInt64": 30
				}
			}`,
			wantErr: false,
			Expected: v1.MongoUpdate{
				Set: bson.M{
					"personResource.name":          "Alice",
					"personResource.ageInt":        30,
					"personResource.ageInt8":       int8(30),
					"personResource.ageInt16":      int16(30),
					"personResource.ageInt32":      int32(30),
					"personResource.ageInt64":      int64(30),
					"personResource.pointerString": "pointer value",
					"personResource.pointerInt":    42,
					"personResource.pointerBool":   true,
				},
				Unset: bson.M{},
				Push:  bson.M{},
				Pull:  bson.M{},
			},
		},
		{
			name: "inline struct fields",
			patchJSON: `{
				"personResource": {
					"name": "Alice",
					"likesCheesecake": true
				}
			}`,
			wantErr: false,
			Expected: v1.MongoUpdate{
				Set: bson.M{
					"personResource.name":            "Alice",
					"personResource.likesCheesecake": true,
				},
				Unset: bson.M{},
				Push:  bson.M{},
				Pull:  bson.M{},
			},
		},
		{
			name: "slice of things",
			patchJSON: `{
			"personResource": {
				"name": "Alice",
				"slice": ["tag1", "tag2", "tag3"],
			    "pointerSlice": ["ptrTag1", "ptrTag2"],
				"intSlice": [1, 2, 3]
				}
			}`,
			wantErr: false,
			Expected: v1.MongoUpdate{
				Set: bson.M{
					"personResource.name":         "Alice",
					"personResource.slice":        []any{"tag1", "tag2", "tag3"},
					"personResource.pointerSlice": []any{"ptrTag1", "ptrTag2"},
					"personResource.intSlice":     []any{1, 2, 3},
				},
				Unset: bson.M{},
				Push:  bson.M{},
				Pull:  bson.M{},
			},
		},
		{
			name: "map of things",
			patchJSON: `{
				"personResource": {
					"name": "Alice",
					"tags": {
						"tag1": "value1",
						"tag2": "value2"
					},
					"pointerTags": {
						"ptrTag1": "ptrValue1",
						"ptrTag2": "ptrValue2"
					}
				}
			}`,
			wantErr: false,
			Expected: v1.MongoUpdate{
				Set: bson.M{
					"personResource.name":                "Alice",
					"personResource.tags.tag1":           "value1",
					"personResource.tags.tag2":           "value2",
					"personResource.pointerTags.ptrTag1": "ptrValue1",
					"personResource.pointerTags.ptrTag2": "ptrValue2",
				},
				Unset: bson.M{},
				Push:  bson.M{},
				Pull:  bson.M{},
			},
		},
		// {
		// 	name: "slice of structs",
		// 	patchJSON: `{
		// 		"personResource": {
		// 			"structSlice":
		// 			[
		// 				{
		// 					"model": "Corolla",
		// 					"horses": {
		// 						"horses": 0
		// 					}
		// 				},
		// 				{
		// 					"model": "Civic",
		// 					"horses": {
		// 						"horses": 0
		// 					}
		// 				}
		// 			]
		// 		}
		// 	}`,
		// 	wantErr: false,
		// 	Expected: v1.MongoUpdate{
		// 		Set: bson.M{
		// 			"personResource.sliceOfStructs": []any{
		// 				bson.M{
		// 					"model": "Corolla",
		// 					"horses": bson.M{
		// 						"horses": 0,
		// 					},
		// 				},
		// 				bson.M{
		// 					"model": "Civic",
		// 					"horses": bson.M{
		// 						"horses": 0,
		// 					},
		// 				},
		// 			},
		// 		},
		// 		Unset: bson.M{},
		// 		Push:  bson.M{},
		// 		Pull:  bson.M{},
		// 	},
		// },
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
