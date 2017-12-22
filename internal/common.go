package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

var ErrUnknownFormat = errors.New("unknown JSON data format")

// Non-200 responses from Mojang servers, incl. potential JSON error types and messages.
type FailedRequestError struct {
	StatusCode   int
	ErrorCode    string
	ErrorMessage string
}

func (err *FailedRequestError) Error() string {
	if err.ErrorCode != "" {
		if err.ErrorMessage != "" {
			return err.ErrorCode + ": " + err.ErrorMessage
		} else {
			return err.ErrorCode
		}
	} else if err.ErrorMessage != "" {
		return err.ErrorMessage
	} else {
		code := err.StatusCode
		return fmt.Sprintf("%d %s", code, http.StatusText(code))
	}
}

// GET JSON from an url and parse it into a map hierarchy
// If a non-200 response is returned, the returned url.Error wraps an FailedRequestError.
func FetchJSON(ctx context.Context, client *http.Client, endpoint string) (interface{}, error) {
	// Fetch JSON
	req, _ := http.NewRequest("GET", endpoint, nil) // Error only occurs if endpoint is bad
	req = req.WithContext(ctx)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return parseResponse(resp.Body, resp.StatusCode, "Get", endpoint)
}

// POST JSON to an url and parse the response JSON into a map hierarchy
// If a non-200 response is returned, the returned url.Error wraps an FailedRequestError.
func ExchangeJSON(ctx context.Context, client *http.Client, endpoint string, data interface{}) (interface{}, error) {
	buf := bytes.Buffer{}
	err := json.NewEncoder(&buf).Encode(data)
	if err != nil {
		return nil, err
	}

	req, _ := http.NewRequest("POST", endpoint, &buf) // Error only occurs if endpoint is bad
	req = req.WithContext(ctx)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return parseResponse(resp.Body, resp.StatusCode, "Post", endpoint)
}

func parseResponse(r io.ReadCloser, statusCode int, op, endpoint string) (interface{}, error) {
	var j interface{}
	parseErr := json.NewDecoder(r).Decode(&j)

	if statusCode != 200 {
		err := &FailedRequestError{
			StatusCode: statusCode,
		}
		if j, ok := j.(map[string]interface{}); ok && parseErr == nil {
			if e, ok := j["error"]; ok {
				err.ErrorCode, _ = e.(string)
			}
			if em, ok := j["errorMessage"]; ok {
				err.ErrorMessage, _ = em.(string)
			}
		}
		return nil, &url.Error{
			Op:  op,
			URL: endpoint,
			Err: err,
		}
	}
	if parseErr != nil {
		return nil, &url.Error{
			Op:  "Parse",
			URL: endpoint,
			Err: parseErr,
		}
	}
	return j, nil
}

///////////////////

func UnwrapFailedRequestError(uerr error) (err *FailedRequestError, ok bool) {
	if e, match := uerr.(*url.Error); match {
		err, ok = e.Err.(*FailedRequestError)
	}
	return
}
