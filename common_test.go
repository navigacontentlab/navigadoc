package docformat_test

import (
	"fmt"
	"testing"
)

func must(t *testing.T, err error, msg string) {
	t.Helper()

	if err != nil {
		t.Fatalf("%s: %v", msg, err)
	}
}

func checkf(t *testing.T, err error, format string, a ...interface{}) {
	t.Helper()

	if err != nil {
		msg := fmt.Sprintf(format, a...)
		t.Errorf("%s: %v", msg, err)
	}
}
