package database

import (
	"context"
	"errors"
	"fmt"
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
		return nil, fmt.Errorf("failed to connect to mongodb at address %v. %w", uri, err)
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
func (s *PersonStore) Create(p *models.Resource) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := s.coll.InsertOne(ctx, &p)
	if mongo.IsDuplicateKeyError(err) {
		return errors.New("person with this ID already exists")
	}
	return err
}

// ---------- Get ----------
func (s *PersonStore) Get(id string) (models.Resource, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var p models.Resource
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
func (s *PersonStore) List() ([]models.Resource, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cur, err := s.coll.Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var out []models.Resource
	for cur.Next(ctx) {
		var p models.Resource
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
func (s *PersonStore) Update(id string, p models.Resource) (bool, error) {
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

// // ---------- Patch ----------
// func (s *PersonStore) Patch(id string, p jsonpatch.Operation) (bool, error) {
//
// 	update, err := p.Parse()
// 	if err != nil {
// 		return false, fmt.Errorf("failed to parse update. %w", err)
// 	}
//
// 	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
// 	defer cancel()
// 	slog.Info("data", "update", update)
// 	filter := bson.M{"_id": id}
// 	res, err := s.coll.UpdateOne(ctx,
// 		filter,
// 		update,
// 	)
// 	if err != nil {
// 		return false, fmt.Errorf("failed to update record. %w", err)
// 	}
// 	return res.MatchedCount > 0, nil
// }

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
