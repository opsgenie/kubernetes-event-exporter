package batch

import (
	"context"
	"time"
)


// Writer allows to buffer some items and call the Handler function either when the buffer is full or the timeout is
// reached. There will also be support for concurrency for high volume. The handler function is supposed to return an
// array of booleans to indicate whether the transfer was successful or not. It can be replaced with status codes in
// the future to differentiate I/O errors, rate limiting, authorization issues.
type Writer struct {
	cfg      WriterConfig
	Handler  Callback
	buffer   []bufferItem
	len      int
	done     chan bool
	stopDone chan bool
	items    chan interface{}
}

type bufferItem struct {
	v       interface{}
	attempt int
}

type Callback func(ctx context.Context, items []interface{}) []bool

type WriterConfig struct {
	BatchSize  int
	MaxRetries int
	Interval   time.Duration
	Timeout    time.Duration
}

func NewWriter(cfg WriterConfig, cb Callback) *Writer {
	return &Writer{
		cfg:     cfg,
		Handler: cb,
		buffer:  make([]bufferItem, cfg.BatchSize),
	}
}

// Indicates the start to accept the
func (w *Writer) Start() {
	w.done = make(chan bool)
	w.items = make(chan interface{})
	w.stopDone = make(chan bool)
	ticker := time.NewTicker(w.cfg.Interval)

	go func() {
		shouldGoOn := true
		for shouldGoOn {
			select {
			case item := <-w.items:
				if w.len >= w.cfg.BatchSize {
					w.processBuffer(context.Background())
					w.len = 0
				}

				w.buffer[w.len] = bufferItem{v: item, attempt: 0}
				w.len++
			case <-w.done:
				w.processBuffer(context.Background())
				shouldGoOn = false
				w.stopDone <- true
				ticker.Stop()
			case <-ticker.C:
				w.processBuffer(context.Background())
			}
		}
	}()
}

func (w *Writer) processBuffer(ctx context.Context) {
	if w.len == 0 {
		return
	}

	// Need to copy the underlying item to another slice
	slice := make([]interface{}, w.len)
	for i := 0; i < w.len; i++ {
		slice[i] = w.buffer[i].v
	}

	// Call the actual method
	responses := w.Handler(ctx, slice)

	// Overwriting buffer here. Since the newItemsCount will always be equal to or smaller than buffer, it's safe to
	// overwrite the existing items whilst traversing the items.
	var newItemsCount int
	for idx, success := range responses {
		if !success {
			item := w.buffer[idx]
			if item.attempt >= w.cfg.MaxRetries {
				// It's dropped, sorry you asked for it
				continue
			}

			w.buffer[newItemsCount] = bufferItem{
				v:       item.v,
				attempt: item.attempt + 1,
			}

			newItemsCount++
		}
	}

	w.len = newItemsCount
	// TODO(makin) an edge case, if all items fail, and the buffer is full, new item cannot be added to buffer.
}

// Used to signal writer to stop processing items and exit.
func (w *Writer) Stop() {
	w.done <- true
	<-w.stopDone
}

// Submit pushes the items to the income buffer and they are placed onto the actual buffer from there.
func (w *Writer) Submit(items ...interface{}) {
	for _, item := range items {
		w.items <- item
	}
}
