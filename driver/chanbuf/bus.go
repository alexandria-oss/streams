package chanbuf

import (
	"context"
	"log"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/alexandria-oss/streams"
)

// Bus Go channel-backed concurrent-safe messaging bus which implements fan-out messaging pattern.
//
// Zero-value IS NOT ready to use, please call NewBus routine instead.
type Bus struct {
	msgQueue             chan streams.Message
	subscribeLock        sync.Mutex
	readerReg            sync.Map
	readerHandlerTimeout time.Duration

	isReady        atomic.Uint32
	isReadyWg      *sync.WaitGroup
	isShutdown     atomic.Uint32
	inFlightProcWg *sync.WaitGroup
	baseCtx        context.Context
	baseCtxCancel  context.CancelFunc
	logger         *log.Logger
}

// Config Bus configuration parameters.
type Config struct {
	// Internal queue buffer capacity. Bus will allocate an unbuffered channel if <= 0.
	QueueBufferFactor int
	// streams.ReaderHandleFunc maximum process execution duration.
	ReaderHandlerTimeout time.Duration
	// logging instance used by internal processes.
	Logger *log.Logger
}

var (
	// DefaultBus global Bus instance used by both Writer and Reader if passed Bus instance was nil.
	DefaultBus *Bus
	// DefaultBusConfig global Bus instance configuration used by both Writer and Reader if passed Bus instance was nil.
	DefaultBusConfig Config
	defaultBusOnce   = &sync.Once{}
)

// NewBus allocates a new Bus.
func NewBus(cfg Config) *Bus {
	var msgQueue chan streams.Message
	if cfg.QueueBufferFactor > 0 {
		msgQueue = make(chan streams.Message, cfg.QueueBufferFactor)
	} else {
		msgQueue = make(chan streams.Message)
	}

	logger := cfg.Logger
	if logger == nil {
		logger = log.New(os.Stdout, "streams.chanbuf: ", 0)
	}

	readyWg := &sync.WaitGroup{}
	readyWg.Add(1)
	return &Bus{
		msgQueue:             msgQueue,
		subscribeLock:        sync.Mutex{},
		readerReg:            sync.Map{},
		readerHandlerTimeout: cfg.ReaderHandlerTimeout,
		isReady:              atomic.Uint32{},
		isReadyWg:            readyWg,
		inFlightProcWg:       &sync.WaitGroup{},
		logger:               logger,
	}
}

func newDefaultBus() {
	defaultBusOnce.Do(func() {
		DefaultBus = NewBus(DefaultBusConfig)
	})
}

// Start spins up the Bus readers, blocking the I/O until Bus.Shutdown is called.
func (b *Bus) Start() {
	b.baseCtx, b.baseCtxCancel = context.WithCancel(context.Background())
	b.isReady.Store(1)
	b.isReadyWg.Done()
	for msg := range b.msgQueue {
		// fan-out process

		// We implement a root-child lock mechanism.
		// For each message send, we assume at least one process (root proc.) will run at the worker (for range statement in Start).
		// This root process could potentially have N childs where N >= 0. If N = 0, then we just dispose the root lock.
		// Nevertheless, if N >= 1, then besides disposing root lock, we add N as delta to the lock (child locks) to later
		// dispose the root lock as we don't need it anymore (in-flight lock would already have its delta value >= 1).
		//
		// This mechanism guarantees every reader process is waited to finish by other internal mechanisms such as Shutdown.
		subsAny, ok := b.readerReg.Load(msg.StreamName)
		if !ok {
			b.logger.Printf("stream <%s> has no subscribers, skipping", msg.StreamName)
			b.inFlightProcWg.Done() // dispose message root lock
			continue
		}

		subs := subsAny.([]streams.ReaderHandleFunc)
		b.inFlightProcWg.Add(len(subs)) // add child locks
		b.inFlightProcWg.Done()         // dispose message root lock
		for _, sub := range subs {
			go func(handler streams.ReaderHandleFunc, msgCopy streams.Message) {
				defer b.inFlightProcWg.Done()
				timeoutCtx, cancel := context.WithTimeout(b.baseCtx, b.readerHandlerTimeout)
				defer cancel()
				if err := handler(timeoutCtx, msgCopy); err != nil {
					b.logger.Printf("stream <%s> handler failed, err: %s", msgCopy.StreamName,
						err.Error())
				}
			}(sub, msg)
		}
	}
}

// Shutdown disposes allocated resources and signals internal processes for a graceful shutdown.
//
// This operation might block I/O as it waits for in-flight reader processes to finish -either by success or timeout-.
func (b *Bus) Shutdown() {
	if b.isShutdown.Load() == 1 || b.isReady.Load() == 0 {
		return
	}

	b.isShutdown.Store(1)
	b.inFlightProcWg.Wait() // wait for in-flight process to finish
	b.baseCtxCancel()
	close(b.msgQueue)
	b.msgQueue = nil
}

// Publish sends a message to one or many readers subscribed to msg.StreamName (aka. fan-out) in a fire-and-forget
// way meaning the internal writer only cares about writing to the queue whereas subscribers are independent processes.
//
// This routine will fail if Bus is shut down.
func (b *Bus) Publish(msg streams.Message) error {
	if b.isShutdown.Load() == 1 {
		return streams.ErrBusIsShutdown
	} else if isEmptyMsg := msg.StreamName == "" || len(msg.Data) == 0; isEmptyMsg {
		return streams.ErrEmptyMessage
	}

	b.isReadyWg.Wait()
	// We implement a root-child lock mechanism.
	// For each message send, we assume at least one process (root proc.) will run at the worker (for range statement in Start).
	// This root process could potentially have N childs where N >= 0. If N = 0, then we just dispose the root lock.
	// Nevertheless, if N >= 1, then besides disposing root lock, we add N as delta to the lock (child locks) to later
	// perform the actual root lock disposal as we don't need it anymore (in-flight lock would already have its delta value >= 1).
	//
	// This mechanism guarantees every reader process is waited to finish by other internal mechanisms such as Shutdown.
	b.inFlightProcWg.Add(1)
	b.msgQueue <- msg
	return nil
}

// Subscribe appends a reader handler to a stream.
func (b *Bus) Subscribe(stream string, handler streams.ReaderHandleFunc) {
	b.subscribeLock.Lock()
	defer b.subscribeLock.Unlock()

	subscribers := []streams.ReaderHandleFunc{handler}
	subsAny, ok := b.readerReg.Load(stream)
	if ok {
		subscribers = append(subscribers, subsAny.([]streams.ReaderHandleFunc)...)
	}

	b.readerReg.Store(stream, subscribers)
}

// Start spins up the Bus readers, blocking the I/O until Bus.Shutdown is called.
func Start() {
	DefaultBus.Start()
}

// Shutdown disposes allocated resources and signals internal processes for a graceful shutdown.
//
// This operation might block I/O as it waits for in-flight reader processes to finish -either by success or timeout-.
func Shutdown() {
	DefaultBus.Shutdown()
}

// Publish sends a message to one or many readers subscribed to msg.StreamName (aka. fan-out) in a fire-and-forget
// way meaning the internal writer only cares about writing to the queue whereas subscribers are independent processes.
//
// This routine will fail if Bus is shut down.
func Publish(msg streams.Message) error {
	return DefaultBus.Publish(msg)
}

// Subscribe appends a reader handler to a stream.
func Subscribe(stream string, handler streams.ReaderHandleFunc) {
	DefaultBus.Subscribe(stream, handler)
}
