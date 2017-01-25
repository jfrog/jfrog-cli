package utils

import (
	"sync"
	"errors"
)

type Producer interface {
	Produce(Task) error
	Finish()
	SetError(error)
	GetError() error
}

type Consumer interface {
	Consume()
	StopProducer()
	SetError(error)
	GetError() error
}

type Task func(int) error

type ProducerConsumer struct {
	tasks            chan Task
	cancel           chan struct{}
	err              error
	canceled         bool
	cancellationLock sync.Mutex
	numOfConsumers   int
}

// Create new ProducerConsumer.
// numOfConsumers - number of go routines which do the actually consuming, numOfConsumers always will be a positive number.
func NewProducerConsumer(numOfConsumers int) *ProducerConsumer {
	consumers := numOfConsumers
	if consumers < 1 {
		consumers = 1
	}
	return &ProducerConsumer{
		tasks: make(chan Task),
		cancel: make(chan struct{}),
		numOfConsumers: consumers,
	}
}

// Add a task to the producer channel, in case of cancellation event (caused by @StopProducer()) will return non nil error.
func (pc *ProducerConsumer) Produce(t Task) error {
	select {
	case pc.tasks <- t:
		return nil
	case <-pc.cancel:
		return errors.New("Producer stopped")
	}
}

// The producer notify that no more task will be produced.
func (pc *ProducerConsumer) Finish() {
	close(pc.tasks)
}

func (pc *ProducerConsumer) SetError(e error) {
	pc.err = e
}

func (pc *ProducerConsumer) GetError() error {
	return pc.err
}

// Run pc.numOfConsumers go routines in order to consume all the tasks in pc.getTasks().
// If a task return an error all the go routines will stop and a the producer will be notified.
// Notice: Consume() is a blocking operation.
func (pc *ProducerConsumer) Consume() {
	var err error
	var wg sync.WaitGroup
	for i := 0; i < pc.numOfConsumers; i++ {
		wg.Add(1)
		go func(threadId int) {
			defer func() {
				wg.Done()
			}()
			for task := range pc.getTasks() {
				if err != nil {
					break
				}
				// Must use e and not err due to synchronization issue
				e := task(threadId)
				if e != nil {
					err = e
					pc.StopProducer()
					break
				}
			}
		}(i)
	}
	wg.Wait()
}

// Notify the producer to stop producing, done by the consumer.
func (pc *ProducerConsumer) StopProducer() {
	if !pc.canceled {
		pc.cancellationLock.Lock()
		defer pc.cancellationLock.Unlock()
		if !pc.canceled {
			close(pc.cancel)
			pc.canceled = true
		}
	}
}

func (pc *ProducerConsumer) getTasks() <- chan Task {
	return pc.tasks
}

