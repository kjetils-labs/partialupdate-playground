package parsers_test

import (
	"partialupdate/internal/models"
	"partialupdate/internal/parsers"
	"reflect"
	"testing"
)

type testCaseBuildPaths struct {
	TestName            string
	TestStruct          any
	ExpectedPathsCount  int
	ExpectedPaths       []string
	ExpectedFailedPaths []string
}

func TestBuildPaths(t *testing.T) {
	testcases := []testCaseBuildPaths{
		{
			TestName:           "working struct with pointer and sub structs",
			TestStruct:         models.Resource{},
			ExpectedPathsCount: 20,
			ExpectedPaths: []string{
				"ID",
				"PersonResource.ID",
				"PersonResource.Name",
				"PersonResource.Alive",
				"PersonResource.AgeInt",
				"PersonResource.AgeInt8",
				"PersonResource.AgeInt16",
				"PersonResource.AgeInt32",
				"PersonResource.AgeInt64",
				"PersonResource.AgeUint",
				"PersonResource.AgeUint8",
				"PersonResource.AgeUint16",
				"PersonResource.AgeUint32",
				"PersonResource.AgeUint64",
				"PersonResource.AgeFloat32",
				"PersonResource.AgeFloat64",
				"PersonResource.PersonData.PersonalNumber",
				"PersonResource.PersonData.Religious",
				"PersonResource.CheeseCakeLiker.LikesCheesecake",
				"PersonResource.Tags",
			},
			ExpectedFailedPaths: []string{},
		},
		{
			TestName:           "working struct with imbeds and sub structs",
			TestStruct:         models.Person{},
			ExpectedPathsCount: 19,
			ExpectedPaths: []string{
				"ID",
				"Name",
				"Alive",
				"AgeInt",
				"AgeInt8",
				"AgeInt16",
				"AgeInt32",
				"AgeInt64",
				"AgeUint",
				"AgeUint8",
				"AgeUint16",
				"AgeUint32",
				"AgeUint64",
				"AgeFloat32",
				"AgeFloat64",
				"PersonData.PersonalNumber",
				"PersonData.Religious",
				"CheeseCakeLiker.LikesCheesecake",
			},
			ExpectedFailedPaths: []string{},
		},
		{
			TestName:           "Working singular struct",
			TestStruct:         models.CheeseCakeLiker{},
			ExpectedPathsCount: 1,
			ExpectedPaths: []string{
				"LikesCheesecake",
			},
			ExpectedFailedPaths: []string{},
		},
		{
			TestName:           "Bad path",
			TestStruct:         models.CheeseCakeLiker{},
			ExpectedPathsCount: 1,
			ExpectedPaths: []string{
				"fork",
			},
			ExpectedFailedPaths: []string{
				"fork",
			},
		},
	}

	for _, test := range testcases {
		_ = t.Run(test.TestName, func(t *testing.T) {

			paths := parsers.WalkStruct(test.TestStruct)
			if len(paths) != test.ExpectedPathsCount {
				t.Fatalf("received path count %v does not match expected path count %v", len(paths), test.ExpectedPathsCount)
			}

			invalidPaths := make([]string, 0)
			for _, path := range test.ExpectedPaths {
				for _, p := range paths {
					if p.Path == path {
						continue
					}
					invalidPaths = append(invalidPaths, path)
				}

			}

			ok := reflect.DeepEqual(invalidPaths, test.ExpectedFailedPaths)
			if !ok {
				t.Fatalf("received paths %v does not match expected paths %v", invalidPaths, test.ExpectedPaths)
			}

		},
		)

	}
}
