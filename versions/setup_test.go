// +build !integration

package versions

import "net/http"

func init() {
	// Ensure examples normally run as unit tests
	client.Transport = http.NewFileTransport(http.Dir("testdata/cached"))
}
