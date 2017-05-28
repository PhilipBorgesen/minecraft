package profile

import (
	"strconv"
	"strings"
	"testing"
)

func TestErrMaxSizeExceeded_Error(t *testing.T) {
	const exceededCount = 12345
	err := ErrMaxSizeExceeded{exceededCount}
	msg := err.Error()
	if !strings.Contains(msg, strconv.Itoa(exceededCount)) || !strings.Contains(msg, strconv.Itoa(LoadManyMaxSize)) {
		t.Errorf(
			"ErrMaxSizeExceeded{exceededCount}.Error()\n"+
				"  was:  %q\n"+
				"  want: message containing exceededCount (%d) and LoadManyMaxSize (%d)",
			msg, exceededCount, LoadManyMaxSize,
		)
	}
}
