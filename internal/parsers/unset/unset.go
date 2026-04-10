package unset

import (
	"partialupdate/internal/models"

	"go.mongodb.org/mongo-driver/bson"
)

type Request struct {
	Resource *models.Resource `json:"resource" bson:"resource,omitempty"`
	// Dot annotation?
	UnsetMask []string `json:"unsetMask" bson:"unsetMask,omitempty"`
}

func NewRequest(resource *models.Resource, unsetMask []string) *Request {
	return &Request{
		Resource:  resource,
		UnsetMask: unsetMask,
	}
}

func (r *Request) Parse() (bson.M, error) {
	update := bson.M{}

	resourceMap := bson.M{}
	if r.Resource != nil {
		data, err := bson.Marshal(r.Resource)
		if err != nil {
			return nil, err
		}
		if err := bson.Unmarshal(data, &resourceMap); err != nil {
			return nil, err
		}
	}

	// Validate logical paths.
	if err := validateUnsetMask(r.Resource, r.UnsetMask); err != nil {
		return nil, err
	}

	// Normalize + apply unset.
	unsetMap := bson.M{}
	for _, path := range r.UnsetMask {
		bsonPath := normalizePath(path)

		unsetMap[bsonPath] = ""
		deleteNested(resourceMap, bsonPath)
	}

	removeEmptyMaps(resourceMap)

	if len(resourceMap) > 0 {
		flatSet := bson.M{}
		flattenMap("", resourceMap, flatSet)

		if len(flatSet) > 0 {
			update["$set"] = flatSet
		}
	}
	if len(unsetMap) > 0 {
		update["$unset"] = unsetMap
	}

	set, setExists := update["$set"].(bson.M)

	if setExists {
		for k, v := range set {
			// If the value is a zero value and it's not in UnsetMask, remove it from $set
			if isZeroValue(v) && !containsUnsetMaskNormalized(unsetMap, k) {
				delete(set, k)
			}
		}
	}

	// If there are any fields left in $set, include them in the update
	if len(set) > 0 {
		update["$set"] = set
	} else {
		// If no fields remain, remove the $set key from the update
		delete(update, "$set")
	}

	// Ensure the $unset remains intact
	if len(unsetMap) > 0 {
		update["$unset"] = unsetMap
	} else {
		// Remove $unset if it’s empty
		delete(update, "$unset")
	}

	return update, nil
}
