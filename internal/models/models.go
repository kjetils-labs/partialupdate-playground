package models

type Person struct {
	ID    string `json:"id,omitempty" bson:"_id,omitempty"` // we store the id as the document _id
	Name  string `json:"name,omitempty" bson:"name,omitempty"`
	Alive bool   `json:"alive,omitempty" bson:"alive,omitempty"`
	Age   int    `json:"age,omitempty" bson:"age,omitempty"`
}
type PersonPatch struct {
	ID    *string `json:"id,omitempty" bson:"_id,omitempty"` // we store the id as the document _id
	Name  *string `json:"name,omitempty" bson:"name,omitempty"`
	Alive *bool   `json:"alive,omitempty" bson:"alive,omitempty"`
	Age   *int    `json:"age,omitempty" bson:"age,omitempty"`
}
