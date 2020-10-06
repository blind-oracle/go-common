package batcher

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/blind-oracle/go-common/logger"
)

// FlushFunc is a batch flush function signature
type FlushFunc func([]interface{}) error

// Config is a batcher config
type Config struct {
	BufferSize int
	Flush      FlushFunc

	BatchSize     int
	FlushInterval time.Duration

	Logger logger.Logger
}

// Stats is batcher stats
type Stats struct {
	Buffered    uint64
	Sent        uint64
	Dropped     uint64
	FlushFailed uint64
}

// Batcher is a batching dispatcher
type Batcher struct {
	cfg Config

	batch []interface{}
	count int

	chanIn    chan interface{}
	chanClose chan struct{}

	stats Stats
	dErr  error

	wg sync.WaitGroup
	sync.Mutex
	logger.Logger
}

// New creates a Batcher
func New(c Config) (b *Batcher, err error) {
	if c.Flush == nil {
		return nil, fmt.Errorf("Flush function required")
	}

	if c.BufferSize == 0 {
		c.BufferSize = 10000
	}

	if c.BatchSize == 0 {
		c.BatchSize = 100
	}

	if c.FlushInterval == 0 {
		c.FlushInterval = time.Second
	}

	if c.Logger == nil {
		c.Logger = logger.NewSimpleLogger("batcher")
	}

	b = &Batcher{
		cfg: c,

		batch:     make([]interface{}, c.BatchSize),
		chanIn:    make(chan interface{}, c.BufferSize),
		chanClose: make(chan struct{}),

		Logger: c.Logger,
	}

	b.wg.Add(2)
	go b.dispatch()
	go b.periodicFlush()

	return
}

// Queue an object in a buffer
func (b *Batcher) Queue(o interface{}) bool {
	select {
	case b.chanIn <- o:
		atomic.AddUint64(&b.stats.Buffered, 1)
		return true

	default:
		atomic.AddUint64(&b.stats.Dropped, 1)
		return false
	}
}

func (b *Batcher) push(o interface{}) (err error) {
	b.Lock()
	b.batch[b.count] = o

	if b.count++; b.count >= b.cfg.BatchSize {
		err = b.flush()
	}

	b.Unlock()
	return
}

func (b *Batcher) flush() (err error) {
	if err = b.cfg.Flush(b.batch[:b.count]); err != nil {
		atomic.AddUint64(&b.stats.FlushFailed, 1)
		return
	}

	atomic.AddUint64(&b.stats.Sent, uint64(b.count))
	b.count = 0
	return
}

func (b *Batcher) tryFlush() error {
	b.Lock()
	defer b.Unlock()

	if b.count == 0 {
		return nil
	}

	return b.flush()
}

func (b *Batcher) dispatch() {
	defer b.wg.Done()

	for {
		select {
		case <-b.chanClose:
			// Drain channel
			b.Infof("Draining buffer (%d events)", len(b.chanIn))

			for {
				select {
				case o := <-b.chanIn:
					if b.dErr = b.push(o); b.dErr != nil {
						b.Errorf("Unable to flush: %s", b.dErr)
						return
					}

				default:
					b.Infof("Buffer drained, flushing")

					if b.dErr = b.tryFlush(); b.dErr != nil {
						b.Errorf("Unable to flush: %s", b.dErr)
					} else {
						b.Infof("Buffer flushed")
					}

					return
				}
			}

		case o := <-b.chanIn:
			if b.dErr = b.push(o); b.dErr == nil {
				break
			}

			b.Errorf("Unable to flush batch: %s", b.dErr)

			// Try to requeue if there's space
			select {
			case b.chanIn <- o:
			default:
			}

			return
		}
	}
}

// Stats returns batcher's stats
func (b *Batcher) Stats() Stats {
	return b.stats
}

func (b *Batcher) periodicFlush() {
	tick := time.NewTicker(b.cfg.FlushInterval)

	defer func() {
		tick.Stop()
		b.wg.Done()
	}()

	for {
		select {
		case <-tick.C:
			b.tryFlush()
		case <-b.chanClose:
			return
		}
	}
}

// Close flushes the buffers and stops Batcher
func (b *Batcher) Close() error {
	close(b.chanClose)
	b.wg.Wait()
	return b.dErr
}
