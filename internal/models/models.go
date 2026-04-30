package models

type Resource struct {
	ID             string  `json:"id,omitempty" bson:"_id,omitempty"`
	CarResource    Car     `json:"carResource" bson:"carResource,omitempty"`
	PersonResource *Person `json:"personResource,omitempty" bson:"personResource,omitempty"`
}

type Car struct {
	ID    string `json:"id,omitempty" bson:"id,omitempty"`
	Make  string `json:"make,omitempty" bson:"make,omitempty"`
	Model string `json:"model,omitempty" bson:"model,omitempty"`
}

type Person struct {
	ID              string             `json:"id,omitempty" bson:"id,omitempty"`
	Name            string             `json:"name,omitempty" bson:"name,omitempty"`
	Tags            map[string]string  `json:"tags,omitempty" bson:"tags,omitempty"`
	PointerTags     map[string]*string `json:"pointerTags,omitempty" bson:"pointerTags,omitempty"`
	Slice           []string           `json:"slice,omitempty" bson:"slice,omitempty"`
	PointerSlice    []*string          `json:"pointerSlice,omitempty" bson:"pointerSlice,omitempty"`
	Alive           bool               `json:"alive,omitempty" bson:"alive,omitempty"`
	AgeInt          int                `json:"ageInt,omitempty" bson:"ageInt,omitempty"`
	AgeInt8         int8               `json:"ageInt8,omitempty" bson:"ageInt8,omitempty"`
	AgeInt16        int16              `json:"ageInt16,omitempty" bson:"ageInt16,omitempty"`
	AgeInt32        int32              `json:"ageInt32,omitempty" bson:"ageInt32,omitempty"`
	AgeInt64        int64              `json:"ageInt64,omitempty" bson:"ageInt64,omitempty"`
	AgeUint         uint               `json:"ageUint,omitempty" bson:"ageUint,omitempty"`
	AgeUint8        uint8              `json:"ageUint8,omitempty" bson:"ageUint8,omitempty"`
	AgeUint16       uint16             `json:"ageUint16,omitempty" bson:"ageUint16,omitempty"`
	AgeUint32       uint32             `json:"ageUint32,omitempty" bson:"ageUint32,omitempty"`
	AgeUint64       uint64             `json:"ageUint64,omitempty" bson:"ageUint64,omitempty"`
	AgeFloat32      float32            `json:"ageFloat32,omitempty" bson:"ageFloat32,omitempty"`
	AgeFloat64      float64            `json:"ageFloat64,omitempty" bson:"ageFloat64,omitempty"`
	PersonData      PersonData         `json:"personData" bson:"personData,omitempty"`
	CheeseCakeLiker `json:",inline" bson:"inline"`
}

type PersonData struct {
	PersonalNumber int  `json:"personalNumber,omitempty" bson:"personalNumber,omitempty"`
	Religious      bool `json:"religious,omitempty" bson:"religious,omitempty"`
}

type CheeseCakeLiker struct {
	LikesCheesecake bool `json:"likesCheesecake,omitempty" bson:"likesCheesecake,omitempty"`
}
