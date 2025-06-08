package entity

type TestCase struct {
	testID         string
	description    string
	expectedOutput string
	testCode       TestCode
	// status als enum still  missing

}

type TestCode struct {
	code string
}
