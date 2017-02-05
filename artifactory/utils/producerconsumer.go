package utils

import (
	"sync"
	"errors"
)

type ErrorHandler func(error)

type Producer interface {
	AddTask(Task) error
	AddTaskWithError(Task, ErrorHandler) error
	Close()
}

type Consumer interface {
	Run()
	StopProducer()
}

type Task func(int) error

type TaskWrapper struct {
	task    Task
	onError ErrorHandler
}

type ProducerConsumer struct {
	tasks          chan TaskWrapper
	cancel         chan struct{}
	numOfConsumers int
	failFast       bool
}

// Create new ProducerConsumer.
// numOfConsumers - number of go routines which do the actually consuming, numOfConsumers always will be a positive number.
// isFailFast - is set to true the will stop on first error.
func NewProducerConsumer(numOfConsumers int, isFailFast bool) *ProducerConsumer {
	consumers := numOfConsumers
	if consumers < 1 {
		consumers = 1
	}
	return &ProducerConsumer{
		tasks: make(chan TaskWrapper),
		cancel: make(chan struct{}),
		numOfConsumers: consumers,
		failFast: isFailFast,
	}
}

// Add a task to the producer channel, in case of cancellation event (caused by @StopProducer()) will return non nil error.
func (pc *ProducerConsumer) AddTask(t Task) error {
	taskWrapper := TaskWrapper{task:t, onError:func(err error) {}}
	return pc.sendNewTask(taskWrapper)
}

// t - the actual task which will be performed by the consumer.
// errorHandler - execute on the returned error while running t
func (pc *ProducerConsumer) AddTaskWithError(t Task, errorHandler ErrorHandler) error {
	taskWrapper := TaskWrapper{task:t, onError:errorHandler}
	return pc.sendNewTask(taskWrapper)
}

func (pc *ProducerConsumer) sendNewTask(taskWrapper TaskWrapper) error {
	select {
	case pc.tasks <- taskWrapper:
		return nil
	case <-pc.cancel:
		return errors.New("Producer stopped")
	}
}

// The producer notify that no more task will be produced.
func (pc *ProducerConsumer) Close() {
	close(pc.tasks)
}

// Run pc.numOfConsumers go routines in order to consume all the tasks in pc.getTasks().
// If a task return an error all the go routines will stop and a the producer will be notified.
// Notice: Consume() is a blocking operation.
func (pc *ProducerConsumer) Run() {
	var wg sync.WaitGroup
	var once sync.Once
	for i := 0; i < pc.numOfConsumers; i++ {
		wg.Add(1)
		go func(threadId int) {
			defer func() {
				wg.Done()
			}()
			for taskWrapper := range pc.getTasks() {
				e := taskWrapper.task(threadId)
				if e != nil {
					taskWrapper.onError(e)
					if pc.failFast {
						once.Do(pc.StopProducer)
						break
					}
				}
			}
		}(i)
	}
	wg.Wait()
}

func (pc *ProducerConsumer) getTasks() <- chan TaskWrapper {
	return pc.tasks
}

func (pc *ProducerConsumer) StopProducer() {
	close(pc.cancel)
}
