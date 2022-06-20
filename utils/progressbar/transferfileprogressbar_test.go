package progressbar

func MockProgressInitialization() func() {
	originFunc := ShouldInitProgressBar
	ShouldInitProgressBar = func() (bool, error) { return true, nil }
	return func() {
		ShouldInitProgressBar = originFunc
	}
}

//func Test(t *testing.T) {
//	callback := MockProgressInitialization()
//	defer callback()
//	ActualTestProgressbar()
//
//}
