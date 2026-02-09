package adapter

import (
	"io"
	"sync/atomic"
	"testing"
	"time"
)

// stubAdapter is a minimal Adapter for testing detection.
type stubAdapter struct {
	id       string
	detected bool
	delay    time.Duration
}

func (s *stubAdapter) ID() string                                      { return s.id }
func (s *stubAdapter) Name() string                                    { return s.id }
func (s *stubAdapter) Icon() string                                    { return "?" }
func (s *stubAdapter) Capabilities() CapabilitySet                     { return nil }
func (s *stubAdapter) Sessions(_ string) ([]Session, error)            { return nil, nil }
func (s *stubAdapter) Messages(_ string) ([]Message, error)            { return nil, nil }
func (s *stubAdapter) Usage(_ string) (*UsageStats, error)             { return nil, nil }
func (s *stubAdapter) Watch(_ string) (<-chan Event, io.Closer, error) { return nil, nil, nil }
func (s *stubAdapter) Detect(_ string) (bool, error) {
	if s.delay > 0 {
		time.Sleep(s.delay)
	}
	return s.detected, nil
}

func TestDetectAdapters_Concurrent(t *testing.T) {
	// Save and restore global factories.
	saved := adapterFactories
	defer func() { adapterFactories = saved }()

	adapterFactories = nil

	RegisterFactory(func() Adapter {
		return &stubAdapter{id: "a", detected: true, delay: 20 * time.Millisecond}
	})
	RegisterFactory(func() Adapter {
		return &stubAdapter{id: "b", detected: false, delay: 20 * time.Millisecond}
	})
	RegisterFactory(func() Adapter {
		return &stubAdapter{id: "c", detected: true, delay: 20 * time.Millisecond}
	})

	start := time.Now()
	adapters, err := DetectAdapters("/fake/project")
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should find only the detected ones.
	if len(adapters) != 2 {
		t.Fatalf("expected 2 adapters, got %d", len(adapters))
	}
	if _, ok := adapters["a"]; !ok {
		t.Error("expected adapter 'a' to be detected")
	}
	if _, ok := adapters["c"]; !ok {
		t.Error("expected adapter 'c' to be detected")
	}
	if _, ok := adapters["b"]; ok {
		t.Error("adapter 'b' should not be detected")
	}

	// With 3 adapters each taking 20ms, sequential would be ~60ms.
	// Concurrent should complete in ~20-30ms. Allow generous margin.
	if elapsed > 50*time.Millisecond {
		t.Errorf("detection took %v, expected concurrent execution (< 50ms)", elapsed)
	}
}

func TestDetectAdapters_EmptyFactories(t *testing.T) {
	saved := adapterFactories
	defer func() { adapterFactories = saved }()

	adapterFactories = nil

	adapters, err := DetectAdapters("/fake/project")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if adapters != nil {
		t.Fatalf("expected nil, got %v", adapters)
	}
}

func TestDetectAdapters_IndependentInstances(t *testing.T) {
	// Verify each goroutine gets its own instance (factory called per goroutine).
	saved := adapterFactories
	defer func() { adapterFactories = saved }()

	adapterFactories = nil

	var callCount atomic.Int32
	RegisterFactory(func() Adapter {
		callCount.Add(1)
		return &stubAdapter{id: "x", detected: true}
	})
	RegisterFactory(func() Adapter {
		callCount.Add(1)
		return &stubAdapter{id: "y", detected: true}
	})

	_, err := DetectAdapters("/fake/project")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if callCount.Load() != 2 {
		t.Errorf("expected 2 factory calls, got %d", callCount.Load())
	}
}
