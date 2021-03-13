package dark

import (
	"fmt"
	"strings"
	"testing"
)

// Fatalizer is an interface used to raise a fatal error. It is the
// implementers responsibility that Fatal() never returns.
type Fatalizer interface {
	Helper()
	Fatal(args ...interface{})
}

// FatalPanic implements Fatalizer using panic().
type FatalPanic string

func (p FatalPanic) Helper() {}
func (p FatalPanic) Fatal(args ...interface{}) {
	panic(FatalPanic(fmt.Sprint(args...)))
}

func CheckFatalizerBarrierErr(t testing.TB, fn func(), contains string) bool {
	t.Helper()

	err := FatalizerBarrier(fn)
	if err == nil {
		t.Errorf("dark.FatalizerBarrier() did not return an error")
		return false
	}

	if !strings.Contains(err.Error(), contains) {
		t.Errorf("dark.FatalizerBarrier() error `%s'\ndoes not contain `%s'",
			err.Error(), contains)
		return false
	}
	return true
}
