package v1

import (
	"encoding/json"
	"fmt"
	"reflect"

	"go.mongodb.org/mongo-driver/bson"
)

type MongoUpdate struct {
	Set   bson.M `bson:",omitempty"`
	Unset bson.M `bson:",omitempty"`
	Push  bson.M `bson:",omitempty"`
	Pull  bson.M `bson:",omitempty"`
}

func NewMongoUpdate() *MongoUpdate {
	return &MongoUpdate{
		Set:   bson.M{},
		Unset: bson.M{},
		Push:  bson.M{},
		Pull:  bson.M{},
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
	return update
}

// Parse parses a []byte to []Operation, validates the request matches the RC6902,
// validates it matches the requested type and outputs it as a MongoUpdate.
// resourceType expects an instance of the type being updated.
// The supported operations are "add", "remove" and "replace". The "move", "copy" and "test" operations are not supported.
func Parse(request []byte, resourceType any) (*MongoUpdate, error) {

	operations := make([]Operation, 0)
	err := json.Unmarshal(request, &operations)
	if err != nil {
		return nil, fmt.Errorf("failed to parse request: %w", err)
	}

	err = ValidatePatch(operations)
	if err != nil {
		return nil, fmt.Errorf("invalid patch: %w", err)
	}

	update := NewMongoUpdate()

	for _, operation := range operations {

		field, err := WalkStruct(operation.Path, resourceType)
		if err != nil {
			return nil, fmt.Errorf("invalid path %q for type %v: %w", operation.Path, reflect.TypeOf(resourceType).Name(), err)
		}

		err = ApplyOperation(update, operation, field)
		if err != nil {
			return nil, fmt.Errorf("failed to apply operation %q to path %q: %w", operation.Op, operation.Path, err)
		}

	}

	return update, nil
}
