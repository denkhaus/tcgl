// Tideland Common Go Library - Numerics - Unit Tests
//
// Copyright (C) 2009-2011 Frank Mueller / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed 
// by the new BSD license.

package numerics

//--------------------
// IMPORTS
//--------------------

import (
	"sort"
	"testing"
)

//--------------------
// TESTS
//--------------------

// Test simple point.
func TestSimplePoint(t *testing.T) {
	pa := NewPoint(1.0, 5.0)
	pb := NewPoint(2.0, 2.0)
	pc := NewPoint(3.0, 4.0)

	if pa.X() != 1.0 {
		t.Errorf("A.X is not 1.0!")
	}
	if pa.Y() != 5.0 {
		t.Errorf("A.Y is not 5.0!")
	}

	t.Logf("Distance B/C    : %f", pb.DistanceTo(pc))
	t.Logf("Middle Point B/C: %v", MiddlePoint(pb, pc))
	t.Logf("Vector B/C      : %v", PointVector(pb, pc))
}

// Test simple point array.
func TestSimplePointArray(t *testing.T) {
	ps := NewPoints(5)

	t.Logf("Points (before): %v", ps)
	t.Logf("Length (before): %v", ps.Len())

	ps.AppendPoint(2.0, 2.0)
	ps.AppendPoint(5.0, 1.0)
	ps.AppendPoint(4.0, 2.0)
	ps.AppendPoint(3.0, 3.0)
	ps.AppendPoint(1.0, 1.0)

	t.Logf("Points (after): %v", ps)
	t.Logf("Length (after): %v", ps.Len())

	sort.Sort(ps)

	t.Logf("Points (sorted): %v", ps)
	t.Logf("Length (sorted): %v", ps.Len())

	t.Logf("Point 3: %v", ps.At(3))

	if ps.At(3).X() != 4.0 {
		t.Errorf("Point 3 is not (4.0, 2.0)!")
	}
}

// Test simple polynomial function.
func TestSimplePolynomialFunction(t *testing.T) {
	p := NewPolynomialFunction([]float64{2.0, 2.0})

	ya := p.Eval(-2.0)
	yb := p.Eval(-1.0)
	yc := p.Eval(0.0)
	yd := p.Eval(2.0)

	t.Logf("A: %f / B: %f / C: %f / D: %f", ya, yb, yc, yd)

	if ya != -2.0 {
		t.Errorf("A is not -2.0!")
	}
	if yb != 0.0 {
		t.Errorf("B is not 0.0!")
	}
	if yc != 2.0 {
		t.Errorf("C is not 2.0!")
	}
	if yd != 6.0 {
		t.Errorf("D is not 6.0!")
	}
}

// Test polynomial function printing.
func TestPolynomialFunctionPrinting(t *testing.T) {
	p := NewPolynomialFunction([]float64{-7.55, 2.0, -3.1, 2.66, -3.45})

	t.Logf("Function is %v", p)
}

// Test quadratic polynomial function.
func TestQuadraticPolynomialFunction(t *testing.T) {
	p := NewPolynomialFunction([]float64{0.0, 0.0, 1.0})

	ya := p.Eval(-1.0)
	yb := p.Eval(2.0)
	yc := p.Eval(-3.0)

	t.Logf("A: %f / B: %f / C: %f", ya, yb, yc)

	if ya != 1.0 {
		t.Errorf("A is not 1.0!")
	}
	if yb != 4.0 {
		t.Errorf("B is not 4.0!")
	}
	if yc != 9.0 {
		t.Errorf("C is not 9.0!")
	}
}

// Test polynomial function differentiation.
func TestPolynomialFunctionDifferentiation(t *testing.T) {
}

// Test interpolation
func TestInterpolation(t *testing.T) {
	ps := NewPoints(5)

	ps.AppendPoint(1.0, 1.0)
	ps.AppendPoint(2.0, 2.0)
	ps.AppendPoint(3.0, 3.0)
	ps.AppendPoint(4.0, 2.0)
	ps.AppendPoint(5.0, 1.0)

	f := NewCubicSplineFunction(ps)

	t.Logf("X/Y 1: %v", f.EvalPoint(3.5))
}

// Test points evaluation.
func TestPointsEvaluation(t *testing.T) {
	ps := NewPoints(10)

	ps.AppendPoint(0.0, 0.7)
	ps.AppendPoint(1.0, 1.1)
	ps.AppendPoint(2.0, 0.0)
	ps.AppendPoint(3.0, -0.5)
	ps.AppendPoint(4.0, -2.0)
	ps.AppendPoint(5.0, -1.0)
	ps.AppendPoint(6.0, 0.2)
	ps.AppendPoint(7.0, 0.3)
	ps.AppendPoint(8.0, -0.4)
	ps.AppendPoint(9.0, -0.5)

	lsf := ps.CubicSplineFunction().EvalPoints(0.7, 8.1, 50).LeastSquaresFunction()

	y := lsf.Eval(15.0)

	t.Logf("Y: %f", y)
}

// Test least squares function.
func TestLeastSquaresFunction(t *testing.T) {
	lsf := NewLeastSquaresFunction(nil)

	lsf.AppendPoint(1.0, 1.0)
	lsf.AppendPoint(2.0, 0.5)
	lsf.AppendPoint(3.0, 2.0)
	lsf.AppendPoint(4.0, 2.5)
	lsf.AppendPoint(5.0, 1.5)
	lsf.AppendPoint(6.0, 1.0)
	lsf.AppendPoint(7.0, 1.5)

	ya := lsf.Eval(9.0)
	yb := lsf.Eval(4.5)

	t.Logf("YA: %f / YB: %f", ya, yb)
}

// EOF
