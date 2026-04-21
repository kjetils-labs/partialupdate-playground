package parsers_test

import (
	"partialupdate/internal/models"
	"partialupdate/internal/parsers"
	"testing"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
)

type testCasesParse struct {
	name           string // Clear, descriptive name for `t.Run`
	operationJSON  string
	expectedUpdate *parsers.MongoUpdate // *MongoUpdate to allow nil for errors
	expectError    bool                 // whether an error is expected
	// Optional: current document state (for testing move/copy/test, though unsupported now)
	currentDoc *bson.M
}

func TestParse(t *testing.T) {

	tests := []testCasesParse{
		{
			name:          "add to missing field (creates path)",
			operationJSON: `[{"op":"add","path":"/personResource/name","value":"Bob"}]`,
			expectedUpdate: &parsers.MongoUpdate{
				Set:         bson.M{"personResource.name": "Bob"},
				Unset:       bson.M{},
				Push:        bson.M{},
				Pull:        bson.M{},
				SetOnInsert: bson.M{},
				Inc:         bson.M{},
			},
			expectError: false,
		},
		{
			name:          "replace existing field",
			operationJSON: `[{"op":"replace","path":"/personResource/name","value":"Charlie"}]`,
			expectedUpdate: &parsers.MongoUpdate{
				Set:         bson.M{"personResource.name": "Charlie"},
				Unset:       bson.M{},
				Push:        bson.M{},
				Pull:        bson.M{},
				SetOnInsert: bson.M{},
				Inc:         bson.M{},
			},
			expectError: false,
		},
		{
			name:          "remove field",
			operationJSON: `[{"op":"remove","path":"/personResource/name"}]`,
			expectedUpdate: &parsers.MongoUpdate{
				Unset:       bson.M{"personResource.name": ""},
				Set:         bson.M{},
				Push:        bson.M{},
				Pull:        bson.M{},
				SetOnInsert: bson.M{},
				Inc:         bson.M{},
			},
			expectError: false,
		},
		{
			name:          "add nested object (PersonData)",
			operationJSON: `[{"op":"add","path":"/personResource/personData","value":{"personalNumber":987654,"religious":true}}]`,
			expectedUpdate: &parsers.MongoUpdate{
				Set: bson.M{"personResource.personData": bson.M{
					"personalNumber": 987654,
					"religious":      true,
				}},
				Unset:       bson.M{},
				Push:        bson.M{},
				Pull:        bson.M{},
				SetOnInsert: bson.M{},
				Inc:         bson.M{},
			},
			expectError: false,
		},

		// --- Inlined field (CheeseCakeLiker) ---
		{
			name:          "add inlined field",
			operationJSON: `[{"op":"add","path":"/personResource/likesCheesecake","value":false}]`,
			expectedUpdate: &parsers.MongoUpdate{
				Set:         bson.M{"personResource.likesCheesecake": false},
				Unset:       bson.M{},
				Push:        bson.M{},
				Pull:        bson.M{},
				SetOnInsert: bson.M{},
				Inc:         bson.M{},
			},
			expectError: false,
		},
		{
			name:          "add inlined field",
			operationJSON: `[{"op":"add","path":"/personResource/likesCheesecake","value":"false"}]`,
			expectedUpdate: &parsers.MongoUpdate{
				Set:         bson.M{"personResource.likesCheesecake": false},
				Unset:       bson.M{},
				Push:        bson.M{},
				Pull:        bson.M{},
				SetOnInsert: bson.M{},
				Inc:         bson.M{},
			},
			expectError: true,
		},
		{
			name:          "remove inlined field",
			operationJSON: `[{"op":"remove","path":"/personResource/likesCheesecake"}]`,
			expectedUpdate: &parsers.MongoUpdate{
				Unset:       bson.M{"personResource.likesCheesecake": ""},
				Set:         bson.M{},
				Push:        bson.M{},
				Pull:        bson.M{},
				SetOnInsert: bson.M{},
				Inc:         bson.M{},
			},
			expectError: false,
		},

		// --- Numeric & boolean values ---
		{
			name:          "update age (int)",
			operationJSON: `[{"op":"replace","path":"/personResource/age","value":31}]`,
			expectedUpdate: &parsers.MongoUpdate{
				Set:         bson.M{"personResource.age": 31},
				Unset:       bson.M{},
				Push:        bson.M{},
				Pull:        bson.M{},
				SetOnInsert: bson.M{},
				Inc:         bson.M{},
			},
			expectError: false,
		},
		{
			name:          "update age (string)",
			operationJSON: `[{"op":"replace","path":"/personResource/age","value":"31"}]`,
			expectedUpdate: &parsers.MongoUpdate{
				Set:         bson.M{"personResource.age": 31},
				Unset:       bson.M{},
				Push:        bson.M{},
				Pull:        bson.M{},
				SetOnInsert: bson.M{},
				Inc:         bson.M{},
			},
			expectError: true,
		},
		{
			name:          "set alive = false (bool)",
			operationJSON: `[{"op":"replace","path":"/personResource/alive","value":false}]`,
			expectedUpdate: &parsers.MongoUpdate{
				Set:         bson.M{"personResource.alive": false},
				Unset:       bson.M{},
				Push:        bson.M{},
				Pull:        bson.M{},
				SetOnInsert: bson.M{},
				Inc:         bson.M{},
			},
			expectError: false,
		},
		{
			name:          "set religious = null (explicit null)",
			operationJSON: `[{"op":"replace","path":"/personResource/personData/religious","value":null}]`,
			expectedUpdate: &parsers.MongoUpdate{
				Set:         bson.M{"personResource.personData.religious": nil},
				Unset:       bson.M{},
				Push:        bson.M{},
				Pull:        bson.M{},
				SetOnInsert: bson.M{},
				Inc:         bson.M{},
			},
			expectError: false,
		},

		// --- Path with escape sequences (~0 → ~, ~1 → /) ---
		{
			name:          "path with escaped slash (e.g., /foo~1bar)",
			operationJSON: `[{"op":"add","path":"/personResource/foo~1bar","value":"test"}]`,
			expectedUpdate: &parsers.MongoUpdate{
				Set:         bson.M{"personResource.foo/bar": "test"},
				Unset:       bson.M{},
				Push:        bson.M{},
				Pull:        bson.M{},
				SetOnInsert: bson.M{},
				Inc:         bson.M{},
			},
			expectError: false,
		},
	}

	// Run tests
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {

			actualUpdate, err := parsers.Parse[models.Resource]([]byte(tc.operationJSON))

			if tc.expectError {
				require.Error(t, err, "did not fail at parsing when expected")
				return
			}
			require.NoErrorf(t, err, "failed at parsing")

			// Compare (ignore nil vs empty maps)
			require.Equal(t, tc.expectedUpdate, actualUpdate,
				"Mongo update mismatch\nExpected: %+v\nGot:      %+v", tc.expectedUpdate, actualUpdate)
		})
	}
}

func sampleResource() *models.Resource {
	return &models.Resource{
		ID: "doc123",
		PersonResource: &models.Person{
			ID:    "p456",
			Name:  "Alice",
			Alive: true,
			Age:   30,
			PersonData: models.PersonData{
				PersonalNumber: 123456,
				Religious:      false,
			},
			CheeseCakeLiker: models.CheeseCakeLiker{
				LikesCheesecake: true,
			},
		},
	}
}

// BSON doc version
func sampleResourceBSON() bson.M {
	return bson.M{
		"_id": "doc123",
		"personResource": bson.M{
			"id":    "p456",
			"name":  "Alice",
			"alive": true,
			"age":   30,
			"personData": bson.M{
				"personalNumber": 123456,
				"religious":      false,
			},
			"likesCheesecake": true,
		},
	}
}
