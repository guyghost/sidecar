// Package conversations provides content search execution for cross-conversation search.
package conversations

import (
	"context"
	"sort"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/marcus/sidecar/internal/adapter"
)

const (
	// searchConcurrency is the max concurrent search goroutines.
	searchConcurrency = 4
	// searchTimeout is the max duration for the entire search operation.
	searchTimeout = 30 * time.Second
	// maxTotalMatches is the global match limit for early termination.
	maxTotalMatches = 500
	// debounceDelay is the delay before executing search after input.
	debounceDelay = 200 * time.Millisecond
)

// RunContentSearch executes search across all sessions using their adapters.
// Returns a tea.Cmd that produces ContentSearchResultsMsg when complete.
// Search runs in parallel with concurrency limit, timeout, and match cap.
func RunContentSearch(query string, sessions []adapter.Session,
	adapters map[string]adapter.Adapter, opts adapter.SearchOptions) tea.Cmd {
	return func() tea.Msg {
		if query == "" {
			return ContentSearchResultsMsg{Results: nil}
		}

		var results []SessionSearchResult
		var mu sync.Mutex
		var wg sync.WaitGroup
		sem := make(chan struct{}, searchConcurrency)

		ctx, cancel := context.WithTimeout(context.Background(), searchTimeout)
		defer cancel()

		totalMatches := 0
		done := make(chan struct{})

	sessionLoop:
		for _, session := range sessions {
			// Check if we've hit the match limit
			mu.Lock()
			if totalMatches >= maxTotalMatches {
				mu.Unlock()
				break sessionLoop
			}
			mu.Unlock()

			// Check context cancellation
			select {
			case <-ctx.Done():
				break sessionLoop
			default:
			}

			wg.Add(1)
			go func(s adapter.Session) {
				defer wg.Done()

				// Acquire semaphore or bail on context cancel
				select {
				case sem <- struct{}{}:
					defer func() { <-sem }()
				case <-ctx.Done():
					return
				}

				// Get adapter for this session
				adp, ok := adapters[s.AdapterID]
				if !ok {
					return
				}

				// Check if adapter supports search
				searcher, ok := adp.(adapter.MessageSearcher)
				if !ok {
					return
				}

				// Execute search
				matches, err := searcher.SearchMessages(s.ID, query, opts)
				if err != nil || len(matches) == 0 {
					return
				}

				matchCount := countMatches(matches)

				mu.Lock()
				results = append(results, SessionSearchResult{
					Session:   s,
					Messages:  matches,
					Collapsed: false,
				})
				totalMatches += matchCount
				mu.Unlock()
			}(session)
		}

		// Wait for all goroutines in a separate goroutine
		go func() {
			wg.Wait()
			close(done)
		}()

		// Wait for completion or context timeout
		select {
		case <-done:
		case <-ctx.Done():
		}

		// Sort results by session UpdatedAt descending (most recent first)
		sort.Slice(results, func(i, j int) bool {
			return results[i].Session.UpdatedAt.After(results[j].Session.UpdatedAt)
		})

		return ContentSearchResultsMsg{Results: results}
	}
}

// countMatches returns total ContentMatch count across messages.
func countMatches(matches []adapter.MessageMatch) int {
	count := 0
	for _, m := range matches {
		count += len(m.Matches)
	}
	return count
}

// scheduleContentSearch returns a tea.Cmd that triggers search after debounce.
// The returned Tick sends ContentSearchDebounceMsg after debounceDelay.
func scheduleContentSearch(query string, version int) tea.Cmd {
	return tea.Tick(debounceDelay, func(t time.Time) tea.Msg {
		return ContentSearchDebounceMsg{
			Version: version,
			Query:   query,
		}
	})
}
