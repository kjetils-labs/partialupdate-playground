package unset

import (
	"errors"
	"partialupdate/internal/models"
	"reflect"

	"go.mongodb.org/mongo-driver/bson"
)

type Request struct {
	Resource  *models.Resource `json:"resource" bson:"resource,omitempty"`
	UnsetMask []string         `json:"unsetMask" bson:"unsetMask,omitempty"`
}

func NewRequest(resource *models.Resource, unsetMask []string) *Request {
	return &Request{
		Resource:  resource,
		UnsetMask: unsetMask,
	}
}

var (
	ErrNoResourceInRequest = errors.New("no resource in request")
	ErrNoValidFields       = errors.New("no valid fields to update")
)

// Parse converts a resource into a MongoDB update document.
//
// Behavior:
// - Generates $set using dot-notation for non-zero fields
// - Generates $unset for fields listed in UnsetMask
// - Inline (embedded) structs are flattened
// - Zero values are ignored unless explicitly unset
// - _id is never updated
//
// Returns:
// - ErrNoResourceInRequest if Resource is nil
// - ErrNoValidFields if no $set or $unset operations are produced
func (r *Request) Parse() (bson.M, error) {
	if r.Resource == nil {
		return nil, ErrNoResourceInRequest
	}

	if err := validateUnsetMask(r.Resource, r.UnsetMask); err != nil {
		return nil, err
	}

	set := bson.M{}
	unset := bson.M{}

	unsetSet := make(map[string]struct{})
	for _, path := range r.UnsetMask {
		unsetSet[normalizePath(path)] = struct{}{}
	}

	v := reflect.ValueOf(r.Resource)
	buildUpdate(v, "", set, unset, unsetSet)

	if len(set) == 0 && len(unset) == 0 {
		return nil, ErrNoValidFields
	}

	update := bson.M{}
	if len(set) > 0 {
		update["$set"] = set
	}
	if len(unset) > 0 {
		update["$unset"] = unset
	}

	return update, nil
}
