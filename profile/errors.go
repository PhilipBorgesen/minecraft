package profile

import (
	"errors"
	"fmt"
)

var (
	ErrNoCape        = errors.New("profile has no cape")
	ErrNoSuchProfile = errors.New("no such profile")

	// ErrTooManyRequests is returned when the client has exceeded its
	// server communication rate limit. At the time of writing, the load
	// operations have a shared rate limit of 600 requests per 10 minutes.
	//
	// Note that the rate limit for reading profile properties is much
	// stricter: For each profile, profile properties may only be requested
	// once per minute.
	ErrTooManyRequests = errors.New("request rate limit exceeded")
)

// An ErrMaxSizeExceeded error is returned when LoadMany is requested to load more than
// LoadManyMaxSize profiles at once.
type ErrMaxSizeExceeded struct {
	// The number of profiles which were requested
	Size int
}

func (e ErrMaxSizeExceeded) Error() string {
	return fmt.Sprintf("aggregate request size of %d exceeded maximum of %d", e.Size, LoadManyMaxSize)
}
