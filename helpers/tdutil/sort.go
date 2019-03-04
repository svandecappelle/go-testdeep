// Copyright (c) 2019, Maxime Soulé
// All rights reserved.
//
// This source code is licensed under the BSD-style license found in the
// LICENSE file in the root directory of this source tree.

package tdutil

import (
	"math"
	"reflect"
	"sort"

	"github.com/maxatome/go-testdeep/internal/visited"
)

// SortableValues is used to allow the sorting of a []reflect.Value
// slice. It is used with the standard sort package:
//
//   vals := []reflect.Value{a, b, c, d}
//   sort.Sort(SortableValues(vals))
//   // vals contents now sorted
//
// Replace sort.Sort by sort.Stable for a stable sort. See sort documentation.
//
// Sorting rules are as follows:
//   - nil is always lower
//   - different types are sorted by their name
//   - false is lesser than true
//   - float and int numbers are sorted by their value
//   - complex numbers are sorted by their real, then by their imaginary parts
//   - strings are sorted by their value
//   - map: shorter length is lesser, then sorted by address
//   - functions, channels and unsafe pointer are sorted by their address
//   - struct: comparison is spread to each field
//   - pointer: comparison is spred to the pointed value
//   - arrays: comparison is spread to each item
//   - slice: comparison is spread to each item, then shorter length is lesser
//   - interface: comparison is spread to the value
//
// Cyclic references are correctly handled.
func SortableValues(s []reflect.Value) sort.Interface {
	r := &rValues{
		Slice: s,
	}
	if len(s) > 1 {
		r.Visited = visited.NewVisited()
	}
	return r
}

type rValues struct {
	Visited visited.Visited
	Slice   []reflect.Value
}

func (v *rValues) Len() int {
	return len(v.Slice)
}

func (v *rValues) Less(i, j int) bool {
	return cmp(v.Visited, v.Slice[i], v.Slice[j]) < 0
}

func (v *rValues) Swap(i, j int) {
	v.Slice[i], v.Slice[j] = v.Slice[j], v.Slice[i]
}

func cmpRet(less, gt bool) int {
	if less {
		return -1
	}
	if gt {
		return 1
	}
	return 0
}

func cmpFloat(a, b float64) int {
	if math.IsNaN(a) {
		return -1
	}
	if math.IsNaN(b) {
		return 1
	}
	return cmpRet(a < b, a > b)
}

// cmp returns -1 if a < b, 1 if a > b, 0 if a == b.
func cmp(v visited.Visited, a, b reflect.Value) int {
	if !a.IsValid() {
		if !b.IsValid() {
			return 0
		}
		return -1
	}
	if !b.IsValid() {
		return 1
	}

	if at, bt := a.Type(), b.Type(); at != bt {
		sat, sbt := at.String(), bt.String()
		return cmpRet(sat < sbt, sat > sbt)
	}

	// Avoid looping forever on cyclic references
	if v.Record(a, b) {
		return 0
	}

	switch a.Kind() {
	case reflect.Bool:
		if a.Bool() {
			if b.Bool() {
				return 0
			}
			return 1
		}
		if b.Bool() {
			return -1
		}
		return 0

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		na, nb := a.Int(), b.Int()
		return cmpRet(na < nb, na > nb)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32,
		reflect.Uint64, reflect.Uintptr:
		na, nb := a.Uint(), b.Uint()
		return cmpRet(na < nb, na > nb)

	case reflect.Float32, reflect.Float64:
		return cmpFloat(a.Float(), b.Float())

	case reflect.Complex64, reflect.Complex128:
		na, nb := a.Complex(), b.Complex()
		fa, fb := real(na), real(nb)
		if r := cmpFloat(fa, fb); r != 0 {
			return r
		}
		return cmpFloat(imag(na), imag(nb))

	case reflect.String:
		sa, sb := a.String(), b.String()
		return cmpRet(sa < sb, sa > sb)

	case reflect.Array:
		for i := 0; i < a.Len(); i++ {
			if r := cmp(v, a.Index(i), b.Index(i)); r != 0 {
				return r
			}
		}
		return 0

	case reflect.Slice:
		al, bl := a.Len(), b.Len()
		maxl := al
		if al > bl {
			maxl = bl
		}
		for i := 0; i < maxl; i++ {
			if r := cmp(v, a.Index(i), b.Index(i)); r != 0 {
				return r
			}
		}
		return cmpRet(al < bl, al > bl)

	case reflect.Interface:
		if a.IsNil() {
			if b.IsNil() {
				return 0
			}
			return -1
		}
		if b.IsNil() {
			return 1
		}
		return cmp(v, a.Elem(), b.Elem())

	case reflect.Struct:
		for i, m := 0, a.NumField(); i < m; i++ {
			if r := cmp(v, a.Field(i), b.Field(i)); r != 0 {
				return r
			}
		}
		return 0

	case reflect.Ptr:
		if a.Pointer() == b.Pointer() {
			return 0
		}
		if a.IsNil() {
			return -1
		}
		if b.IsNil() {
			return 1
		}
		return cmp(v, a.Elem(), b.Elem())

	case reflect.Map:
		// consider shorter maps are before longer ones
		al, bl := a.Len(), b.Len()
		if r := cmpRet(al < bl, al > bl); r != 0 {
			return r
		}
		// then fallback on pointers comparison. How to say a map is
		// before another one otherwise?
		fallthrough

	case reflect.Func, reflect.Chan, reflect.UnsafePointer:
		pa, pb := a.Pointer(), b.Pointer()
		return cmpRet(pa < pb, pa > pb)

	default:
		panic("don't know how to compare " + a.Kind().String())
	}
}