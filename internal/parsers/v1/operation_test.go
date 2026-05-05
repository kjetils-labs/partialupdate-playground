package v1_test

import (
	"encoding/json"
	v1 "partialupdate/internal/parsers/v1"
	"testing"
)

func TestOperation_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		want     *v1.Operation
		wantErr  bool
		errCheck func(error) bool
	}{
		// --- ADD ---
		{
			name:  "add with value (string)",
			input: `{"op":"add","path":"/a","value":"b"}`,
			want:  &v1.Operation{Op: v1.OperationTypeAdd, Path: "/a", Value: json.RawMessage(`"b"`)},
		},
		{
			name:  "add with value (number)",
			input: `{"op":"add","path":"/count","value":42}`,
			want:  &v1.Operation{Op: v1.OperationTypeAdd, Path: "/count", Value: json.RawMessage(`42`)},
		},
		{
			name:  "add with value (object)",
			input: `{"op":"add","path":"/user","value":{"name":"Alice"}}`,
			want:  &v1.Operation{Op: v1.OperationTypeAdd, Path: "/user", Value: json.RawMessage(`{"name":"Alice"}`)},
		},
		{
			name:     "add missing value",
			input:    `{"op":"add","path":"/x"}`,
			wantErr:  true,
			errCheck: func(err error) bool { return containsStr(err.Error(), "requires 'value'") },
		},

		// --- REMOVE ---
		{
			name:  "remove (no value needed)",
			input: `{"op":"remove","path":"/foo"}`,
			want:  &v1.Operation{Op: v1.OperationTypeRemove, Path: "/foo"},
		},
		{
			name:  "remove with extra fields ignored",
			input: `{"op":"remove","path":"/bar","value":"ignored"}`,
			want:  &v1.Operation{Op: v1.OperationTypeRemove, Path: "/bar"}, // value is ignored (but not validated against)
		},

		// --- REPLACE ---
		{
			name:  "replace with value",
			input: `{"op":"replace","path":"/name","value":"Bob"}`,
			want:  &v1.Operation{Op: v1.OperationTypeReplace, Path: "/name", Value: json.RawMessage(`"Bob"`)},
		},
		{
			name:     "replace missing value",
			input:    `{"op":"replace","path":"/x"}`,
			wantErr:  true,
			errCheck: func(err error) bool { return containsStr(err.Error(), "requires 'value'") },
		},

		// --- MOVE ---
		{
			name:  "move valid",
			input: `{"op":"move","from":"/a","path":"/b"}`,
			want:  &v1.Operation{Op: v1.OperationTypeMove, Path: "/b", From: "/a"},
		},
		{
			name:     "move missing from",
			input:    `{"op":"move","path":"/b"}`,
			wantErr:  true,
			errCheck: func(err error) bool { return containsStr(err.Error(), "requires 'from'") },
		},

		// --- COPY ---
		{
			name:  "copy valid",
			input: `{"op":"copy","from":"/x","path":"/y"}`,
			want:  &v1.Operation{Op: v1.OperationTypeCopy, Path: "/y", From: "/x"},
		},
		{
			name:     "copy missing from",
			input:    `{"op":"copy","path":"/y"}`,
			wantErr:  true,
			errCheck: func(err error) bool { return containsStr(err.Error(), "requires 'from'") },
		},

		// --- TEST ---
		{
			name:  "test with value",
			input: `{"op":"test","path":"/status","value":"active"}`,
			want:  &v1.Operation{Op: v1.OperationTypeTest, Path: "/status", Value: json.RawMessage(`"active"`)},
		},
		{
			name:     "test missing value",
			input:    `{"op":"test","path":"/x"}`,
			wantErr:  true,
			errCheck: func(err error) bool { return containsStr(err.Error(), "requires 'value'") },
		},

		// --- INVALID/EDGE CASES ---
		{
			name:     "empty path",
			input:    `{"op":"add","path":"","value":1}`,
			wantErr:  true,
			errCheck: func(err error) bool { return containsStr(err.Error(), "path can not be empty") },
		},
		{
			name:    "empty input",
			input:   ``,
			wantErr: true,
		},
		{
			name:  "value can be array",
			input: `{"op":"add","path":"/items","value":[1,2,3]}`,
			want:  &v1.Operation{Op: v1.OperationTypeAdd, Path: "/items", Value: json.RawMessage(`[1,2,3]`)},
		},
		{
			name:  "value can be boolean/null",
			input: `{"op":"test","path":"/flag","value":true}`,
			want:  &v1.Operation{Op: v1.OperationTypeTest, Path: "/flag", Value: json.RawMessage(`true`)},
		},
		{
			name:  "value can be null",
			input: `{"op":"add","path":"/opt","value":null}`,
			want:  &v1.Operation{Op: v1.OperationTypeAdd, Path: "/opt", Value: json.RawMessage(`null`)},
		},

		// --- Whitespace/extra fields (should be ignored) ---
		{
			name:  "extra fields allowed (JSON permissive)",
			input: `{"op":"add","path":"/a","value":"b","extra":"ignored"}`,
			want:  &v1.Operation{Op: v1.OperationTypeAdd, Path: "/a", Value: json.RawMessage(`"b"`)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var op v1.Operation
			err := op.UnmarshalJSON([]byte(tt.input))

			if (err != nil) != tt.wantErr {
				t.Fatalf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil && tt.wantErr {
				if tt.errCheck != nil && !tt.errCheck(err) {
					t.Errorf("error %q doesn’t match expected pattern", err)
				}
				return
			}

			if err == nil && !tt.wantErr {
				if !Equal(tt.want, &op) { // ⚠️ need to implement `Equal` — see below
					t.Errorf("Operation = %+v, want %+v", op, tt.want)
				}
			}
		})
	}
}

// Helper: check if error message contains substring (case-insensitive)
func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Add this to your Operation type for comparison
func Equal(o1, o2 *v1.Operation) bool {
	if o1 == o2 {
		return true
	}
	if o1 == nil || o2 == nil {
		return false
	}
	return o1.Op == o2.Op &&
		o1.Path == o2.Path &&
		o1.From == o2.From &&
		string(o1.Value) == string(o2.Value)
}
