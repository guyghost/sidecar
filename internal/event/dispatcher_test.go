package event

import (
	"sync"
	"testing"
	"time"
)

func TestDispatcher_SubscribePublish(t *testing.T) {
	d := New()
	defer d.Close()

	ch := d.Subscribe("test")
	e := NewEvent(TypeFileChanged, "test", "data")

	d.Publish("test", e)

	select {
	case received := <-ch:
		if received.Type != e.Type {
			t.Errorf("got type %s, want %s", received.Type, e.Type)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("timeout waiting for event")
	}
}

func TestDispatcher_MultipleSubscribers(t *testing.T) {
	d := New()
	defer d.Close()

	ch1 := d.Subscribe("test")
	ch2 := d.Subscribe("test")
	e := NewEvent(TypeTDUpdate, "test", nil)

	d.Publish("test", e)

	for i, ch := range []<-chan Event{ch1, ch2} {
		select {
		case received := <-ch:
			if received.Type != e.Type {
				t.Errorf("sub %d: got type %s, want %s", i, received.Type, e.Type)
			}
		case <-time.After(100 * time.Millisecond):
			t.Errorf("sub %d: timeout waiting for event", i)
		}
	}
}

func TestDispatcher_NonBlocking(t *testing.T) {
	d := New()
	defer d.Close()

	ch := d.Subscribe("test")
	e := NewEvent(TypeError, "test", nil)

	// Fill buffer
	for i := 0; i < defaultBufferSize; i++ {
		d.Publish("test", e)
	}

	// This should not block
	done := make(chan bool)
	go func() {
		d.Publish("test", e) // Should drop
		done <- true
	}()

	select {
	case <-done:
		// Success - didn't block
	case <-time.After(100 * time.Millisecond):
		t.Error("Publish blocked with full buffer")
	}

	// Drain and verify we got buffer size events
	count := 0
	for {
		select {
		case <-ch:
			count++
		default:
			if count != defaultBufferSize {
				t.Errorf("got %d events, want %d", count, defaultBufferSize)
			}
			return
		}
	}
}

func TestDispatcher_Close(t *testing.T) {
	d := New()
	ch := d.Subscribe("test")

	d.Close()

	// Channel should be closed
	select {
	case _, ok := <-ch:
		if ok {
			t.Error("channel should be closed")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("timeout - channel not closed")
	}

	// Publish after close should not panic
	d.Publish("test", NewEvent(TypeError, "test", nil))
}

func TestDispatcher_Concurrent(t *testing.T) {
	d := New()

	const numSubscribers = 5
	const numPublishers = 5
	const numEvents = 20

	var pubWg sync.WaitGroup
	pubWg.Add(numPublishers)

	// Subscribe before publishing
	channels := make([]<-chan Event, numSubscribers)
	for i := 0; i < numSubscribers; i++ {
		channels[i] = d.Subscribe("concurrent")
	}

	// Publishers
	for i := 0; i < numPublishers; i++ {
		go func() {
			defer pubWg.Done()
			for j := 0; j < numEvents; j++ {
				d.Publish("concurrent", NewEvent(TypeRefreshNeeded, "concurrent", j))
			}
		}()
	}

	// Wait for publishers to finish
	pubWg.Wait()

	// Close dispatcher - this closes all channels
	d.Close()

	// Drain channels - verify no panic and channels are closed
	for _, ch := range channels {
		for range ch {
			// Drain
		}
	}
}

func TestDispatcher_WithBufferSize(t *testing.T) {
	d := New()
	defer d.Close()

	// Subscribe with custom buffer size
	ch := d.Subscribe("test", WithBufferSize(10))

	// Fill buffer
	e := NewEvent(TypeFileChanged, "test", nil)
	for i := 0; i < 10; i++ {
		d.Publish("test", e)
	}

	// Verify all 10 events are in buffer
	for i := 0; i < 10; i++ {
		select {
		case <-ch:
		case <-time.After(100 * time.Millisecond):
			t.Errorf("failed to receive event %d", i)
		}
	}

	// Channel should be empty now
	select {
	case <-ch:
		t.Error("unexpected event in channel")
	default:
		// Expected
	}
}

func TestDispatcher_BackwardCompatibility(t *testing.T) {
	d := New()
	defer d.Close()

	// Subscribe without options (backward compatibility)
	ch := d.Subscribe("test")

	e := NewEvent(TypeFileChanged, "test", "data")
	d.Publish("test", e)

	select {
	case received := <-ch:
		if received.Type != e.Type {
			t.Errorf("got type %s, want %s", received.Type, e.Type)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("timeout waiting for event")
	}
}

func TestDispatcher_WithOverflowDrop(t *testing.T) {
	d := New()
	defer d.Close()

	// Subscribe with Drop strategy (default)
	ch := d.Subscribe("test", WithOverflow(OverflowDrop))

	e := NewEvent(TypeFileChanged, "test", nil)

	// Fill buffer (default is 64)
	for i := 0; i < 64; i++ {
		d.Publish("test", e)
	}

	// Publish one more - should be dropped
	d.Publish("test", e)

	stats := d.Stats()
	if stats.Published != 65 {
		t.Errorf("expected 65 published, got %d", stats.Published)
	}
	if stats.Dropped != 1 {
		t.Errorf("expected 1 dropped, got %d", stats.Dropped)
	}

	// Drain buffer - should only have 64 events
	count := 0
	for {
		select {
		case <-ch:
			count++
		default:
			if count != 64 {
				t.Errorf("expected 64 events, got %d", count)
			}
			return
		}
	}
}

func TestDispatcher_WithOverflowDropOldest(t *testing.T) {
	d := New()
	defer d.Close()

	// Subscribe with DropOldest strategy
	ch := d.Subscribe("test", WithOverflow(OverflowDropOldest), WithBufferSize(5))

	// Fill buffer with numbered events
	for i := 0; i < 5; i++ {
		d.Publish("test", NewEvent(TypeFileChanged, "test", i))
	}

	// Publish one more - should drop oldest (event 0)
	d.Publish("test", NewEvent(TypeFileChanged, "test", 5))

	stats := d.Stats()
	if stats.Published != 6 {
		t.Errorf("expected 6 published, got %d", stats.Published)
	}
	if stats.Dropped != 1 {
		t.Errorf("expected 1 dropped, got %d", stats.Dropped)
	}

	// Drain buffer - should have events 1,2,3,4,5 (not 0)
	expectedValues := []any{1, 2, 3, 4, 5}
	for i, expected := range expectedValues {
		select {
		case received := <-ch:
			if received.Data != expected {
				t.Errorf("event %d: got %v, want %v", i, received.Data, expected)
			}
		case <-time.After(100 * time.Millisecond):
			t.Errorf("timeout waiting for event %d", i)
		}
	}
}

func TestDispatcher_Stats(t *testing.T) {
	d := New()
	defer d.Close()

	// Initial stats
	stats := d.Stats()
	if stats.Published != 0 || stats.Dropped != 0 {
		t.Errorf("initial stats should be zero, got Published=%d, Dropped=%d", stats.Published, stats.Dropped)
	}

	ch := d.Subscribe("test", WithBufferSize(2))
	e := NewEvent(TypeFileChanged, "test", nil)

	// Publish 2 events
	d.Publish("test", e)
	d.Publish("test", e)

	stats = d.Stats()
	if stats.Published != 2 || stats.Dropped != 0 {
		t.Errorf("after 2 publishes: got Published=%d, Dropped=%d", stats.Published, stats.Dropped)
	}

	// Verify both events are in buffer
	<-ch
	<-ch

	// Publish 3 more - 2 will fit (buffer was empty), 1 will drop (buffer full)
	d.Publish("test", e)
	d.Publish("test", e)
	d.Publish("test", e)

	stats = d.Stats()
	if stats.Published != 5 {
		t.Errorf("expected 5 published, got %d", stats.Published)
	}
	if stats.Dropped != 1 {
		t.Errorf("expected 1 dropped, got %d", stats.Dropped)
	}
}

func TestDispatcher_OverflowDropOldestMultiple(t *testing.T) {
	d := New()
	defer d.Close()

	// Subscribe with DropOldest strategy and small buffer
	ch := d.Subscribe("test", WithOverflow(OverflowDropOldest), WithBufferSize(3))

	// Publish more events than buffer
	for i := 0; i < 10; i++ {
		d.Publish("test", NewEvent(TypeFileChanged, "test", i))
	}

	stats := d.Stats()
	if stats.Published != 10 {
		t.Errorf("expected 10 published, got %d", stats.Published)
	}
	if stats.Dropped != 7 {
		t.Errorf("expected 7 dropped, got %d", stats.Dropped)
	}

	// Drain buffer - should have last 3 events (7, 8, 9)
	expectedValues := []any{7, 8, 9}
	for i, expected := range expectedValues {
		select {
		case received := <-ch:
			if received.Data != expected {
				t.Errorf("event %d: got %v, want %v", i, received.Data, expected)
			}
		case <-time.After(100 * time.Millisecond):
			t.Errorf("timeout waiting for event %d", i)
		}
	}
}
