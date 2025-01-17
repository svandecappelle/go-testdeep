// Copyright (c) 2020, Maxime Soulé
// All rights reserved.
//
// This source code is licensed under the BSD-style license found in the
// LICENSE file in the root directory of this source tree.

package td

import (
	"reflect"
	"sync"

	"github.com/maxatome/go-testdeep/internal/color"
	"github.com/maxatome/go-testdeep/internal/ctxerr"
)

type tdDelay struct {
	base
	operator TestDeep
	once     sync.Once
	delayed  func() TestDeep
}

var _ TestDeep = &tdDelay{}

// summary(Delay): delays the operator construction till first use
// input(Delay): all

// Delay operator allows to delay the construction of an operator to
// the time it is used for the first time. Most of the time, it is
// used with helpers. See the example for a very simple use case.
func Delay(delayed func() TestDeep) TestDeep {
	if delayed == nil {
		panic(color.Bad("Delay(DELAYED): DELAYED must be non-nil"))
	}
	return &tdDelay{
		base:    newBase(3),
		delayed: delayed,
	}
}

func (d *tdDelay) Match(ctx ctxerr.Context, got reflect.Value) *ctxerr.Error {
	op := d.getOperator()
	ctx.CurOperator = op // to have correct location
	return op.Match(ctx, got)
}

func (d *tdDelay) String() string {
	return d.getOperator().String()
}

func (d *tdDelay) TypeBehind() reflect.Type {
	return d.getOperator().TypeBehind()
}

func (d *tdDelay) HandleInvalid() bool {
	return d.getOperator().HandleInvalid()
}

func (d *tdDelay) getOperator() TestDeep {
	d.once.Do(func() { d.operator = d.delayed() })
	return d.operator
}
