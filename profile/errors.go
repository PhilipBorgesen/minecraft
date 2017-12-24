package profile

import (
	"errors"
	"fmt"
)

var (
	ErrNoCape        = errors.New("minecraft/profile: profile has no cape")
	ErrNoSuchProfile = errors.New("minecraft/profile: no such profile")
	ErrUnsetPlayerID = errors.New("minecraft/profile: player id is not set")
	ErrUnknownModel  = errors.New("minecraft/profile: unknown model")

	// ErrTooManyRequests is returned if the client has exceeded its server
	// communication rate limit. At the time of writing, the load operations
	// have a shared rate limit of 600 requests per 10 minutes.
	//
	// Note that the rate limit for reading profile properties is much
	// stricter: For each profile, profile properties may only be requested
	// once per minute.
	ErrTooManyRequests = errors.New("minecraft/profile: request rate limit exceeded")
)

// An ErrMaxSizeExceeded error is returned when LoadMany is requested to load
// more than LoadManyMaxSize profiles at once.
type ErrMaxSizeExceeded struct {
	Size int // Number of profiles which were requested.
}

func (e ErrMaxSizeExceeded) Error() string {
	return fmt.Sprintf("minecraft/profile: aggregate request size of %d exceeded maximum of %d", e.Size, LoadManyMaxSize)
}
