package v1

import (
	"encoding/json"
	"fmt"
)

type OperationType string

const (
	OperationTypeAdd     OperationType = "add"
	OperationTypeRemove  OperationType = "remove"
	OperationTypeReplace OperationType = "replace"
	OperationTypeMove    OperationType = "move"
	OperationTypeCopy    OperationType = "copy"
	OperationTypeTest    OperationType = "test"
)

type Operation struct {
	Op    OperationType   `json:"op"`
	Path  string          `json:"path"`
	From  string          `json:"from,omitempty"`
	Value json.RawMessage `json:"value,omitempty"`
}

func (o *Operation) UnmarshalJSON(b []byte) error {

	data := struct {
		Op    string          `json:"op"`
		Path  string          `json:"path"`
		From  string          `json:"from,omitempty"`
		Value json.RawMessage `json:"value,omitempty"`
	}{}

	err := json.Unmarshal(b, &data)
	if err != nil {
		return fmt.Errorf("failed to unmarshal operation %v. %w", data.Op, err)
	}

	o.Op = OperationType(data.Op)
	if data.Path == "" {
		return fmt.Errorf("path can not be empty")
	}
	o.Path = data.Path

	if data.From != "" {
		o.From = data.From
	}
	// Validate required value for operations that need it
	switch o.Op {
	case OperationTypeAdd, OperationTypeReplace, OperationTypeTest:
		if len(data.Value) == 0 || data.Value == nil {
			return fmt.Errorf("operation %q requires 'value' field", o.Op)
		}
		o.Value = data.Value
	case OperationTypeMove, OperationTypeCopy:
		if data.From == "" {
			return fmt.Errorf("operation %q requires 'from' field", o.Op)
		}
	default:
		// remove: value not required
	}

	return nil
}
