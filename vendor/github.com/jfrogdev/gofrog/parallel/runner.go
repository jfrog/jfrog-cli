package parallel

import (
	"errors"
	"sync"
	"sync/atomic"
)

type Runner interface {
	AddTask(TaskFunc) (int, error)
	AddTaskWithError(TaskFunc, OnErrorFunc) (int, error)
	Run()
	Done()
	Cancel()
	Errors() map[int]error
}

type TaskFunc func(int) error

type OnErrorFunc func(error)

type task struct {
	run     TaskFunc
	onError OnErrorFunc
	num     uint32
}

type taskError struct {
	num int
	err error
}

type runner struct {
	tasks     chan *task
	taskCount uint32

	cancel      chan struct{}
	maxParallel int
	failFast    bool

	errors map[int]error
}

// Create a new capacity runner - a runner we can add tasks to without blocking as long as the capacity is not reached.
// maxParallel - number of go routines for task processing, maxParallel always will be a positive number.
// acceptBeforeBlocking - number of tasks that can be added until a free processing goruntine is needed.
// failFast - is set to true the will stop on first error.
func NewRunner(maxParallel int, capacity uint, failFast bool) *runner {
	consumers := maxParallel
	if consumers < 1 {
		consumers = 1
	}
	if capacity < 1 {
		capacity = 1
	}
	r := &runner{
		tasks:       make(chan *task, capacity),
		cancel:      make(chan struct{}),
		maxParallel: consumers,
		failFast:    failFast,
	}
	r.errors = make(map[int]error)
	return r
}

// Create a new single capacity runner - a runner we can only add tasks to as long as there is a free goroutine in the
// Run() loop to handle it.
// maxParallel - number of go routines for task processing, maxParallel always will be a positive number.
// failFast - is set to true the will stop on first error.
func NewBounedRunner(maxParallel int, failFast bool) *runner {
	return NewRunner(maxParallel, 1, failFast)
}

// Add a task to the producer channel, in case of cancellation event (caused by @Cancel()) will return non nil error.
func (r *runner) AddTask(t TaskFunc) (int, error) {
	return r.addTask(t, nil)
}

// t - the actual task which will be performed by the consumer.
// onError - execute on the returned error while running t
// Return the task number assigned to t. Useful to collect errors from the errors map (see @Errors())
func (r *runner) AddTaskWithError(t TaskFunc, errorHandler OnErrorFunc) (int, error) {
	return r.addTask(t, errorHandler)
}

func (r *runner) addTask(t TaskFunc, errorHandler OnErrorFunc) (int, error) {
	nextCount := atomic.AddUint32(&r.taskCount, 1)
	task := &task{run: t, num: nextCount - 1, onError: errorHandler}

	select {
	case r.tasks <- task:
		return int(task.num), nil
	case <-r.cancel:
		return -1, errors.New("Runner stopped!")
	}
}

// The producer notify that no more task will be produced.
func (r *runner) Close() {
	close(r.tasks)
}

// Run r.maxParallel go routines in order to consume all the tasks
// If a task returns an error and failFast is on all goroutines will stop and a the runner will be notified.
// Notice: Run() is a blocking operation.
func (r *runner) Run() {
	var wg sync.WaitGroup
	var m sync.Mutex
	var once sync.Once
	for i := 0; i < r.maxParallel; i++ {
		wg.Add(1)
		go func(threadId int) {
			defer wg.Done()
			for t := range r.tasks {
				e := t.run(threadId)
				if e != nil {
					if t.onError != nil {
						t.onError(e)
					}
					m.Lock()
					r.errors[int(t.num)] = e
					m.Unlock()
					if r.failFast {
						once.Do(r.Cancel)
						break
					}
				}
			}
		}(i)
	}
	wg.Wait()
}

func (r *runner) Done() {
	select {
	case <-r.cancel:
	//Already canceled
	default:
		close(r.tasks)
	}
}

func (r *runner) Cancel() {
	//Nil the tasks channel if it has buffering to avoid its selection
	if cap(r.tasks) > 1 {
		r.tasks = nil
	}
	close(r.cancel)
}

// Returns a map of errors keyed by the task number
func (r *runner) Errors() map[int]error {
	return r.errors
}
