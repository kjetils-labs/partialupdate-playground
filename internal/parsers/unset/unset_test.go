package unset_test

import (
	"partialupdate/internal/models"
	"partialupdate/internal/parsers/unset"
	"reflect"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

// normalize ensures BSON-consistent comparison
func normalize(v any) bson.M {
	data, err := bson.Marshal(v)
	if err != nil {
		panic(err)
	}
	var out bson.M
	if err := bson.Unmarshal(data, &out); err != nil {
		panic(err)
	}
	return out
}

func TestParsePersonRemoveZeroValuesNested(t *testing.T) {
	req := unset.Request{
		Resource: &models.Resource{
			PersonResource: &models.Person{
				ID:    "123",
				Name:  "",
				Alive: false,
				Age:   0,
				PersonData: models.PersonData{
					PersonalNumber: 0,
					Religious:      false,
				},
				CheeseCakeLiker: models.CheeseCakeLiker{
					LikesCheesecake: false,
				},
			},
		},
	}

	result, err := req.Parse()
	if err != nil {
		t.Fatal(err)
	}

	expected := bson.M{
		"$set": bson.M{
			"personResource": bson.M{
				"_id": "123",
			},
		},
	}

	if !reflect.DeepEqual(normalize(result), normalize(expected)) {
		t.Errorf("expected %+v, got %+v", normalize(expected), normalize(result))
	}
}

func TestParsePersonUnsetMaskNested(t *testing.T) {
	req := unset.Request{
		Resource: &models.Resource{
			PersonResource: &models.Person{
				ID:    "123",
				Name:  "John",
				Alive: true,
				Age:   30,
				PersonData: models.PersonData{
					PersonalNumber: 12345,
					Religious:      true,
				},
				CheeseCakeLiker: models.CheeseCakeLiker{
					LikesCheesecake: true,
				},
			},
		},
		UnsetMask: []string{
			"personResource.name",
			"personResource.personData.religious",
			"personResource.likesCheesecake",
		},
	}

	result, err := req.Parse()
	if err != nil {
		t.Fatal(err)
	}

	expected := bson.M{
		"$set": bson.M{
			"personResource": bson.M{
				"_id":   "123",
				"alive": true,
				"age":   30,
				"personData": bson.M{
					"personalNumber": 12345,
				},
			},
		},
		"$unset": bson.M{
			"personResource.name":                 "",
			"personResource.personData.religious": "",
			"personResource.likesCheesecake":      "",
		},
	}

	if !reflect.DeepEqual(normalize(result), normalize(expected)) {
		t.Errorf("expected %+v, got %+v", normalize(expected), normalize(result))
	}
}

func TestParsePersonPartialUpdateNested(t *testing.T) {
	req := unset.Request{
		Resource: &models.Resource{
			PersonResource: &models.Person{
				ID: "123",
				PersonData: models.PersonData{
					PersonalNumber: 67890,
				},
			},
		},
	}

	result, err := req.Parse()
	if err != nil {
		t.Fatal(err)
	}

	expected := bson.M{
		"$set": bson.M{
			"personResource": bson.M{
				"_id": "123",
				"personData": bson.M{
					"personalNumber": 67890,
				},
			},
		},
	}

	if !reflect.DeepEqual(normalize(result), normalize(expected)) {
		t.Errorf("expected %+v, got %+v", normalize(expected), normalize(result))
	}
}
