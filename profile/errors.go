package profile

import (
	"errors"
	"fmt"
)

/******************
* EXPORTED ERRORS *
******************/

// An ErrNoCape error signals that an operation failed because a profile had no cape associated with it.
type ErrNoCape string

func (e ErrNoCape) Error() string {

	return string(e)
}

var errNoCape ErrNoCape = "profile has no cape"

// /////////////

// An ErrNoSkin error signals that an operation failed because a profile had no custom skin set.
type ErrNoSkin string

func (e ErrNoSkin) Error() string {

	return string(e)
}

var errNoSkin ErrNoSkin = "profile has no custom skin"

// /////////////

// An ErrNoSuchUser error signals that an operation failed because no profile exists with the denoted username.
type ErrNoSuchUser struct {

	// The username for which no profile could be found
	Name string
}

func (e ErrNoSuchUser) Error() string {

	return fmt.Sprintf("user %s: no such profile", e.Name)
}

// /////////////

// An ErrNoSuchID error signals that an operation failed because no profile exists with the denoted ID.
type ErrNoSuchID struct {

	// The ID for which no profile could be found
	ID string
}

func (e ErrNoSuchID) Error() string {

	return fmt.Sprintf("ID %s: no such profile", e.ID)
}

// /////////////

// An ErrTooManyRequests error occurs when the client has exceeded its server communication rate limit.
// At the time of writing, the load operations have a shared rate limit of 600 requests per 10 minutes.
type ErrTooManyRequests string

func (e ErrTooManyRequests) Error() string {

	return string(e)
}

var errTooManyRequests ErrTooManyRequests = "request rate limit exceeded"

// /////////////

// An ErrMaxSizeExceeded error occurs when LoadMany is requested to load more than LoadManyMaxSize profiles at once.
type ErrMaxSizeExceeded struct {

	// The number of profiles which were requested
	Size int
}

func (e ErrMaxSizeExceeded) Error() string {

	return fmt.Sprintf("aggregate request size of %d exceeded maximum of %d", e.Size, LoadManyMaxSize)
}

// /////////////

// Used by LoadMany to call buildProfile to exclude demo profiles from its results
var errDemo = errors.New("demo profile detected")

/************
* INTERNALS *
************/

// Extracts any Mojang error from a piece of JSON decoded using the encoding/json package
// Mojang errors are JSON objects with "error" and "errorMessage" fields
func getJsonError(json interface{}) error {

	if m, isMap := json.(map[string]interface{}); isMap {

		if e, failed := m["error"]; failed {

			error := e.(string)
			switch error {

			case "TooManyRequestsException":
				return errTooManyRequests

			default:
				const errMsg = "Mojang API error: %s; message: %s"
				return errors.New(fmt.Sprintf(errMsg, error, m["errorMessage"].(string)))
			}
		}
	}

	return nil
}
