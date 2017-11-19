package parallel

import (
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

const numOfProducerCycles = 100
const numOfConsumers = 10

type taskCreatorFunc func(int, chan int) TaskFunc

var rndSrc = rand.NewSource(time.Now().UnixNano())
var random = rand.New(rndSrc)

func TestSuccessfulFlow(t *testing.T) {
	var expectedTotal int
	results := make(chan int, numOfProducerCycles)
	runner := NewBounedRunner(numOfConsumers, true)
	errorsQueue := NewErrorsQueue(1)
	var wg sync.WaitGroup

	// Produce
	wg.Add(1)
	go func() {
		defer wg.Done()
		expectedTotal = produceTasks(runner, results, errorsQueue, createSuccessfulFlowTaskFunc)
	}()

	// Consume
	wg.Add(1)
	go func() {
		defer func() {
			wg.Done()
			close(results)
		}()
		runner.Run()
	}()

	wg.Wait()
	checkResult(expectedTotal, results, t)
}

func TestStopOperationsOnTaskError(t *testing.T) {
	expectedTotal := 1275
	results := make(chan int, numOfProducerCycles)
	runner := NewBounedRunner(numOfConsumers, true)
	errorsQueue := NewErrorsQueue(1)
	var wg sync.WaitGroup

	// Produce
	wg.Add(1)
	go func() {
		defer func() {
			wg.Done()
		}()
		produceTasks(runner, results, errorsQueue, createTaskWithErrorFunc)
	}()

	// Consume
	wg.Add(1)
	go func() {
		defer func() {
			wg.Done()
			close(results)
		}()
		runner.Run()
	}()

	wg.Wait()
	err := errorsQueue.GetError().Error()
	if !strings.Contains(err, "above 50 going to stop") {
		t.Error("Unexpected Error message. Expected: num: 51, above 50 going to stop", "Got:", err)
	}
	checkResult(expectedTotal, results, t)
}

func TestContinueOperationsOnTaskError(t *testing.T) {
	expectedTotal := 1275
	errorsExpectedTotal := 3675
	results := make(chan int, numOfProducerCycles)
	errorsQueue := NewErrorsQueue(100)
	runner := NewBounedRunner(numOfConsumers, false)
	var wg sync.WaitGroup

	// Produce
	wg.Add(1)
	go func() {
		defer func() {
			wg.Done()
		}()
		produceTasks(runner, results, errorsQueue, createTaskWithIntAsErrorFunc)
	}()

	// Consume
	wg.Add(1)
	go func() {
		defer func() {
			wg.Done()
			close(results)
		}()
		runner.Run()
	}()

	wg.Wait()
	checkResult(expectedTotal, results, t)
	checkErrorsResult(errorsExpectedTotal, errorsQueue, t)
}

func TestFailFastOnTaskError(t *testing.T) {
	expectedTotal := 1275
	errorsExpectedTotal := 51
	results := make(chan int, numOfProducerCycles)
	errorsQueue := NewErrorsQueue(100)
	runner := NewBounedRunner(numOfConsumers, true)
	var wg sync.WaitGroup

	// Produce
	wg.Add(1)
	go func() {
		defer func() {
			runner.Done()
			wg.Done()
		}()
		produceTasks(runner, results, errorsQueue, createTaskWithIntAsErrorFunc)
	}()

	// Consume
	wg.Add(1)
	go func() {
		defer func() {
			wg.Done()
			close(results)
		}()
		runner.Run()
	}()

	wg.Wait()
	checkResult(expectedTotal, results, t)

	//TODO: Fix this test
	//This test is fragile since 1 or more goroutines may be executing and failing fast in parallel,
	//calling the error handler and increasing the result. So we cannot use accurate comparison.
	//Here we only take care of uo to 1 additional concurrent failfast.
	errTotal := 0
	for {
		err := errorsQueue.GetError()
		if err == nil {
			break
		}
		x, _ := strconv.Atoi(err.Error())
		errTotal += x
	}
	if errTotal < errorsExpectedTotal || errTotal > 2*errorsExpectedTotal+1 {
		t.Error("Unexpected results err total. Expected:", errorsExpectedTotal, "Got:", errTotal)
	}
}

func checkErrorsResult(errorsExpectedTotal int, errorsQueue *ErrorsQueue, t *testing.T) {
	resultsTotal := 0
	for {
		err := errorsQueue.GetError()
		if err == nil {
			break
		}
		x, _ := strconv.Atoi(err.Error())
		resultsTotal += x
	}
	if resultsTotal != errorsExpectedTotal {
		t.Error("Unexpected results err total. Expected:", errorsExpectedTotal, "Got:", resultsTotal)
	}
}

func checkResult(expectedTotal int, results <-chan int, t *testing.T) {
	var resultsTotal int
	for result := range results {
		resultsTotal += result
	}
	if resultsTotal != expectedTotal {
		t.Error("Unexpected results total. Expected:", expectedTotal, "Got:", resultsTotal)
	}
}

func produceTasks(runner *runner, results chan int, errorsQueue *ErrorsQueue, taskCreator taskCreatorFunc) int {
	defer runner.Close()
	var expectedTotal int
	for i := 0; i < numOfProducerCycles; i++ {
		taskFunc := taskCreator(i, results)
		_, err := runner.AddTaskWithError(taskFunc, errorsQueue.AddErrorNonBlocking)
		if err != nil {
			break
		}
		expectedTotal += i
	}
	fmt.Println("Runner done")
	return expectedTotal
}

func createSuccessfulFlowTaskFunc(num int, result chan int) TaskFunc {
	return func(threadId int) error {
		result <- num
		time.Sleep(time.Millisecond * time.Duration(random.Intn(50)))
		fmt.Printf("[Thread %d] %d\n", threadId, num)
		return nil
	}
}

func createTaskWithErrorFunc(num int, result chan int) TaskFunc {
	return func(threadId int) error {
		if num > 50 {
			return errors.New(fmt.Sprintf("num: %d, above 50 going to stop.", num))
		}
		result <- num
		time.Sleep(time.Millisecond * time.Duration(random.Intn(50)))
		fmt.Printf("[Thread %d] %d\n", threadId, num)
		return nil
	}
}

func createTaskWithIntAsErrorFunc(num int, result chan int) TaskFunc {
	return func(threadId int) error {
		if num > 50 {
			return errors.New(fmt.Sprintf("%d", num))
		}
		result <- num
		time.Sleep(time.Millisecond * time.Duration(random.Intn(50)))
		fmt.Printf("[Thread %d] %d\n", threadId, num)
		return nil
	}
}

type ErrorsQueue struct {
	errorsChan chan (error)
}

func NewErrorsQueue(size int) *ErrorsQueue {
	queueSize := 1
	if size > 1 {
		queueSize = size
	}
	return &ErrorsQueue{errorsChan: make(chan error, queueSize)}
}

func (errQueue *ErrorsQueue) AddErrorNonBlocking(err error) {
	select {
	case errQueue.errorsChan <- err:
	default:
		return
	}
}

func (errQueue *ErrorsQueue) GetError() error {
	select {
	case err := <-errQueue.errorsChan:
		return err
	default:
		return nil
	}
}
