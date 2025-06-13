package entity

type TestCase struct {
	TestID         string
	Description    string
	ExpectedOutput string
	TestCode       TestCode
	Status         TestStatus // implemented as enum
}

type TestCode struct {
	Code string
}
