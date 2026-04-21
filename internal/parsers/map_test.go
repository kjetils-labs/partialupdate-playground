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
			TestName:           "working struct with imbeds and sub structs",
			TestStruct:         models.Person{},
			ExpectedPathsCount: 7,
			ExpectedPaths: []string{
				"ID",
				"Name",
				"Alive",
				"Age",
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

	// Tests a bad retrieval works as expected.
	_ = parsers.FieldMaps.Get("Fakestruct")

	for _, test := range testcases {
		_ = t.Run(test.TestName, func(t *testing.T) {

			structName := reflect.TypeOf(test.TestStruct).Name()
			err := parsers.FieldMaps.Add(test.TestStruct)
			if err != nil {
				t.Fatalf("failed to parse valid struct. %v", err)
			}

			fieldmap := parsers.FieldMaps.Get(structName)
			if fieldmap.Length() != test.ExpectedPathsCount {
				t.Fatalf("received field count %v does not match expected field count %v", fieldmap.Length(), test.ExpectedPathsCount)
			}

			invalidPaths := make([]string, 0)
			for _, path := range test.ExpectedPaths {
				field := fieldmap.Validate(path)
				if field == nil {
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
