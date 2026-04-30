package parsers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"slices"

	"go.mongodb.org/mongo-driver/bson"
)

type MongoUpdate struct {
	Set         bson.M `bson:",omitempty"`
	Unset       bson.M `bson:",omitempty"`
	Push        bson.M `bson:",omitempty"`
	Pull        bson.M `bson:",omitempty"`
	SetOnInsert bson.M `bson:",omitempty"`
	Inc         bson.M `bson:",omitempty"`
}

func NewMongoUpdate() *MongoUpdate {
	return &MongoUpdate{
		Set:         bson.M{},
		Unset:       bson.M{},
		Push:        bson.M{},
		Pull:        bson.M{},
		SetOnInsert: bson.M{},
		Inc:         bson.M{},
	}
}

func (mu *MongoUpdate) ToBSON() bson.D {
	update := bson.D{}
	if mu.Set != nil {
		update = append(update, bson.E{Key: "$set", Value: mu.Set})
	}
	if mu.Unset != nil {
		update = append(update, bson.E{Key: "$unset", Value: mu.Unset})
	}
	if mu.Push != nil {
		update = append(update, bson.E{Key: "$push", Value: mu.Push})
	}
	if mu.Pull != nil {
		update = append(update, bson.E{Key: "$pull", Value: mu.Pull})
	}
	if mu.SetOnInsert != nil {
		update = append(update, bson.E{Key: "$setOnInsert", Value: mu.SetOnInsert})
	}
	if mu.Inc != nil {
		update = append(update, bson.E{Key: "$inc", Value: mu.Inc})
	}
	return update
}

// Parse parses a []byte to []Operation, validates the request matches the RC6902,
// validates it matches the requested type and outputs it as a MongoUpdate.
// resourceType expects an instance of the type being updated.
func Parse(request []byte, resourceType any) (*MongoUpdate, error) {

	operations := make([]Operation, 0)

	err := json.Unmarshal(request, &operations)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal request. %w", err)
	}
	slog.Debug("parsing patch request", "operations_count", len(operations))

	// Validate the request operation(s) are valid per the spec.
	// Does not validate the path(s) are valid for the resource.
	err = ValidatePatch(operations)
	if err != nil {
		return nil, fmt.Errorf("request operation is invalid. %w", err)
	}

	fields := WalkStruct(resourceType)

	if len(fields) == 0 {
		return nil, fmt.Errorf("no fields found for type")
	}

	if readBeforeWrite(operations) {
		return nil, fmt.Errorf("read before write operations are unsupported")
	}

	output := NewMongoUpdate()

	err = parse(operations, fields, output)
	if err != nil {
		return nil, fmt.Errorf("failed to parse operations to instructions. %w", err)
	}

	return output, nil
}

func parse(operations []Operation, fields []FieldInfo, update *MongoUpdate) error {

	for _, operation := range operations {

		var field *FieldInfo
		for _, p := range fields {
			if p.Path == operation.Path {
				field = &p
				break
			}
		}
		if field == nil {
			return fmt.Errorf("path %v is not a valid path for type", operation.Path)
		}

		path, err := convertJSONPathMongo(operation.Path)
		if err != nil {
			return fmt.Errorf("path error in op %q: %w", operation.Op, err)
		}

		switch operation.Op {
		case OperationTypeAdd:
			if err := applyAdd(update, operation, field, path); err != nil {
				return fmt.Errorf("failed to apply add. %w", err)
			}
		case OperationTypeRemove:
			if err := applyRemove(update, path); err != nil {
				return fmt.Errorf("failed to apply remove. %w", err)
			}
		case OperationTypeReplace:
			if err := applyReplace(update, operation, field, path); err != nil {
				return fmt.Errorf("failed to apply replace. %w", err)
			} // Will not be implemented as that requires read-before-write
		case OperationTypeCopy:
			return fmt.Errorf("not implemented")
		case OperationTypeMove:
			return fmt.Errorf("not implemented")
		case OperationTypeTest:
			return fmt.Errorf("not implemented")
		}
	}

	return nil
}

// getReadBeforeWriteOperations returns operations that require Read before writing operations.
func getReadBeforeWriteOperations() []OperationType {
	return []OperationType{OperationTypeMove, OperationTypeCopy, OperationTypeTest}
}

func readBeforeWrite(operations []Operation) bool {

	for _, operation := range operations {

		if slices.Contains(getReadBeforeWriteOperations(), operation.Op) {
			return true
		}
	}

	return false
}
