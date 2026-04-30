package parsers_test

//
// type testCasesParse struct {
// 	name           string // Clear, descriptive name for `t.Run`
// 	operationJSON  string
// 	expectedUpdate *parsers.MongoUpdate // *MongoUpdate to allow nil for errors
// 	expectError    bool                 // whether an error is expected
// 	// Optional: current document state (for testing move/copy/test, though unsupported now)
// 	currentDoc *bson.M
// }
//
// func TestParse(t *testing.T) {
//
// 	tests := []testCasesParse{
// 		// ======================
// 		// 1. String Fields
// 		// ======================
// 		{
// 			name:          "replace name (string)",
// 			operationJSON: `[{"op":"replace","path":"/personResource/name","value":"Zara"}]`,
// 			expectedUpdate: &parsers.MongoUpdate{
// 				Set:         bson.M{"personResource.name": "Zara"},
// 				Unset:       bson.M{},
// 				Push:        bson.M{},
// 				Pull:        bson.M{},
// 				SetOnInsert: bson.M{},
// 				Inc:         bson.M{},
// 			},
// 			expectError: false,
// 		},
// 		{
// 			name:           "replace name with number (type mismatch)",
// 			operationJSON:  `[{"op":"replace","path":"/personResource/name","value":123}]`,
// 			expectedUpdate: nil,
// 			expectError:    true,
// 		},
//
// 		// ======================
// 		// 2. Boolean Fields
// 		// ======================
// 		{
// 			name:          "replace alive (bool)",
// 			operationJSON: `[{"op":"replace","path":"/personResource/alive","value":false}]`,
// 			expectedUpdate: &parsers.MongoUpdate{
// 				Set:         bson.M{"personResource.alive": false},
// 				Unset:       bson.M{},
// 				Push:        bson.M{},
// 				Pull:        bson.M{},
// 				SetOnInsert: bson.M{},
// 				Inc:         bson.M{},
// 			},
// 			expectError: false,
// 		},
// 		{
// 			name:           "replace alive with string 'true' (should fail)",
// 			operationJSON:  `[{"op":"replace","path":"/personResource/alive","value":"true"}]`,
// 			expectedUpdate: nil,
// 			expectError:    true,
// 		},
// 		{
// 			name:          "replace religious (bool)",
// 			operationJSON: `[{"op":"replace","path":"/personResource/personData/religious","value":true}]`,
// 			expectedUpdate: &parsers.MongoUpdate{
// 				Set:         bson.M{"personResource.personData.religious": true},
// 				Unset:       bson.M{},
// 				Push:        bson.M{},
// 				Pull:        bson.M{},
// 				SetOnInsert: bson.M{},
// 				Inc:         bson.M{},
// 			},
// 			expectError: false,
// 		},
// 		{
// 			name:           "set religious to null (invalid value)",
// 			operationJSON:  `[{"op":"replace","path":"/personResource/personData/religious","value":null}]`,
// 			expectedUpdate: nil,
// 			expectError:    true,
// 		},
// 		{
// 			name:          "replace likesCheesecake (bool)",
// 			operationJSON: `[{"op":"replace","path":"/personResource/likesCheesecake","value":false}]`,
// 			expectedUpdate: &parsers.MongoUpdate{
// 				Set:         bson.M{"personResource.likesCheesecake": false},
// 				Unset:       bson.M{},
// 				Push:        bson.M{},
// 				Pull:        bson.M{},
// 				SetOnInsert: bson.M{},
// 				Inc:         bson.M{},
// 			},
// 			expectError: false,
// 		},
// 		{
// 			name:           "replace likesCheesecake with int (type mismatch)",
// 			operationJSON:  `[{"op":"replace","path":"/personResource/likesCheesecake","value":1}]`,
// 			expectedUpdate: nil,
// 			expectError:    true,
// 		},
//
// 		// ======================
// 		// 3. Integer Fields — Signed
// 		// ======================
// 		{
// 			name:          "replace ageInt (int)",
// 			operationJSON: `[{"op":"replace","path":"/personResource/ageInt","value":42}]`,
// 			expectedUpdate: &parsers.MongoUpdate{
// 				Set:         bson.M{"personResource.ageInt": int(42)},
// 				Unset:       bson.M{},
// 				Push:        bson.M{},
// 				Pull:        bson.M{},
// 				SetOnInsert: bson.M{},
// 				Inc:         bson.M{},
// 			},
// 			expectError: false,
// 		},
// 		{
// 			name:           "replace ageInt with float (type mismatch)",
// 			operationJSON:  `[{"op":"replace","path":"/personResource/ageInt","value":42.5}]`,
// 			expectedUpdate: nil,
// 			expectError:    true,
// 		},
// 		{
// 			name:          "replace ageInt8 (valid range)",
// 			operationJSON: `[{"op":"replace","path":"/personResource/ageInt8","value":100}]`,
// 			expectedUpdate: &parsers.MongoUpdate{
// 				Set:         bson.M{"personResource.ageInt8": int8(100)},
// 				Unset:       bson.M{},
// 				Push:        bson.M{},
// 				Pull:        bson.M{},
// 				SetOnInsert: bson.M{},
// 				Inc:         bson.M{},
// 			},
// 			expectError: false,
// 		},
// 		{
// 			name:           "replace ageInt8 (out of range → should error)",
// 			operationJSON:  `[{"op":"replace","path":"/personResource/ageInt8","value":300}]`,
// 			expectedUpdate: nil,
// 			expectError:    true,
// 		},
// 		{
// 			name:          "replace ageInt16",
// 			operationJSON: `[{"op":"replace","path":"/personResource/ageInt16","value":-30000}]`,
// 			expectedUpdate: &parsers.MongoUpdate{
// 				Set:         bson.M{"personResource.ageInt16": int16(-30000)},
// 				Unset:       bson.M{},
// 				Push:        bson.M{},
// 				Pull:        bson.M{},
// 				SetOnInsert: bson.M{},
// 				Inc:         bson.M{},
// 			},
// 			expectError: false,
// 		},
// 		{
// 			name:          "replace ageInt32",
// 			operationJSON: `[{"op":"replace","path":"/personResource/ageInt32","value":2147483647}]`,
// 			expectedUpdate: &parsers.MongoUpdate{
// 				Set:         bson.M{"personResource.ageInt32": int32(2147483647)},
// 				Unset:       bson.M{},
// 				Push:        bson.M{},
// 				Pull:        bson.M{},
// 				SetOnInsert: bson.M{},
// 				Inc:         bson.M{},
// 			},
// 			expectError: false,
// 		},
// 		{
// 			name:          "replace ageInt64",
// 			operationJSON: `[{"op":"replace","path":"/personResource/ageInt64","value":9223372036854775807}]`,
// 			expectedUpdate: &parsers.MongoUpdate{
// 				Set:         bson.M{"personResource.ageInt64": int64(9223372036854775807)},
// 				Unset:       bson.M{},
// 				Push:        bson.M{},
// 				Pull:        bson.M{},
// 				SetOnInsert: bson.M{},
// 				Inc:         bson.M{},
// 			},
// 			expectError: false,
// 		},
// 		{
// 			name:           "replace ageInt64 with string (should fail)",
// 			operationJSON:  `[{"op":"replace","path":"/personResource/ageInt64","value":"123"}]`,
// 			expectedUpdate: nil,
// 			expectError:    true,
// 		},
//
// 		// ======================
// 		// 4. Integer Fields — Unsigned
// 		// ======================
// 		{
// 			name:          "replace ageUint",
// 			operationJSON: `[{"op":"replace","path":"/personResource/ageUint","value":25}]`,
// 			expectedUpdate: &parsers.MongoUpdate{
// 				Set:         bson.M{"personResource.ageUint": uint(25)},
// 				Unset:       bson.M{},
// 				Push:        bson.M{},
// 				Pull:        bson.M{},
// 				SetOnInsert: bson.M{},
// 				Inc:         bson.M{},
// 			},
// 			expectError: false,
// 		},
// 		{
// 			name:          "replace ageUint8 (max)",
// 			operationJSON: `[{"op":"replace","path":"/personResource/ageUint8","value":255}]`,
// 			expectedUpdate: &parsers.MongoUpdate{
// 				Set:         bson.M{"personResource.ageUint8": uint8(255)},
// 				Unset:       bson.M{},
// 				Push:        bson.M{},
// 				Pull:        bson.M{},
// 				SetOnInsert: bson.M{},
// 				Inc:         bson.M{},
// 			},
// 			expectError: false,
// 		},
// 		{
// 			name:           "replace ageUint8 (negative → error)",
// 			operationJSON:  `[{"op":"replace","path":"/personResource/ageUint8","value":-1}]`,
// 			expectedUpdate: nil,
// 			expectError:    true,
// 		},
// 		{
// 			name:          "replace ageUint32",
// 			operationJSON: `[{"op":"replace","path":"/personResource/ageUint32","value":4294967295}]`,
// 			expectedUpdate: &parsers.MongoUpdate{
// 				Set:         bson.M{"personResource.ageUint32": uint32(4294967295)},
// 				Unset:       bson.M{},
// 				Push:        bson.M{},
// 				Pull:        bson.M{},
// 				SetOnInsert: bson.M{},
// 				Inc:         bson.M{},
// 			},
// 			expectError: false,
// 		},
// 		{
// 			name:          "replace ageUint64",
// 			operationJSON: `[{"op":"replace","path":"/personResource/ageUint64","value":18446744073709551615}]`,
// 			expectedUpdate: &parsers.MongoUpdate{
// 				Set:         bson.M{"personResource.ageUint64": uint64(18446744073709551615)},
// 				Unset:       bson.M{},
// 				Push:        bson.M{},
// 				Pull:        bson.M{},
// 				SetOnInsert: bson.M{},
// 				Inc:         bson.M{},
// 			},
// 			expectError: false,
// 		},
// 		{
// 			name:           "replace ageUint with negative (should error)",
// 			operationJSON:  `[{"op":"replace","path":"/personResource/ageUint","value":-5}]`,
// 			expectedUpdate: nil,
// 			expectError:    true,
// 		},
//
// 		// ======================
// 		// 5. Float Fields
// 		// ======================
// 		{
// 			name:          "replace ageFloat32",
// 			operationJSON: `[{"op":"replace","path":"/personResource/ageFloat32","value":30.5}]`,
// 			expectedUpdate: &parsers.MongoUpdate{
// 				Set:         bson.M{"personResource.ageFloat32": float32(30.5)},
// 				Unset:       bson.M{},
// 				Push:        bson.M{},
// 				Pull:        bson.M{},
// 				SetOnInsert: bson.M{},
// 				Inc:         bson.M{},
// 			},
// 			expectError: false,
// 		},
// 		{
// 			name:          "replace ageFloat64",
// 			operationJSON: `[{"op":"replace","path":"/personResource/ageFloat64","value":30.123456789}]`,
// 			expectedUpdate: &parsers.MongoUpdate{
// 				Set:         bson.M{"personResource.ageFloat64": float64(30.123456789)},
// 				Unset:       bson.M{},
// 				Push:        bson.M{},
// 				Pull:        bson.M{},
// 				SetOnInsert: bson.M{},
// 				Inc:         bson.M{},
// 			},
// 			expectError: false,
// 		},
// 		{
// 			name:          "replace ageFloat32 with int (allowed)",
// 			operationJSON: `[{"op":"replace","path":"/personResource/ageFloat32","value":100}]`,
// 			expectedUpdate: &parsers.MongoUpdate{
// 				Set:         bson.M{"personResource.ageFloat32": float32(100)},
// 				Unset:       bson.M{},
// 				Push:        bson.M{},
// 				Pull:        bson.M{},
// 				SetOnInsert: bson.M{},
// 				Inc:         bson.M{},
// 			},
// 			expectError: false,
// 		},
// 		{
// 			name:           "replace ageFloat32 with string (should fail)",
// 			operationJSON:  `[{"op":"replace","path":"/personResource/ageFloat32","value":"99.9"}]`,
// 			expectedUpdate: nil,
// 			expectError:    true,
// 		},
//
// 		// ======================
// 		// 6. PersonData Fields
// 		// ======================
// 		{
// 			name:          "replace personalNumber (int)",
// 			operationJSON: `[{"op":"replace","path":"/personResource/personData/personalNumber","value":999888}]`,
// 			expectedUpdate: &parsers.MongoUpdate{
// 				Set:         bson.M{"personResource.personData.personalNumber": int(999888)},
// 				Unset:       bson.M{},
// 				Push:        bson.M{},
// 				Pull:        bson.M{},
// 				SetOnInsert: bson.M{},
// 				Inc:         bson.M{},
// 			},
// 			expectError: false,
// 		},
// 		{
// 			name:           "replace personalNumber with float (should fail)",
// 			operationJSON:  `[{"op":"replace","path":"/personResource/personData/personalNumber","value":123.45}]`,
// 			expectedUpdate: nil,
// 			expectError:    true,
// 		},
// 		{
// 			name:           "replace personalNumber with string (should fail)",
// 			operationJSON:  `[{"op":"replace","path":"/personResource/personData/personalNumber","value":"123"}]`,
// 			expectedUpdate: nil,
// 			expectError:    true,
// 		},
//
// 		// ======================
// 		// 7. Tags Map
// 		// ======================
// 		{
// 			name:          "replace whole tags map",
// 			operationJSON: `[{"op":"replace","path":"/personResource/tags","value":{"role":"admin"}}]`,
// 			expectedUpdate: &parsers.MongoUpdate{
// 				Set:         bson.M{"personResource.tags": bson.M{"role": "admin"}},
// 				Unset:       bson.M{},
// 				Push:        bson.M{},
// 				Pull:        bson.M{},
// 				SetOnInsert: bson.M{},
// 				Inc:         bson.M{},
// 			},
// 			expectError: false,
// 		},
// 		{
// 			name:          "add to tags map",
// 			operationJSON: `[{"op":"add","path":"/personResource/tags/category","value":"premium"}]`,
// 			expectedUpdate: &parsers.MongoUpdate{
// 				Set:         bson.M{"personResource.tags.category": "premium"},
// 				Unset:       bson.M{},
// 				Push:        bson.M{},
// 				Pull:        bson.M{},
// 				SetOnInsert: bson.M{},
// 				Inc:         bson.M{},
// 			},
// 			expectError: false,
// 		},
// 		{
// 			name:           "replace tags with non-map (should fail)",
// 			operationJSON:  `[{"op":"replace","path":"/personResource/tags","value":"not-a-map"}]`,
// 			expectedUpdate: nil,
// 			expectError:    true,
// 		},
// 		{
// 			name:          "remove tags subfield",
// 			operationJSON: `[{"op":"remove","path":"/personResource/tags/oldField"}]`,
// 			expectedUpdate: &parsers.MongoUpdate{
// 				Unset:       bson.M{"personResource.tags.oldField": ""},
// 				Set:         bson.M{},
// 				Push:        bson.M{},
// 				Pull:        bson.M{},
// 				SetOnInsert: bson.M{},
// 				Inc:         bson.M{},
// 			},
// 			expectError: false,
// 		},
//
// 		// ======================
// 		// 8. Inlined CheeseCakeLiker
// 		// ======================
// 		{
// 			name:          "replace likesCheesecake",
// 			operationJSON: `[{"op":"replace","path":"/personResource/likesCheesecake","value":true}]`,
// 			expectedUpdate: &parsers.MongoUpdate{
// 				Set:         bson.M{"personResource.likesCheesecake": true},
// 				Unset:       bson.M{},
// 				Push:        bson.M{},
// 				Pull:        bson.M{},
// 				SetOnInsert: bson.M{},
// 				Inc:         bson.M{},
// 			},
// 			expectError: false,
// 		},
//
// 		// ======================
// 		// 9. ID & dynamic fields
// 		// ======================
// 		{
// 			name:          "replace id (string)",
// 			operationJSON: `[{"op":"replace","path":"/personResource/id","value":"new-p789"}]`,
// 			expectedUpdate: &parsers.MongoUpdate{
// 				Set:         bson.M{"personResource.id": "new-p789"},
// 				Unset:       bson.M{},
// 				Push:        bson.M{},
// 				Pull:        bson.M{},
// 				SetOnInsert: bson.M{},
// 				Inc:         bson.M{},
// 			},
// 			expectError: false,
// 		},
// 		{
// 			name:          "add to nonexistent path (dynamic field)",
// 			operationJSON: `[{"op":"add","path":"/personResource/nonexistent","value":"ok"}]`,
// 			expectedUpdate: &parsers.MongoUpdate{
// 				Set:         bson.M{"personResource.nonexistent": "ok"},
// 				Unset:       bson.M{},
// 				Push:        bson.M{},
// 				Pull:        bson.M{},
// 				SetOnInsert: bson.M{},
// 				Inc:         bson.M{},
// 			},
// 			expectError: false,
// 		},
//
// 		// ======================
// 		// 10. Zero / nil / empty
// 		// ======================
// 		{
// 			name:          "set ageInt to zero",
// 			operationJSON: `[{"op":"replace","path":"/personResource/ageInt","value":0}]`,
// 			expectedUpdate: &parsers.MongoUpdate{
// 				Set:         bson.M{"personResource.ageInt": 0},
// 				Unset:       bson.M{},
// 				Push:        bson.M{},
// 				Pull:        bson.M{},
// 				SetOnInsert: bson.M{},
// 				Inc:         bson.M{},
// 			},
// 			expectError: false,
// 		},
// 		{
// 			name:          "set religious to false",
// 			operationJSON: `[{"op":"replace","path":"/personResource/personData/religious","value":false}]`,
// 			expectedUpdate: &parsers.MongoUpdate{
// 				Set:         bson.M{"personResource.personData.religious": false},
// 				Unset:       bson.M{},
// 				Push:        bson.M{},
// 				Pull:        bson.M{},
// 				SetOnInsert: bson.M{},
// 				Inc:         bson.M{},
// 			},
// 			expectError: false,
// 		},
// 		{
// 			name:          "replace tags with empty map",
// 			operationJSON: `[{"op":"replace","path":"/personResource/tags","value":{}}]`,
// 			expectedUpdate: &parsers.MongoUpdate{
// 				Set:         bson.M{"personResource.tags": bson.M{}},
// 				Unset:       bson.M{},
// 				Push:        bson.M{},
// 				Pull:        bson.M{},
// 				SetOnInsert: bson.M{},
// 				Inc:         bson.M{},
// 			},
// 			expectError: false,
// 		},
//
// 		// ======================
// 		// 11. Original test cases (updated to new fields)
// 		// ======================
// 		{
// 			name:          "add to missing field (creates path)",
// 			operationJSON: `[{"op":"add","path":"/personResource/name","value":"Bob"}]`,
// 			expectedUpdate: &parsers.MongoUpdate{
// 				Set:         bson.M{"personResource.name": "Bob"},
// 				Unset:       bson.M{},
// 				Push:        bson.M{},
// 				Pull:        bson.M{},
// 				SetOnInsert: bson.M{},
// 				Inc:         bson.M{},
// 			},
// 			expectError: false,
// 		},
// 		{
// 			name:          "replace existing field",
// 			operationJSON: `[{"op":"replace","path":"/personResource/name","value":"Charlie"}]`,
// 			expectedUpdate: &parsers.MongoUpdate{
// 				Set:         bson.M{"personResource.name": "Charlie"},
// 				Unset:       bson.M{},
// 				Push:        bson.M{},
// 				Pull:        bson.M{},
// 				SetOnInsert: bson.M{},
// 				Inc:         bson.M{},
// 			},
// 			expectError: false,
// 		},
// 		{
// 			name:          "remove field",
// 			operationJSON: `[{"op":"remove","path":"/personResource/name"}]`,
// 			expectedUpdate: &parsers.MongoUpdate{
// 				Unset:       bson.M{"personResource.name": ""},
// 				Set:         bson.M{},
// 				Push:        bson.M{},
// 				Pull:        bson.M{},
// 				SetOnInsert: bson.M{},
// 				Inc:         bson.M{},
// 			},
// 			expectError: false,
// 		},
// 		{
// 			name:          "add nested object (PersonData)",
// 			operationJSON: `[{"op":"add","path":"/personResource/personData","value":{"personalNumber":987654,"religious":true}}]`,
// 			expectedUpdate: &parsers.MongoUpdate{
// 				Set: bson.M{"personResource.personData": bson.M{
// 					"personalNumber": 987654,
// 					"religious":      true,
// 				}},
// 				Unset:       bson.M{},
// 				Push:        bson.M{},
// 				Pull:        bson.M{},
// 				SetOnInsert: bson.M{},
// 				Inc:         bson.M{},
// 			},
// 			expectError: false,
// 		},
// 		{
// 			name:          "add inlined field (likesCheesecake)",
// 			operationJSON: `[{"op":"add","path":"/personResource/likesCheesecake","value":false}]`,
// 			expectedUpdate: &parsers.MongoUpdate{
// 				Set:         bson.M{"personResource.likesCheesecake": false},
// 				Unset:       bson.M{},
// 				Push:        bson.M{},
// 				Pull:        bson.M{},
// 				SetOnInsert: bson.M{},
// 				Inc:         bson.M{},
// 			},
// 			expectError: false,
// 		},
// 		{
// 			name:           "add inlined field with string value (should fail)",
// 			operationJSON:  `[{"op":"add","path":"/personResource/likesCheesecake","value":"false"}]`,
// 			expectedUpdate: nil,
// 			expectError:    true,
// 		},
// 		{
// 			name:          "remove inlined field",
// 			operationJSON: `[{"op":"remove","path":"/personResource/likesCheesecake"}]`,
// 			expectedUpdate: &parsers.MongoUpdate{
// 				Unset:       bson.M{"personResource.likesCheesecake": ""},
// 				Set:         bson.M{},
// 				Push:        bson.M{},
// 				Pull:        bson.M{},
// 				SetOnInsert: bson.M{},
// 				Inc:         bson.M{},
// 			},
// 			expectError: false,
// 		},
// 		{
// 			name:          "update ageInt (int)",
// 			operationJSON: `[{"op":"replace","path":"/personResource/ageInt","value":31}]`,
// 			expectedUpdate: &parsers.MongoUpdate{
// 				Set:         bson.M{"personResource.ageInt": 31},
// 				Unset:       bson.M{},
// 				Push:        bson.M{},
// 				Pull:        bson.M{},
// 				SetOnInsert: bson.M{},
// 				Inc:         bson.M{},
// 			},
// 			expectError: false,
// 		},
// 		{
// 			name:           "update ageInt (string → type mismatch)",
// 			operationJSON:  `[{"op":"replace","path":"/personResource/ageInt","value":"31"}]`,
// 			expectedUpdate: nil,
// 			expectError:    true,
// 		},
// 		{
// 			name:          "set alive = false (bool)",
// 			operationJSON: `[{"op":"replace","path":"/personResource/alive","value":false}]`,
// 			expectedUpdate: &parsers.MongoUpdate{
// 				Set:         bson.M{"personResource.alive": false},
// 				Unset:       bson.M{},
// 				Push:        bson.M{},
// 				Pull:        bson.M{},
// 				SetOnInsert: bson.M{},
// 				Inc:         bson.M{},
// 			},
// 			expectError: false,
// 		},
// 		{
// 			name:          "set religious = null",
// 			operationJSON: `[{"op":"replace","path":"/personResource/personData/religious","value":null}]`,
// 			expectedUpdate: &parsers.MongoUpdate{
// 				Set:         bson.M{"personResource.personData.religious": false},
// 				Unset:       bson.M{},
// 				Push:        bson.M{},
// 				Pull:        bson.M{},
// 				SetOnInsert: bson.M{},
// 				Inc:         bson.M{},
// 			},
// 			expectError: false,
// 		},
// 	}
//
// 	// Run tests
// 	for _, tc := range tests {
// 		t.Run(tc.name, func(t *testing.T) {
//
// 			actualUpdate, err := parsers.Parse([]byte(tc.operationJSON), models.Resource{})
//
// 			if tc.expectError {
// 				require.Error(t, err, "did not fail at parsing when expected")
// 			} else {
// 				require.NoErrorf(t, err, "failed at parsing when not expected,: %v", err)
// 			}
//
// 			// Compare (ignore nil vs empty maps)
// 			require.Equal(t, tc.expectedUpdate, actualUpdate,
// 				"Mongo update mismatch\nExpected: %+v\nGot:      %+v", tc.expectedUpdate, actualUpdate)
// 		})
// 	}
// }
//
// func sampleResource() *models.Resource {
// 	return &models.Resource{
// 		ID: "doc123",
// 		PersonResource: &models.Person{
// 			ID:   "p456",
// 			Name: "Alice",
// 			// Optional: set multiple age fields — here we set AgeInt = 30 to match original intent
// 			Alive:      true,
// 			AgeInt:     30,
// 			AgeInt8:    30,
// 			AgeInt16:   30,
// 			AgeInt32:   30,
// 			AgeInt64:   30,
// 			AgeUint:    30,
// 			AgeUint8:   30,
// 			AgeUint16:  30,
// 			AgeUint32:  30,
// 			AgeUint64:  30,
// 			AgeFloat32: 30.0,
// 			AgeFloat64: 30.0,
// 			Tags: map[string]string{
// 				"priority": "high",
// 				"verified": "true",
// 			},
// 			PersonData: models.PersonData{
// 				PersonalNumber: 123456,
// 				Religious:      false,
// 			},
// 			CheeseCakeLiker: models.CheeseCakeLiker{
// 				LikesCheesecake: true,
// 			},
// 		},
// 	}
// }
//
// // BSON doc version
// func sampleResourceBSON() bson.M {
// 	return bson.M{
// 		"_id": "doc123",
// 		"personResource": bson.M{
// 			"id":         "p456",
// 			"name":       "Alice",
// 			"alive":      true,
// 			"ageInt":     30,
// 			"ageInt8":    30,
// 			"ageInt16":   30,
// 			"ageInt32":   30,
// 			"ageInt64":   30,
// 			"ageUint":    30,
// 			"ageUint8":   30,
// 			"ageUint16":  30,
// 			"ageUint32":  30,
// 			"ageUint64":  30,
// 			"ageFloat32": 30.0,
// 			"ageFloat64": 30.0,
// 			"tags": bson.M{
// 				"priority": "high",
// 				"verified": "true",
// 			},
// 			"personData": bson.M{
// 				"personalNumber": 123456,
// 				"religious":      false,
// 			},
// 			"likesCheesecake": true, // from inline CheeseCakeLiker
// 		},
// 	}
// }
