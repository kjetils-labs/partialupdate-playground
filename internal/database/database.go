package database

import (
	"context"
	"errors"
	"log/slog"
	"partialupdate/internal/models"
	"reflect"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ---------------------------------------------------------------------
// MongoDB Store (implements the same methods as the old in‑memory store)
// ---------------------------------------------------------------------

type PersonStore struct {
	coll *mongo.Collection // the collection where Person documents live
}

// NewMongoStore creates a store that talks to MongoDB.
//
//	uri        – e.g. "mongodb://localhost:27017"
//	dbName     – database name, e.g. "demo"
//	collName   – collection name, e.g. "people"
func NewMongoStore(uri, dbName, collName string) (*PersonStore, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOpts := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, err
	}

	// Ping to verify the connection works
	if err = client.Ping(ctx, nil); err != nil {
		return nil, err
	}
	slog.Info("verified connection to db")

	coll := client.Database(dbName).Collection(collName)

	return &PersonStore{coll: coll}, nil
}

// ---------- Create ----------
func (s *PersonStore) Create(p *models.Person) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := s.coll.InsertOne(ctx, &p)
	if mongo.IsDuplicateKeyError(err) {
		return errors.New("person with this ID already exists")
	}
	return err
}

// ---------- Get ----------
func (s *PersonStore) Get(id string) (models.Person, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var p models.Person
	err := s.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&p)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return p, false, nil
		}
		return p, false, err
	}
	return p, true, nil
}

// ---------- List ----------
func (s *PersonStore) List() ([]models.Person, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cur, err := s.coll.Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var out []models.Person
	for cur.Next(ctx) {
		var p models.Person
		if err = cur.Decode(&p); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	if err = cur.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// ---------- Update (full replace) ----------
func (s *PersonStore) Update(id string, p models.Person) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// Ensure the document's _id stays the same regardless of what the caller sent.
	p.ID = id
	res, err := s.coll.ReplaceOne(ctx, bson.M{"_id": id}, p)
	if err != nil {
		return false, err
	}
	// MatchedCount == 0 → nothing existed with that id
	return res.MatchedCount > 0, nil
}

// ---------- Patch ----------
func (s *PersonStore) Patch(id string, p models.PersonPatch) (bool, error) {

	// Build the $set document: {$set: {field1: val1, field2: val2, …}}
	updates := bson.D{}

	// get the type of struct == Book
	typeData := reflect.TypeOf(p)

	// get the values from the provided object: author -> Paulo Coelho
	values := reflect.ValueOf(p)

	for i := 1; i < typeData.NumField(); i++ {
		field := typeData.Field(i) // get the field from the struct definition
		val := values.Field(i)     // get the value from the specified field position
		tag := field.Tag.Get("bson")

		if isZeroType(val) {
			continue
		}
		update := bson.E{
			Key:   tag,
			Value: val.Interface(),
		}
		updates = append(updates, update)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	res, err := s.coll.UpdateOne(ctx,
		bson.M{"_id": id},
		bson.M{"$set": updates},
	)
	if err != nil {
		return false, err
	}
	return res.MatchedCount > 0, nil
}

// ---------- Delete ----------
func (s *PersonStore) Delete(id string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	res, err := s.coll.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return false, err
	}
	return res.DeletedCount > 0, nil
}

// isZeroType checks if the value from the struct is the zero value of its type
func isZeroType(value reflect.Value) bool {
	zero := reflect.Zero(value.Type()).Interface()

	switch value.Kind() {
	case reflect.Slice, reflect.Array, reflect.Chan, reflect.Map:
		return value.Len() == 0
	default:
		return reflect.DeepEqual(zero, value.Interface())
	}
}
