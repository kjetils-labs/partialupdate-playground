package parsers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"reflect"
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
	return update
}

// Parse parses a []byte to []Operation, validates the request matches the RC6902,
// validates it matches the requested type and outputs it as a MongoUpdate.
func Parse[T any](request []byte) (*MongoUpdate, error) {

	operations := make([]Operation, 0)

	err := json.Unmarshal(request, &operations)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal request. %w", err)
	}

	// Validate the request operation(s) are valid per the spec.
	// Does not validate the path(s) are valid for the resource.
	err = ValidatePatch(operations)
	if err != nil {
		return nil, fmt.Errorf("request operation is invalid. %w", err)
	}

	typ := reflect.TypeFor[T]()
	fieldMap := FieldMaps.Get(typ.Name())
	if fieldMap == nil {
		err = FieldMaps.Add(typ)
		if err != nil {
		}
	}

	if readBeforeWrite(operations) {
		return nil, fmt.Errorf("read before write operations are unsupported")
	}

	output := NewMongoUpdate()

	err = parse(operations, output)
	if err != nil {
		return nil, fmt.Errorf("failed to parse operations to instructions. %w", err)
	}

	return output, nil
}

func parse(operations []Operation, update *MongoUpdate) error {

	for _, operation := range operations {

		// path, err := toMongoPath(operation.Path)
		// if err != nil {
		// 	return fmt.Errorf("failed to convert path %v to mongo path. %w", operation.Path, err)
		// }
		// slog.Info("converted path", "original_path", operation.Path, "mongo_path", path)
		path, err := convertJSONPathMongo(operation.Path)
		if err != nil {
			return fmt.Errorf("path error in op %q: %w", operation.Op, err)
		}
		slog.Info("converted path", "original_path", operation.Path, "path", path)

		switch operation.Op {
		case OperationTypeAdd:
			if err := applyAdd(update, operation, path); err != nil {
				return fmt.Errorf("failed to apply add. %w", err)
			}
		case OperationTypeRemove:
			if err := applyRemove(update, path); err != nil {
				return fmt.Errorf("failed to apply remove. %w", err)
			}
		case OperationTypeReplace:
			if err := applyReplace(update, operation, path); err != nil {
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
