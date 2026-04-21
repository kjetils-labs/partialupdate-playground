package parsers_test

import (
	"encoding/json"
	"partialupdate/internal/parsers"
	"testing"
)

type testCase struct {
	name        string // descriptive name for the sub‑test
	opsJSON     string // JSON representation of []Operation
	wantErr     bool   // true → an error is expected
	errContains string // optional substring that must appear in the error message
}

func TestValidatePatch(t *testing.T) {

	cases := []testCase{
		{
			name: "valid add (value required)",
			opsJSON: `[{
                "op":"add",
                "path":"/personResource/name",
                "value":"Alice"
            }]`,
			wantErr: false,
		},
		{
			name: "valid replace",
			opsJSON: `[{
                "op":"replace",
                "path":"/personResource/age",
                "value":42
            }]`,
			wantErr: false,
		},
		{
			name: "valid remove",
			opsJSON: `[{
                "op":"remove",
                "path":"/personResource/id"
            }]`,
			wantErr: false,
		},
		{
			name: "valid copy",
			opsJSON: `[{
                "op":"copy",
                "from":"/personResource/name",
                "path":"/personResource/displayName"
            }]`,
			wantErr: false,
		},
		{
			name: "valid move",
			opsJSON: `[{
                "op":"move",
                "from":"/personResource/alive",
                "path":"/personResource/isAlive"
            }]`,
			wantErr: false,
		},
		{
			name: "valid test",
			opsJSON: `[{
                "op":"test",
                "path":"/personResource/likesCheesecake",
                "value":false
            }]`,
			wantErr: false,
		},
		{
			name: "missing value on add",
			opsJSON: `[{
                "op":"add",
                "path":"/personResource/name"
            }]`,
			wantErr:     true,
			errContains: "operation \"add\" requires 'value'",
		},
		{
			name: "missing value on replace",
			opsJSON: `[{
                "op":"replace",
                "path":"/personResource/age"
            }]`,
			wantErr:     true,
			errContains: "operation \"replace\" requires 'value'",
		},
		{
			name: "value supplied on copy",
			opsJSON: `[{
                "op":"copy",
                "from":"/personResource/name",
                "path":"/personResource/displayName",
                "value":"oops"
            }]`,
			wantErr:     false,
			errContains: "",
		},
		{
			name: "missing from on copy",
			opsJSON: `[{
                "op":"copy",
                "path":"/personResource/displayName"
            }]`,
			wantErr:     true,
			errContains: "operation \"copy\" requires 'from'",
		},
		{
			name: "missing from on move",
			opsJSON: `[{
                "op":"move",
                "path":"/personResource/newField"
            }]`,
			wantErr:     true,
			errContains: "operation \"move\" requires 'from'",
		},
		{
			name:    "empty patch – always valid",
			opsJSON: `[]`,
			wantErr: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Decode the JSON test case into []Operation.
			var ops []parsers.Operation
			if err := json.Unmarshal([]byte(tc.opsJSON), &ops); err != nil {
				if tc.wantErr && contains(err.Error(), tc.errContains) {
					return
				} else {
					t.Fatalf("failed to unmarshal test case JSON: %v", err)
				}
			}

			// Run the validator.
			err := parsers.ValidatePatch(ops)

			switch {
			case tc.wantErr && err == nil:
				t.Fatalf("expected an error but got nil")
			case tc.wantErr && err != nil && tc.errContains != "" && !contains(err.Error(), tc.errContains):
				t.Fatalf("error does not contain expected substring.\nGot: %q\nWant substring: %q",
					err.Error(), tc.errContains)
			case !tc.wantErr && err != nil:
				t.Fatalf("unexpected error: %v", err)
			}

		})
	}
}

// simple substring check – avoids pulling in the strings package for a single call.
func contains(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(substr) > len(s) {
		return false
	}
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
