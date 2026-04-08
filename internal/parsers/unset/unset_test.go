package unset_test

import (
	"partialupdate/internal/models"
	"partialupdate/internal/parsers/unset"
	"reflect"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

func TestParse_Person_SetAll(t *testing.T) {
	req := unset.Request{
		Resource: models.Person{
			ID:    "123",
			Name:  "John",
			Alive: true,
			Age:   30,
			PersonData: models.PersonData{
				PersonalNumber: 123456,
				Religious:      false,
			},
		},
	}

	result, err := req.Parse()
	if err != nil {
		t.Fatal(err)
	}

	expected := normalize(bson.M{
		"$set": bson.M{
			"_id":   "123",
			"name":  "John",
			"alive": true,
			"age":   30,
		},
	})

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %+v, got %+v", expected, result)
	}
}

func TestParse_Unset(t *testing.T) {
	req := unset.Request{
		Resource: models.Person{
			ID:    "123",
			Name:  "John",
			Alive: true,
			Age:   30,
		},
		UnsetMask: []string{"age"},
	}

	result, err := req.Parse()
	if err != nil {
		t.Fatal(err)
	}

	expected := normalize(bson.M{
		"$set": bson.M{
			"_id":   "123",
			"name":  "John",
			"alive": true,
		},
		"$unset": bson.M{
			"age": "",
		},
	})

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %+v, got %+v", expected, result)
	}
}

func TestParse_InvalidPath(t *testing.T) {
	req := unset.Request{
		Resource: models.Person{
			Name: "John",
		},
		UnsetMask: []string{"invalid"},
	}

	_, err := req.Parse()
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestParse_ArraySupport(t *testing.T) {
	type Wrapper struct {
		People []models.Person `bson:"people"`
	}

	req := unset.Request{
		Resource: Wrapper{
			People: []models.Person{
				{Name: "A", Age: 10},
				{Name: "B", Age: 20},
			},
		},
		UnsetMask: []string{"people.0.age"},
	}

	_, err := req.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// normalize is used to mimic the action mongodb would do on data to get
// the expected behavior as if talking to mongodb.
func normalize(m any) bson.M {
	data, _ := bson.Marshal(m)
	var out bson.M
	_ = bson.Unmarshal(data, &out)
	return out
}

func TestPatch_Person_RemoveZeroValues(t *testing.T) {
	req := unset.Request{
		Resource: models.Person{
			ID:    "123", // Non-zero value
			Name:  "",    // Zero value (will be removed)
			Alive: false, // Zero value (will be removed)
			Age:   0,     // Zero value (will be removed)
		},
	}

	result, err := req.Parse()
	if err != nil {
		t.Fatal(err)
	}

	// Expect only the ID field to remain in $set
	expected := normalize(bson.M{
		"$set": bson.M{
			"_id": "123",
		},
	})

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %+v, got %+v", expected, result)
	}
}
