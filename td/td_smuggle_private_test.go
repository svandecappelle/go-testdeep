// Copyright (c) 2021, Maxime Soulé
// All rights reserved.
//
// This source code is licensed under the BSD-style license found in the
// LICENSE file in the root directory of this source tree.

package td

import (
	"reflect"
	"testing"

	"github.com/maxatome/go-testdeep/internal/test"
)

func TestFieldPath(t *testing.T) {
	check := func(in string, expected ...string) []smuggleField {
		t.Helper()

		got, err := splitFieldPath(in)
		test.NoError(t, err)

		var gotStr []string
		for _, s := range got {
			gotStr = append(gotStr, s.Name)
		}

		if !reflect.DeepEqual(gotStr, expected) {
			t.Errorf("Failed:\n       got: %v\n  expected: %v", got, expected)
		}

		test.EqualStr(t, in, joinFieldPath(got))

		return got
	}

	check("test", "test")
	check("test.foo.bar", "test", "foo", "bar")
	check("test.foo.bar", "test", "foo", "bar")
	check("test[foo.bar]", "test", "foo.bar")
	check("test[foo][bar]", "test", "foo", "bar")
	fp := check("test[foo][bar].zip", "test", "foo", "bar", "zip")

	// "." can be omitted just after "]"
	got, err := splitFieldPath("test[foo][bar]zip")
	test.NoError(t, err)
	if !reflect.DeepEqual(got, fp) {
		t.Errorf("Failed:\n       got: %v\n  expected: %v", got, fp)
	}

	//
	// Errors
	checkErr := func(in, expectedErr string) {
		t.Helper()

		_, err := splitFieldPath(in)

		if test.Error(t, err) {
			test.EqualStr(t, err.Error(), expectedErr)
		}
	}

	checkErr("", "FIELD_PATH cannot be empty")
	checkErr("foo[bar", `cannot find final ']' in FIELD_PATH "foo[bar"`)
	checkErr("test.%foo", `unexpected '%' in field name "%foo" in FIELDS_PATH "test.%foo"`)
	checkErr("test.f%oo", `unexpected '%' in field name "f%oo" in FIELDS_PATH "test.f%oo"`)
}

func TestBuildFieldPathFn(t *testing.T) {
	_, err := buildFieldPathFn("bad[path")
	test.Error(t, err)

	//
	// Struct
	type Build struct {
		Field struct {
			Path string
		}
		Iface interface{}
	}

	fn, err := buildFieldPathFn("Field.Path.Bad")
	if test.NoError(t, err) {
		_, err = fn(Build{})
		if test.Error(t, err) {
			test.EqualStr(t, err.Error(),
				`field "Field.Path" is a string and should be a struct`)
		}

		_, err = fn(123)
		if test.Error(t, err) {
			test.EqualStr(t, err.Error(), "it is a int and should be a struct")
		}
	}

	fn, err = buildFieldPathFn("Field.Unknown")
	if test.NoError(t, err) {
		_, err = fn(Build{})
		if test.Error(t, err) {
			test.EqualStr(t, err.Error(), `field "Field.Unknown" not found`)
		}
	}

	//
	// Map
	fn, err = buildFieldPathFn("Iface[str].Field")
	if test.NoError(t, err) {
		_, err = fn(Build{Iface: map[int]Build{}})
		if test.Error(t, err) {
			test.EqualStr(t, err.Error(),
				`field "Iface[str]", "str" is not an integer and so cannot match int map key type`)
		}

		_, err = fn(Build{Iface: map[uint]Build{}})
		if test.Error(t, err) {
			test.EqualStr(t, err.Error(),
				`field "Iface[str]", "str" is not an unsigned integer and so cannot match uint map key type`)
		}

		_, err = fn(Build{Iface: map[float32]Build{}})
		if test.Error(t, err) {
			test.EqualStr(t, err.Error(),
				`field "Iface[str]", "str" is not a float and so cannot match float32 map key type`)
		}

		_, err = fn(Build{Iface: map[complex128]Build{}})
		if test.Error(t, err) {
			test.EqualStr(t, err.Error(),
				`field "Iface[str]", "str" is not a complex number and so cannot match complex128 map key type`)
		}

		_, err = fn(Build{Iface: map[struct{ A int }]Build{}})
		if test.Error(t, err) {
			test.EqualStr(t, err.Error(),
				`field "Iface[str]", "str" cannot match unsupported struct { A int } map key type`)
		}

		_, err = fn(Build{Iface: map[string]Build{}})
		if test.Error(t, err) {
			test.EqualStr(t, err.Error(), `field "Iface[str]", "str" map key not found`)
		}
	}

	//
	// Array / Slice
	fn, err = buildFieldPathFn("Iface[str].Field")
	if test.NoError(t, err) {
		_, err = fn(Build{Iface: []int{}})
		if test.Error(t, err) {
			test.EqualStr(t, err.Error(),
				`field "Iface[str]", "str" is not a slice/array index`)
		}
	}

	fn, err = buildFieldPathFn("Iface[18].Field")
	if test.NoError(t, err) {
		_, err = fn(Build{Iface: []int{1, 2, 3}})
		if test.Error(t, err) {
			test.EqualStr(t, err.Error(),
				`field "Iface[18]", 18 is out of slice/array range (len 3)`)
		}

		_, err = fn(Build{Iface: 42})
		if test.Error(t, err) {
			test.EqualStr(t, err.Error(),
				`field "Iface" is a int, but a map, array or slice is expected`)
		}
	}

	fn, err = buildFieldPathFn("[18].Field")
	if test.NoError(t, err) {
		_, err = fn(42)
		test.EqualStr(t, err.Error(),
			`it is a int, but a map, array or slice is expected`)
	}
}
