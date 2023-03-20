package streams

import (
	"context"
	"sync"

	"github.com/hashicorp/go-multierror"
)

// A SubscriberScheduler is a high-level component used to manage and schedule Reader tasks.
//
// Zero value is NOT ready to use.
type SubscriberScheduler struct {
	reader          Reader
	eventReg        EventRegistry
	reg             []*ReadTask
	baseCtx         context.Context
	baseCtxCancel   context.CancelFunc
	inFlightWorkers sync.WaitGroup
	isShutdown      bool
}

// NewSubscriberScheduler allocates a new SubscriberScheduler instance ready to be used.
// Returns a ReadTask instance ready to be chained to another builder routine (Fluent API-like).
func NewSubscriberScheduler(r Reader, eventReg EventRegistry) SubscriberScheduler {
	return SubscriberScheduler{
		reader:          r,
		eventReg:        eventReg,
		reg:             make([]*ReadTask, 0),
		baseCtx:         nil,
		baseCtxCancel:   nil,
		inFlightWorkers: sync.WaitGroup{},
	}
}

// SubscribeTopic registers a stream reading job to a specific topic.
// Returns a ReadTask instance ready to be chained to another builder routine (Fluent API-like).
func (r *SubscriberScheduler) SubscribeTopic(topic string, handler ReaderHandleFunc) *ReadTask {
	task := &ReadTask{
		Stream:  topic,
		Handler: handler,
	}
	r.reg = append(r.reg, task)
	return task
}

// Subscribe registers a stream reading job using Event registered topic from EventRegistry.
// This routine will panic if Event was not previously registered.
//
// Returns a ReadTask instance ready to be chained to another builder routine (Fluent API-like).
func (r *SubscriberScheduler) Subscribe(event Event, handler ReaderHandleFunc) *ReadTask {
	task, err := r.SubscribeSafe(event, handler)
	if err != nil {
		panic(err)
	}
	return task
}

// SubscribeSafe registers a stream reading job using Event registered topic from EventRegistry.
// Returns ErrEventNotFound if Event was not previously registered.
func (r *SubscriberScheduler) SubscribeSafe(event Event, handler ReaderHandleFunc) (*ReadTask, error) {
	topic, err := r.eventReg.GetEventTopic(event)
	if err != nil {
		return nil, err
	}

	task := &ReadTask{
		Stream:  topic,
		Handler: handler,
	}
	r.reg = append(r.reg, task)
	return task, nil
}

// Start schedules and spins up a worker for each registered ReadTask(s).
func (r *SubscriberScheduler) Start() error {
	if len(r.reg) == 0 {
		return ErrNoSubscriberRegistered
	} else if r.isShutdown {
		return ErrBusIsShutdown
	}

	errs := &multierror.Error{}
	errMu := sync.Mutex{}
	r.baseCtx, r.baseCtxCancel = context.WithCancel(context.Background())
	r.inFlightWorkers.Add(len(r.reg))
	for _, readerTask := range r.reg {
		go func(task ReadTask) {
			defer r.inFlightWorkers.Done()
			if errRead := r.reader.Read(r.baseCtx, task); errRead != nil {
				errMu.Lock()
				errs = multierror.Append(errRead, errs)
				errMu.Unlock()
			}
		}(*readerTask)
	}
	return errs.ErrorOrNil()
}

// Shutdown triggers graceful shutdown of running ReadTask(s) worker(s). This routine will block I/O until
// all workers have been properly shutdown.
func (r *SubscriberScheduler) Shutdown() error {
	r.isShutdown = true
	r.baseCtxCancel()
	r.inFlightWorkers.Wait()
	return nil
}
