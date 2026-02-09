package adapter

import "sync"

// adapterFactories holds registered adapter constructors.
var adapterFactories []func() Adapter

// RegisterFactory registers an adapter constructor.
func RegisterFactory(factory func() Adapter) {
	adapterFactories = append(adapterFactories, factory)
}

// DetectAdapters scans for available adapters for the given project.
// All registered adapters are probed concurrently. The returned map
// is keyed by adapter ID; registration order is preserved when iterating
// is not required.
func DetectAdapters(projectRoot string) (map[string]Adapter, error) {
	factories := adapterFactories // snapshot (populated at init-time, immutable after)
	n := len(factories)
	if n == 0 {
		return nil, nil
	}

	// Each goroutine writes to its own slot — no synchronization needed
	// for the slice elements themselves.
	results := make([]Adapter, n)

	var wg sync.WaitGroup
	wg.Add(n)
	for i, factory := range factories {
		go func(idx int, fn func() Adapter) {
			defer wg.Done()
			instance := fn()
			detected, err := instance.Detect(projectRoot)
			if err == nil && detected {
				results[idx] = instance
			}
		}(i, factory)
	}
	wg.Wait()

	adapters := make(map[string]Adapter)
	for _, a := range results {
		if a != nil {
			adapters[a.ID()] = a
		}
	}
	return adapters, nil
}
