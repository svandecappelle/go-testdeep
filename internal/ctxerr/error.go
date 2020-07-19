// Copyright (c) 2018, Maxime Soulé
// All rights reserved.
//
// This source code is licensed under the BSD-style license found in the
// LICENSE file in the root directory of this source tree.

package ctxerr

import (
	"bytes"
	"strings"

	"github.com/maxatome/go-testdeep/internal/color"
	"github.com/maxatome/go-testdeep/internal/location"
	"github.com/maxatome/go-testdeep/internal/util"
)

// Error represents errors generated by td (go-testdeep) functions.
type Error struct {
	// Context when the error occurred
	Context Context
	// Message describes the error
	Message string
	// Got value
	Got interface{}
	// Expected value
	Expected interface{}
	// If not nil, Summary is used to display summary instead of using
	// Got + Expected fields
	Summary ErrorSummary
	// If initialized, location of TestDeep operator originator of the error
	Location location.Location
	// If defined, the current Error comes from this Error
	Origin *Error
	// If defined, points to the next Error
	Next *Error
}

var (
	// BooleanError is the *Error returned when an error occurs in a
	// boolean context.
	BooleanError = &Error{}

	// ErrTooManyErrors is chained to the last error encountered when
	// the maximum number of errors has been reached.
	ErrTooManyErrors = &Error{
		Message: "Too many errors (use TESTDEEP_MAX_ERRORS=-1 to see all)",
	}
)

// Error implements error interface.
func (e *Error) Error() string {
	buf := bytes.Buffer{}

	e.Append(&buf, "")

	return buf.String()
}

// Append appends the Error contents to "buf" using prefix "prefix"
// for each line.
func (e *Error) Append(buf *bytes.Buffer, prefix string) {
	if e == BooleanError {
		return
	}

	color.Init()

	var writeEolPrefix func()
	if prefix != "" {
		eolPrefix := make([]byte, 1+len(prefix))
		eolPrefix[0] = '\n'
		copy(eolPrefix[1:], prefix)

		writeEolPrefix = func() {
			buf.Write(eolPrefix)
		}
		buf.WriteString(prefix)
	} else {
		writeEolPrefix = func() {
			buf.WriteByte('\n')
		}
	}

	if e == ErrTooManyErrors {
		buf.WriteString(color.TitleOn)
		buf.WriteString(e.Message)
		buf.WriteString(color.TitleOff)
		return
	}

	buf.WriteString(color.TitleOn)
	if pos := strings.Index(e.Message, "%%"); pos >= 0 {
		buf.WriteString(e.Message[:pos])
		buf.WriteString(e.Context.Path.String())
		buf.WriteString(e.Message[pos+2:])
	} else {
		buf.WriteString(e.Context.Path.String())
		buf.WriteString(": ")
		buf.WriteString(e.Message)
	}
	buf.WriteString(color.TitleOff)

	if e.Summary != nil {
		buf.WriteByte('\n')
		e.Summary.AppendSummary(buf, prefix+"\t")
	} else {
		writeEolPrefix()
		buf.WriteString(color.BadOnBold)
		buf.WriteString("\t     got: ")
		buf.WriteString(color.BadOn)
		util.IndentStringIn(buf, e.GotString(), prefix+"\t          ")
		buf.WriteString(color.BadOff)
		writeEolPrefix()
		buf.WriteString(color.OKOnBold)
		buf.WriteString("\texpected: ")
		buf.WriteString(color.OKOn)
		util.IndentStringIn(buf, e.ExpectedString(), prefix+"\t          ")
		buf.WriteString(color.OKOff)
	}

	// This error comes from another one
	if e.Origin != nil {
		writeEolPrefix()
		buf.WriteString("Originates from following error:\n")

		e.Origin.Append(buf, prefix+"\t")
	}

	if e.Location.IsInitialized() &&
		!e.Location.BehindCmp && // no need to log Cmp* func
		(e.Next == nil || e.Next.Location != e.Location) {
		writeEolPrefix()
		buf.WriteString("[under TestDeep operator ")
		buf.WriteString(e.Location.String())
		buf.WriteByte(']')
	}

	if e.Next != nil {
		buf.WriteByte('\n')
		e.Next.Append(buf, prefix) // next error at same level
	}
}

// GotString returns the string corresponding to the Got
// field. Returns the empty string if the Error Summary field is not
// nil.
func (e *Error) GotString() string {
	if e.Summary != nil {
		return ""
	}
	return util.ToString(e.Got)
}

// ExpectedString returns the string corresponding to the Expected
// field. Returns the empty string if the Error Summary field is not
// nil.
func (e *Error) ExpectedString() string {
	if e.Summary != nil {
		return ""
	}
	return util.ToString(e.Expected)
}

// SummaryString returns the string corresponding to the Summary
// field. Returns the empty string if the Error Summary field is nil.
func (e *Error) SummaryString() string {
	if e.Summary == nil {
		return ""
	}

	var buf bytes.Buffer
	e.Summary.AppendSummary(&buf, "")
	return buf.String()
}
