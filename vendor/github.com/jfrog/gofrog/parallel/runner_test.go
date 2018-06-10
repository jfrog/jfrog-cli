package parallel

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"testing"
	"time"
)

var src = rand.NewSource(time.Now().UnixNano())
var rnd = rand.New(src)

func TestTask(t *testing.T) {
	const count = 70
	results := make(chan int, 100)

	runner := NewRunner(4, count, false)
	var expectedTotal int
	var expectedErrorTotal int
	for i := 0; i < count; i++ {
		expectedTotal += i
		if float64(i) > math.Floor(float64(count)/2) {
			expectedErrorTotal += i
		}

		x := i
		runner.AddTask(func(i int) error {
			results <- x
			time.Sleep(time.Millisecond * time.Duration(rnd.Intn(50)))
			fmt.Printf("Run: %d\n", x)
			if float64(x) > math.Floor(float64(count)/2) {
				return errors.New(fmt.Sprintf("Second half value %d not counted", x))
			}
			return nil
		})
	}
	runner.Done()
	runner.Run()

	errs := runner.Errors()

	close(results)
	var resultsTotal int
	for result := range results {
		resultsTotal += result
	}
	if resultsTotal != expectedTotal {
		t.Error("Unexpected results total:", resultsTotal)
	}

	var errorsTotal int
	for k, v := range errs {
		fmt.Printf("err %v: %v\n", k, v)

		if v != nil {
			errorsTotal += k
		}
	}
	if errorsTotal != expectedErrorTotal {
		t.Error("Unexpected errs total:", errorsTotal)
	}
	if errorsTotal == 0 {
		t.Error("Unexpected 0 errs total")
	}

	fmt.Println("expectedTotal=", expectedTotal)
	fmt.Println("expectedErrorTotal=", expectedErrorTotal)
}
