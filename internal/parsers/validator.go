package parsers

import "fmt"

// ValidatePatch valides that the operation(s) are valid per the RFC6902 spec,
// that the operation is valid, required field(s) for that type of operation is present.
// It does not validate that the paths used in the operation(s) are valid for the request.
// TODO: Add path validation
func ValidatePatch(patch []Operation) error {
	for i, op := range patch {
		switch op.Op {
		case OperationTypeAdd, OperationTypeReplace, OperationTypeTest:
			if len(op.Value) == 0 || op.Value == nil {
				return fmt.Errorf("operation %d (%s) requires a non‑empty \"value\" field", i, op.Op)
			}
		case OperationTypeRemove:
			if len(op.Value) != 0 || op.From != "" {
				return fmt.Errorf("operation %d (%s) must not contain \"value\" or \"from\"", i, op.Op)
			}
		case OperationTypeCopy, OperationTypeMove:
			if op.From == "" {
				return fmt.Errorf("operation %d (%s) requires a non‑empty \"from\" field", i, op.Op)
			}
			if op.Op == OperationTypeCopy && len(op.Value) != 0 {
				return fmt.Errorf("operation %d (%s) must not contain \"value\"", i, op.Op)
			}
		default:
			return fmt.Errorf("operation %d has unknown op %q", i, op.Op)
		}
	}
	return nil
}
