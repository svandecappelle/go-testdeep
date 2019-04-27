// Copyright (c) 2018, Maxime Soulé
// All rights reserved.
//
// This source code is licensed under the BSD-style license found in the
// LICENSE file in the root directory of this source tree.

package testdeep

import (
	"fmt"
	"reflect"
	"unsafe"

	"github.com/maxatome/go-testdeep/internal/ctxerr"
	"github.com/maxatome/go-testdeep/internal/types"
)

type tdShallow struct {
	Base
	expectedKind    reflect.Kind
	expectedPointer uintptr
	expectedStr     string // in reflect.String case, to avoid contents  GC
}

var _ TestDeep = &tdShallow{}

func stringPointer(s string) uintptr {
	return (*reflect.StringHeader)(unsafe.Pointer(&s)).Data
}

// Shallow operator compares pointers only, not their contents. It
// applies on channels, functions (with some restrictions), maps,
// pointers, slices and strings.
//
// During a match, the compared data must be the same as
// "expectedPointer" to succeed.
//
//   a, b := 123, 123
//   Cmp(t, &a, Shallow(&a)) // succeeds
//   Cmp(t, &a, Shallow(&b)) // fails even if a == b as &a != &b
//
//   back := "foobarfoobar"
//   a, b := back[:6], back[6:]
//   // a == b but...
//   Cmp(t, &a, Shallow(&b)) // fails
//
// Be careful for slices and strings! Shallow can succeed but the
// slices/strings not be identical because of their different
// lengths. For example:
//
//   a := "foobar yes!"
//   b := a[:1]                    // aka. "f"
//   Cmp(t, &a, Shallow(&b)) // succeeds as both strings point to the same area, even if len() differ
//
// The same behavior occurs for slices:
//
//   a := []int{1, 2, 3, 4, 5, 6}
//   b := a[:2]                    // aka. []int{1, 2}
//   Cmp(t, &a, Shallow(&b)) // succeeds as both slices point to the same area, even if len() differ
func Shallow(expectedPtr interface{}) TestDeep {
	vptr := reflect.ValueOf(expectedPtr)

	shallow := tdShallow{
		Base:         NewBase(3),
		expectedKind: vptr.Kind(),
	}

	// Note from reflect documentation:
	// If v's Kind is Func, the returned pointer is an underlying code
	// pointer, but not necessarily enough to identify a single function
	// uniquely. The only guarantee is that the result is zero if and
	// only if v is a nil func Value.

	switch shallow.expectedKind {
	case reflect.Chan,
		reflect.Func,
		reflect.Map,
		reflect.Ptr,
		reflect.Slice,
		reflect.UnsafePointer:
		shallow.expectedPointer = vptr.Pointer()
		return &shallow

	case reflect.String:
		shallow.expectedStr = vptr.String()
		shallow.expectedPointer = stringPointer(shallow.expectedStr)
		return &shallow

	default:
		panic("usage: Shallow(CHANNEL|FUNC|MAP|PTR|SLICE|UNSAFE_PTR|STRING)")
	}
}

func (s *tdShallow) Match(ctx ctxerr.Context, got reflect.Value) *ctxerr.Error {
	if got.Kind() != s.expectedKind {
		if ctx.BooleanError {
			return ctxerr.BooleanError
		}
		return ctx.CollectError(&ctxerr.Error{
			Message:  "bad kind",
			Got:      types.RawString(got.Kind().String()),
			Expected: types.RawString(s.expectedKind.String()),
		})
	}

	var ptr uintptr

	// Special case for strings
	if s.expectedKind == reflect.String {
		ptr = stringPointer(got.String())
	} else {
		ptr = got.Pointer()
	}

	if ptr != s.expectedPointer {
		if ctx.BooleanError {
			return ctxerr.BooleanError
		}
		return ctx.CollectError(&ctxerr.Error{
			Message:  fmt.Sprintf("%s pointer mismatch", s.expectedKind),
			Got:      types.RawString(fmt.Sprintf("0x%x", ptr)),
			Expected: types.RawString(fmt.Sprintf("0x%x", s.expectedPointer)),
		})
	}
	return nil
}

func (s *tdShallow) String() string {
	return fmt.Sprintf("(%s) 0x%x", s.expectedKind, s.expectedPointer)
}
