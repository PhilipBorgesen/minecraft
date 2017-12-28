package internal

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

var testErrFailedRequests = [...]struct {
	err      *FailedRequestError
	expError string
}{
	{
		err: &FailedRequestError{
			StatusCode:   404,
			ErrorCode:    "",
			ErrorMessage: "",
		},
		expError: "404 Not Found",
	},
	{
		err: &FailedRequestError{
			StatusCode:   404,
			ErrorCode:    "ErrorCode",
			ErrorMessage: "",
		},
		expError: "ErrorCode",
	},
	{
		err: &FailedRequestError{
			StatusCode:   404,
			ErrorCode:    "",
			ErrorMessage: "ErrorMessage",
		},
		expError: "ErrorMessage",
	},
	{
		err: &FailedRequestError{
			StatusCode:   404,
			ErrorCode:    "ErrorCode",
			ErrorMessage: "ErrorMessage",
		},
		expError: "ErrorCode: ErrorMessage",
	},
}

func TestErrFailedRequest_Error(t *testing.T) {
	for _, tc := range testErrFailedRequests {
		msg := tc.err.Error()
		if msg != tc.expError {
			t.Errorf(
				"%#v.Error() was %q; want %q",
				tc.err, msg, tc.expError,
			)
		}
	}
}

var testErrFailedRequest = &FailedRequestError{}
var testUnwrapErrors = [...]struct {
	err    error
	expErr *FailedRequestError
	expOk  bool
}{
	{
		err: &url.Error{
			Op:  "",
			URL: "",
			Err: testErrFailedRequest,
		},
		expErr: testErrFailedRequest,
		expOk:  true,
	},
	{
		err: &url.Error{
			Op:  "",
			URL: "",
			Err: errors.New("test"),
		},
		expErr: nil,
		expOk:  false,
	},
	{
		err:    errors.New("test"),
		expErr: nil,
		expOk:  false,
	},
}

func TestUnwrapErrFailedRequest(t *testing.T) {
	for _, tc := range testUnwrapErrors {
		err, ok := UnwrapFailedRequestError(tc.err)
		if err != tc.expErr || ok != tc.expOk {
			t.Errorf(
				"UnwrapFailedRequestError(%#v) was %q, %t; want %q, %t",
				tc.err, err, ok, tc.expErr, tc.expOk,
			)
		}
	}
}

var testParseResponseInput = [...]struct {
	response   string
	statusCode int
	op         string
	endpoint   string

	expRes interface{}
	expErr error
}{
	{
		response:   "{}",
		statusCode: 200,
		op:         "Dummy",
		endpoint:   "dummyURL",
		expRes:     make(map[string]interface{}),
		expErr:     nil,
	},
	{
		response:   "[]",
		statusCode: 200,
		op:         "Dummy",
		endpoint:   "dummyURL",
		expRes:     make([]interface{}, 0),
		expErr:     nil,
	},
	{
		response: "{\"error\": \"IllegalArgumentException\"," +
			"\"errorMessage\": \"Invalid timestamp.\"}",
		statusCode: 400,
		op:         "Dummy",
		endpoint:   "dummyURL",
		expRes:     nil,
		expErr: &url.Error{
			Op:  "Dummy",
			URL: "dummyURL",
			Err: &FailedRequestError{
				StatusCode:   400,
				ErrorCode:    "IllegalArgumentException",
				ErrorMessage: "Invalid timestamp.",
			},
		},
	},
	{
		response:   "{unexpectedError: \"Unspecified JSON error format\"}",
		statusCode: 400,
		op:         "Dummy",
		endpoint:   "dummyURL",
		expRes:     nil,
		expErr: &url.Error{
			Op:  "Dummy",
			URL: "dummyURL",
			Err: &FailedRequestError{
				StatusCode: 400,
			},
		},
	},
	{
		response:   "[\"Another unspecified JSON error format\"]",
		statusCode: 400,
		op:         "Dummy",
		endpoint:   "dummyURL",
		expRes:     nil,
		expErr: &url.Error{
			Op:  "Dummy",
			URL: "dummyURL",
			Err: &FailedRequestError{
				StatusCode: 400,
			},
		},
	},
	{
		response:   "Some non-JSON error message",
		statusCode: 400,
		op:         "Dummy",
		endpoint:   "dummyURL",
		expRes:     nil,
		expErr: &url.Error{
			Op:  "Dummy",
			URL: "dummyURL",
			Err: &FailedRequestError{
				StatusCode: 400,
			},
		},
	},
	{
		response:   "",
		statusCode: 204,
		op:         "Dummy",
		endpoint:   "dummyURL",
		expRes:     nil,
		expErr: &url.Error{
			Op:  "Dummy",
			URL: "dummyURL",
			Err: &FailedRequestError{
				StatusCode: 204,
			},
		},
	},
	{
		response:   "",
		statusCode: 200,
		op:         "Dummy",
		endpoint:   "dummyURL",
		expRes:     nil,
		expErr: &url.Error{
			Op:  "Parse",
			URL: "dummyURL",
			Err: io.EOF,
		},
	},
}

func TestParseResponse(t *testing.T) {
	for _, tc := range testParseResponseInput {
		r := ioutil.NopCloser(strings.NewReader(tc.response))
		res, err := parseResponse(r, tc.statusCode, tc.op, tc.endpoint)
		if !reflect.DeepEqual(res, tc.expRes) || !reflect.DeepEqual(err, tc.expErr) {
			t.Errorf(
				"parseResponse(%q, %d, %q, %q)\n"+
					"  was  %#v, %s\n"+
					"  want %#v, %s",
				tc.response, tc.statusCode, tc.op, tc.endpoint,
				res, p(err),
				tc.expRes, p(tc.expErr),
			)
		}
	}
}

var testFetchJSONInput = [...]struct {
	transport http.RoundTripper
	endpoint  string
	expRes    interface{}
	expErr    error
}{
	{
		transport: errorTransport{testError},
		endpoint:  "dummyURL",
		expRes:    nil,
		expErr: &url.Error{
			Op:  "Get",
			URL: "dummyURL",
			Err: testError,
		},
	},
	{
		transport: http.NewFileTransport(http.Dir("testdata")),
		endpoint:  "http://does.not/exist/at.all",
		expRes:    nil,
		expErr: &url.Error{
			Op:  "Get",
			URL: "http://does.not/exist/at.all",
			Err: &FailedRequestError{
				StatusCode: 404,
			},
		},
	},
	{
		transport: http.NewFileTransport(http.Dir("testdata")),
		endpoint:  "data.json",
		expRes:    make(map[string]interface{}),
		expErr:    nil,
	},
}

func TestFetchJSON(t *testing.T) {
	for _, tc := range testFetchJSONInput {
		ctx := context.Background()
		client := &http.Client{Transport: tc.transport}

		res, err := FetchJSON(ctx, client, tc.endpoint)
		if !reflect.DeepEqual(res, tc.expRes) || !reflect.DeepEqual(err, tc.expErr) {
			t.Errorf(
				"FetchJSON(ctx, client(%#v), %q)\n"+
					"  was  %#v, %s\n"+
					"  want %#v, %s",
				tc.transport, tc.endpoint,
				res, p(err),
				tc.expRes, p(tc.expErr),
			)
		}
	}
}

func TestFetchJSONContextUsed(t *testing.T) {
	ctx := context.WithValue(context.Background(), dummy, nil)
	ct := CtxStoreTransport{}

	client := &http.Client{}
	client.Transport = &ct
	FetchJSON(ctx, client, "dummyURL")

	if ct.Context != ctx {
		t.Error("FetchJSON(ctx, client, endpoint) didn't pass context to underlying http.Client")
	}
}

var testExchangeJSONInput = [...]struct {
	transport http.RoundTripper
	endpoint  string
	data      interface{}
	expRes    interface{}
	expErr    error
}{
	{
		transport: errorTransport{testError},
		endpoint:  "dummyURL",
		data:      make(map[string]interface{}),
		expRes:    nil,
		expErr: &url.Error{
			Op:  "Post",
			URL: "dummyURL",
			Err: testError,
		},
	},
	{
		transport: http.NewFileTransport(http.Dir("testdata")),
		endpoint:  "data.json",
		data:      func() {},
		expRes:    nil,
		expErr:    &json.UnsupportedTypeError{Type: reflect.TypeOf(func() {})},
	},
	{
		transport: http.NewFileTransport(http.Dir("testdata")),
		endpoint:  "http://does.not/exist/at.all",
		data:      make(map[string]interface{}),
		expRes:    nil,
		expErr: &url.Error{
			Op:  "Post",
			URL: "http://does.not/exist/at.all",
			Err: &FailedRequestError{
				StatusCode: 404,
			},
		},
	},
	{
		transport: http.NewFileTransport(http.Dir("testdata")),
		endpoint:  "data.json",
		data:      make(map[string]interface{}),
		expRes:    make(map[string]interface{}),
		expErr:    nil,
	},
}

func TestExchangeJSON(t *testing.T) {
	for _, tc := range testExchangeJSONInput {
		ctx := context.Background()
		client := &http.Client{Transport: tc.transport}

		res, err := ExchangeJSON(ctx, client, tc.endpoint, tc.data)
		if !reflect.DeepEqual(res, tc.expRes) || !reflect.DeepEqual(err, tc.expErr) {
			t.Errorf(
				"ExchangeJSON(ctx, client(%#v), %q, %#v)\n"+
					"  was  %#v, %s\n"+
					"  want %#v, %s",
				tc.transport, tc.endpoint, tc.data,
				res, p(err),
				tc.expRes, p(tc.expErr),
			)
		}
	}
}

func TestExchangeJSONContextUsed(t *testing.T) {
	ctx := context.WithValue(context.Background(), dummy, nil)
	ct := CtxStoreTransport{}

	client := &http.Client{}
	client.Transport = &ct
	ExchangeJSON(ctx, client, "dummyURL", nil)

	if ct.Context != ctx {
		t.Error("ExchangeJSON(ctx, client, endpoint, nil) didn't pass context to underlying http.Client")
	}
}

/*************
* TEST UTILS *
*************/

var dummy struct{}

type errorTransport struct {
	err error
}

func (et errorTransport) RoundTrip(_ *http.Request) (*http.Response, error) {
	return nil, et.err
}

type CtxStoreTransport struct {
	Context context.Context
}

func (ct *CtxStoreTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	ct.Context = req.Context()
	return nil, errors.New("RoundTrip was called")
}

var testError = errors.New("test")

func p(x interface{}) interface{} {
	if x == nil {
		return "<nil>"
	} else {
		return x
	}
}
