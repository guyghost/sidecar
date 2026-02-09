package adapter

import (
	"errors"
	"fmt"
)

// PartialResult wraps adapter results that may be incomplete.
// When an adapter encounters a parse error but has already recovered
// some data, it should return the partial data alongside a PartialResult
// error. Callers can use IsPartial to check for this condition and
// display the recovered data with a warning indicator.
type PartialResult struct {
	// Err is the underlying parse error encountered.
	Err error
	// ParsedCount is the number of items successfully parsed before the error.
	ParsedCount int
	// Reason is a short human-readable description of the partial condition.
	Reason string
}

func (p *PartialResult) Error() string {
	return fmt.Sprintf("partial result (%s): %d items recovered: %v", p.Reason, p.ParsedCount, p.Err)
}

func (p *PartialResult) Unwrap() error {
	return p.Err
}

// IsPartial checks if an error contains partial result metadata.
// Returns the PartialResult and true if the error (or any wrapped error)
// is a *PartialResult.
func IsPartial(err error) (*PartialResult, bool) {
	var pr *PartialResult
	if errors.As(err, &pr) {
		return pr, true
	}
	return nil, false
}
