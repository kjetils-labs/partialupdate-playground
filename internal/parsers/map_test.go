package parsers_test

// type testCaseBuildPaths struct {
// 	TestName            string
// 	TestStruct          any
// 	ExpectedPathsCount  int
// 	ExpectedPaths       []string
// 	ExpectedFailedPaths []string
// }
//
// func TestBuildPaths(t *testing.T) {
// 	testcases := []testCaseBuildPaths{
// 		{
// 			TestName:           "working struct with pointer and sub structs",
// 			TestStruct:         models.Resource{},
// 			ExpectedPathsCount: 23,
// 			ExpectedPaths: []string{
// 				"/id",
// 				"/personResource/id",
// 				"/personResource/name",
// 				"/personResource/alive",
// 				"/personResource/ageInt",
// 				"/personResource/ageInt8",
// 				"/personResource/ageInt16",
// 				"/personResource/ageInt32",
// 				"/personResource/ageInt64",
// 				"/personResource/ageUint",
// 				"/personResource/ageUint8",
// 				"/personResource/ageUint16",
// 				"/personResource/ageUint32",
// 				"/personResource/ageUint64",
// 				"/personResource/ageFloat32",
// 				"/personResource/ageFloat64",
// 				"/personResource/personData/personalNumber",
// 				"/personResource/personData/religious",
// 				"/personResource/likesCheesecake",
// 				"/personResource/tags",
// 			},
// 			ExpectedFailedPaths: []string{},
// 		},
// 		{
// 			TestName:           "Working singular struct",
// 			TestStruct:         models.CheeseCakeLiker{},
// 			ExpectedPathsCount: 1,
// 			ExpectedPaths: []string{
// 				"/likesCheesecake",
// 			},
// 			ExpectedFailedPaths: []string{},
// 		},
// 		{
// 			TestName:           "Bad path",
// 			TestStruct:         models.CheeseCakeLiker{},
// 			ExpectedPathsCount: 1,
// 			ExpectedPaths: []string{
// 				"fork",
// 			},
// 			ExpectedFailedPaths: []string{
// 				"fork",
// 			},
// 		},
// 	}
//
// 	for _, test := range testcases {
// 		_ = t.Run(test.TestName, func(t *testing.T) {
//
// 			fieldPaths := parsers.WalkStruct(test.TestStruct)
// 			if len(fieldPaths) != test.ExpectedPathsCount {
// 				t.Fatalf("received path count %v does not match expected path count %v", len(fieldPaths), test.ExpectedPathsCount)
// 			}
//
// 			invalidPaths := make([]string, 0)
// 			for _, path := range test.ExpectedPaths {
// 				if !fieldInfoContains(fieldPaths, path) {
// 					invalidPaths = append(invalidPaths, path)
// 				}
// 			}
//
// 			ok := reflect.DeepEqual(invalidPaths, test.ExpectedFailedPaths)
// 			if !ok {
// 				t.Fatalf("received paths %v does not match expected paths %v. invalid paths %v", fieldInfoPaths(fieldPaths), test.ExpectedPaths, invalidPaths)
//
// 			}
//
// 		},
// 		)
//
// 	}
// }
//
// func fieldInfoPaths(fieldInfos []parsers.FieldInfo) []string {
// 	paths := make([]string, 0)
// 	for _, fieldInfo := range fieldInfos {
// 		paths = append(paths, fieldInfo.Path)
// 	}
// 	return paths
// }
//
// func fieldInfoContains(fieldInfos []parsers.FieldInfo, path string) bool {
// 	for _, fieldInfo := range fieldInfos {
// 		if fieldInfo.Path == path {
// 			return true
// 		}
// 	}
// 	return false
// }
