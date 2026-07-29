package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
)

var ErrWriterClosed = errors.New("database writer is closed")

type WriteFunc func(context.Context, *sql.Tx) error

type writeRequest struct {
	ctx    context.Context
	write  WriteFunc
	result chan error
}

// Writer serializes mutation transactions while the database permits bounded
// concurrent readers. Start exactly one Writer for each DB owner process.
type Writer struct {
	db     *sql.DB
	queue  chan writeRequest
	done   chan struct{}
	mu     sync.RWMutex
	closed bool
}

func NewWriter(db *sql.DB, queueSize int) *Writer {
	if queueSize < 1 {
		queueSize = 1
	}
	w := &Writer{db: db, queue: make(chan writeRequest, queueSize), done: make(chan struct{})}
	go w.run()
	return w
}

func (w *Writer) Submit(ctx context.Context, write WriteFunc) error {
	if write == nil {
		return errors.New("write function is required")
	}
	w.mu.RLock()
	if w.closed {
		w.mu.RUnlock()
		return ErrWriterClosed
	}
	request := writeRequest{ctx: ctx, write: write, result: make(chan error, 1)}
	select {
	case w.queue <- request:
		w.mu.RUnlock()
	case <-ctx.Done():
		w.mu.RUnlock()
		return ctx.Err()
	}
	select {
	case err := <-request.result:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (w *Writer) Close() {
	w.mu.Lock()
	if !w.closed {
		w.closed = true
		close(w.queue)
	}
	w.mu.Unlock()
	<-w.done
}

func (w *Writer) run() {
	defer close(w.done)
	for request := range w.queue {
		if err := request.ctx.Err(); err != nil {
			request.result <- err
			continue
		}
		tx, err := w.db.BeginTx(request.ctx, nil)
		if err == nil {
			err = request.write(request.ctx, tx)
		}
		if err != nil {
			if tx != nil {
				_ = tx.Rollback()
			}
			request.result <- fmt.Errorf("database write: %w", err)
			continue
		}
		if err = tx.Commit(); err != nil {
			request.result <- fmt.Errorf("commit database write: %w", err)
			continue
		}
		request.result <- nil
	}
}
