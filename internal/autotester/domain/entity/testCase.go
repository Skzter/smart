package entity

type TestCase struct {
	TestID      string
	Description string
	TestCode    TestCode
	Status      TestStatus // implemented as enumeration
}

type TestCode struct {
	Code string
}
