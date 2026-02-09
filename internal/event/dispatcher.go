package event

import (
	"log/slog"
	"sync"
	"sync/atomic"
)

const defaultBufferSize = 64

// OverflowStrategy defines behavior when subscriber buffer is full.
type OverflowStrategy int

const (
	// OverflowDrop drops new events when buffer is full (current behavior, default).
	OverflowDrop OverflowStrategy = iota
	// OverflowDropOldest drops oldest events to make room for new ones.
	OverflowDropOldest
)

// SubscribeOption configures a subscription.
type SubscribeOption func(*subscriptionConfig)

type subscriptionConfig struct {
	bufferSize int
	strategy   OverflowStrategy
}

// WithBufferSize sets a custom buffer size for the subscription.
func WithBufferSize(size int) SubscribeOption {
	return func(cfg *subscriptionConfig) {
		cfg.bufferSize = size
	}
}

// WithOverflow sets the overflow strategy for the subscription.
func WithOverflow(strategy OverflowStrategy) SubscribeOption {
	return func(cfg *subscriptionConfig) {
		cfg.strategy = strategy
	}
}

// DispatcherStats holds statistics about the dispatcher.
type DispatcherStats struct {
	Published int64
	Dropped   int64
}

// Dispatcher handles fan-out event routing between plugins.
type Dispatcher struct {
	subscribers map[string][]chan Event
	strategies  map[string][]OverflowStrategy // Overflow strategy per subscriber
	mu          sync.RWMutex
	closed      bool
	logger      *slog.Logger

	// Metrics (atomic for lock-free access on hot path)
	published atomic.Int64
	dropped   atomic.Int64
}

// New creates a new event dispatcher.
func New() *Dispatcher {
	return &Dispatcher{
		subscribers: make(map[string][]chan Event),
		strategies:  make(map[string][]OverflowStrategy),
		logger:      slog.Default(),
	}
}

// NewWithLogger creates a dispatcher with custom logger.
func NewWithLogger(logger *slog.Logger) *Dispatcher {
	return &Dispatcher{
		subscribers: make(map[string][]chan Event),
		strategies:  make(map[string][]OverflowStrategy),
		logger:      logger,
	}
}

// Subscribe creates a buffered channel for receiving events on a topic.
// Accepts optional configuration (e.g., WithBufferSize, WithOverflow).
// Backward compatible: Subscribe(topic) works as before.
func (d *Dispatcher) Subscribe(topic string, opts ...SubscribeOption) <-chan Event {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		ch := make(chan Event)
		close(ch)
		return ch
	}

	// Apply options with defaults
	cfg := subscriptionConfig{
		bufferSize: defaultBufferSize,
		strategy:   OverflowDrop,
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	ch := make(chan Event, cfg.bufferSize)
	d.subscribers[topic] = append(d.subscribers[topic], ch)
	d.strategies[topic] = append(d.strategies[topic], cfg.strategy)
	return ch
}

// Publish sends an event to all subscribers of a topic.
// Non-blocking: drops events if subscriber buffer is full.
func (d *Dispatcher) Publish(topic string, e Event) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if d.closed {
		return
	}

	subs, ok := d.subscribers[topic]
	if !ok {
		return
	}

	strategies, hasStrategies := d.strategies[topic]

	for i, ch := range subs {
		d.published.Add(1)

		// Get strategy for this subscriber
		strategy := OverflowDrop // default
		if hasStrategies && i < len(strategies) {
			strategy = strategies[i]
		}

		switch strategy {
		case OverflowDropOldest:
			// DropOldest: if buffer is full, drain oldest to make room for new event
			// Check if buffer is full using len/ch (safe under RLock)
			if len(ch) == cap(ch) {
				// Buffer is full, drain oldest event
				<-ch
				d.dropped.Add(1)
			}

			// Now try to send the new event
			select {
			case ch <- e:
				// Success
			default:
				// Shouldn't happen if buffer was full and we drained one, but handle race condition
				d.dropped.Add(1)
				d.logger.Warn("event dropped", "topic", topic, "type", e.Type, "strategy", "DropOldest")
			}

		default: // OverflowDrop
			select {
			case ch <- e:
				// Success
			default:
				// Buffer full, drop event (best-effort delivery)
				d.dropped.Add(1)
				d.logger.Warn("event dropped", "topic", topic, "type", e.Type, "strategy", "Drop")
			}
		}
	}
}

// PublishAll sends an event to all subscribers of all topics.
func (d *Dispatcher) PublishAll(e Event) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if d.closed {
		return
	}

	for topic, subs := range d.subscribers {
		strategies, hasStrategies := d.strategies[topic]

		for i, ch := range subs {
			d.published.Add(1)

			// Get strategy for this subscriber
			strategy := OverflowDrop // default
			if hasStrategies && i < len(strategies) {
				strategy = strategies[i]
			}

			switch strategy {
			case OverflowDropOldest:
				// DropOldest: if buffer is full, drain oldest to make room for new event
				// Check if buffer is full using len/ch (safe under RLock)
				if len(ch) == cap(ch) {
					// Buffer is full, drain oldest event
					<-ch
					d.dropped.Add(1)
				}

				// Now try to send the new event
				select {
				case ch <- e:
					// Success
				default:
					// Shouldn't happen if buffer was full and we drained one, but handle race condition
					d.dropped.Add(1)
					d.logger.Warn("event dropped", "topic", topic, "type", e.Type, "strategy", "DropOldest")
				}

			default: // OverflowDrop
				select {
				case ch <- e:
					// Success
				default:
					d.dropped.Add(1)
					d.logger.Warn("event dropped", "topic", topic, "type", e.Type, "strategy", "Drop")
				}
			}
		}
	}
}

// Close shuts down the dispatcher and all subscriber channels.
func (d *Dispatcher) Close() {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return
	}

	d.closed = true
	for _, subs := range d.subscribers {
		for _, ch := range subs {
			close(ch)
		}
	}
	d.subscribers = nil
	d.strategies = nil
}

// Stats returns the current dispatcher statistics.
func (d *Dispatcher) Stats() DispatcherStats {
	return DispatcherStats{
		Published: d.published.Load(),
		Dropped:   d.dropped.Load(),
	}
}
