package models

type Resource struct {
	ID             string  `json:"id,omitempty" bson:"_id,omitempty"`
	PersonResource *Person `json:"personResource,omitempty" bson:"personResource,omitempty"`
}

type Person struct {
	ID              string     `json:"id,omitempty" bson:"id,omitempty"`
	Name            string     `json:"name,omitempty" bson:"name,omitempty"`
	Alive           bool       `json:"alive,omitempty" bson:"alive,omitempty"`
	Age             int        `json:"age,omitempty" bson:"age,omitempty"`
	PersonData      PersonData `json:"personData" bson:"personData,omitempty"`
	CheeseCakeLiker `json:",inline" bson:"inline"`
}

type PersonData struct {
	PersonalNumber int  `json:"personalNumber,omitempty" bson:"personalNumber,omitempty"`
	Religious      bool `json:"religious,omitempty" bson:"religious,omitempty"`
}

type CheeseCakeLiker struct {
	LikesCheesecake bool `json:"likesCheesecake,omitempty" bson:"likesCheesecake,omitempty"`
}
