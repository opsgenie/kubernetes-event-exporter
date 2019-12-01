package batch

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestSimpleWriter(t *testing.T) {
	cfg := WriterConfig{
		BatchSize:  10,
		MaxRetries: 3,
		Interval:   time.Second * 2,
	}

	var allItems []interface{}
	w := NewWriter(cfg, func(ctx context.Context, items []interface{}) []bool {
		resp := make([]bool, len(items))
		for idx := range resp {
			resp[idx] = true
		}

		allItems = items
		return resp
	})

	w.Start()
	w.Submit(1, 2, 3, 4, 5, 6, 7)
	assert.Len(t, allItems, 0)
	w.Stop()

	assert.Len(t, allItems, 7)
	for i := 0; i < 7; i++ {
		assert.Equal(t, allItems[i], i+1)
	}
}

func TestCorrectnessManyTimes(t *testing.T) {
	// Surely this is not the proper way to do it but anyways
	for i := 0; i < 10000; i++ {
		TestSimpleWriter(t)
	}
}

func TestLargerThanBatchSize(t *testing.T) {
	cfg := WriterConfig{
		BatchSize:  3,
		MaxRetries: 3,
		Interval:   time.Second * 2,
	}

	allItems := make([][]interface{}, 0)
	w := NewWriter(cfg, func(ctx context.Context, items []interface{}) []bool {
		resp := make([]bool, len(items))
		for idx := range resp {
			resp[idx] = true
		}

		allItems = append(allItems, items)
		return resp
	})

	w.Start()
	w.Submit(1, 2, 3, 4, 5, 6, 7)
	w.Stop()

	assert.Len(t, allItems, 3)
	assert.Equal(t, allItems[0], []interface{}{1, 2, 3})
	assert.Equal(t, allItems[1], []interface{}{4, 5, 6})
	assert.Equal(t, allItems[2], []interface{}{7})
}

func TestSimpleInterval(t *testing.T) {
	cfg := WriterConfig{
		BatchSize:  5,
		MaxRetries: 3,
		Interval:   time.Millisecond * 20,
	}

	allItems := make([][]interface{}, 0)
	w := NewWriter(cfg, func(ctx context.Context, items []interface{}) []bool {
		resp := make([]bool, len(items))
		for idx := range resp {
			resp[idx] = true
		}

		allItems = append(allItems, items)
		return resp
	})

	w.Start()
	w.Submit(1, 2)
	time.Sleep(time.Millisecond * 5)
	assert.Len(t, allItems, 0)

	time.Sleep(time.Millisecond * 50)
	assert.Len(t, allItems, 1)
	assert.Equal(t, allItems[0], []interface{}{1, 2})

	w.Stop()
	assert.Len(t, allItems, 1)
}

func TestIntervalComplex(t *testing.T) {
	cfg := WriterConfig{
		BatchSize:  5,
		MaxRetries: 3,
		Interval:   time.Millisecond * 20,
	}

	allItems := make([][]interface{}, 0)
	w := NewWriter(cfg, func(ctx context.Context, items []interface{}) []bool {
		resp := make([]bool, len(items))
		for idx := range resp {
			resp[idx] = true
		}

		allItems = append(allItems, items)
		return resp
	})

	w.Start()
	w.Submit(1, 2)
	time.Sleep(time.Millisecond * 5)
	w.Submit(3, 4)
	assert.Len(t, allItems, 0)

	time.Sleep(time.Millisecond * 50)
	assert.Len(t, allItems, 1)
	assert.Equal(t, allItems[0], []interface{}{1, 2, 3, 4})

	w.Stop()
	assert.Len(t, allItems, 1)
}

func TestIntervalComplexAfterFlush(t *testing.T) {
	cfg := WriterConfig{
		BatchSize:  5,
		MaxRetries: 3,
		Interval:   time.Millisecond * 20,
	}

	allItems := make([][]interface{}, 0)
	w := NewWriter(cfg, func(ctx context.Context, items []interface{}) []bool {
		resp := make([]bool, len(items))
		for idx := range resp {
			resp[idx] = true
		}

		allItems = append(allItems, items)
		return resp
	})

	w.Start()
	w.Submit(1, 2)
	time.Sleep(time.Millisecond * 5)
	w.Submit(3, 4)
	assert.Len(t, allItems, 0)

	time.Sleep(time.Millisecond * 50)
	assert.Len(t, allItems, 1)
	assert.Equal(t, allItems[0], []interface{}{1, 2, 3, 4})

	w.Submit(5, 6, 7)
	w.Stop()

	assert.Len(t, allItems, 2)
	assert.Equal(t, allItems[1], []interface{}{5, 6, 7})
}

func TestRetry(t *testing.T) {
	cfg := WriterConfig{
		BatchSize:  5,
		MaxRetries: 3,
		Interval:   time.Millisecond * 10,
	}

	allItems := make([][]interface{}, 0)
	w := NewWriter(cfg, func(ctx context.Context, items []interface{}) []bool {
		resp := make([]bool, len(items))
		for idx := range resp {
			resp[idx] = items[idx] != 2
		}

		allItems = append(allItems, items)
		return resp
	})

	w.Start()
	w.Submit(1, 2, 3)
	assert.Len(t, allItems, 0)

	time.Sleep(time.Millisecond * 200)
	assert.Len(t, allItems, 4)

	assert.Equal(t, allItems[0], []interface{}{1, 2, 3})
	assert.Equal(t, allItems[1], []interface{}{2})
	assert.Equal(t, allItems[2], []interface{}{2})
	assert.Equal(t, allItems[3], []interface{}{2})
}
