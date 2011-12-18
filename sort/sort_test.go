// Tideland Common Go Library - Sort - Unit Tests
//
// Copyright (C) 2009-2011 Frank Mueller / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed 
// by the new BSD license.

package sort

//--------------------
// IMPORTS
//--------------------

import (
	"fmt"
	"rand"
	"runtime"
	"sort"
	"testing"
	"time"
	"code.google.com/p/tcgl/identifier"
)

//--------------------
// TESTS
//--------------------

// Test pivot.
func TestPivot(t *testing.T) {
	a := make(sort.IntSlice, 15)

	for i := 0; i < len(a); i++ {
		a[i] = rand.Intn(99)
	}

	plo, phi := partition(a, 0, len(a)-1)

	t.Logf("PLO  : %v", plo)
	t.Logf("PHI  : %v", phi)
	t.Logf("PDATA: %v", a[phi-1])
	t.Logf("PIVOT: %v", a)
}

// Test sort shootout.
func TestSort(t *testing.T) {
	ola := generateTestOrders(25000)
	olb := generateTestOrders(25000)
	olc := generateTestOrders(25000)
	old := generateTestOrders(25000)

	ta := time.Nanoseconds()
	Sort(ola)
	tb := time.Nanoseconds()
	sort.Sort(olb)
	tc := time.Nanoseconds()
	insertionSort(olc, 0, len(olc)-1)
	td := time.Nanoseconds()
	sequentialQuickSort(old, 0, len(olc)-1)
	te := time.Nanoseconds()

	t.Logf("PQS: %v", tb-ta)
	t.Logf(" QS: %v", tc-tb)
	t.Logf(" IS: %v", td-tc)
	t.Logf("SQS: %v", te-td)
}

// Test the parallel quicksort function.
func TestParallelQuickSort(t *testing.T) {
	t.Logf("PQS MaxProcs: %v", runtime.GOMAXPROCS(0))

	ol := generateTestOrders(1000000)

	Sort(ol)

	cn := 0

	for _, o := range ol {
		if cn > o.CustomerNo {
			t.Errorf("Customer No %v in wrong order!", o.CustomerNo)

			cn = o.CustomerNo
		} else {
			cn = o.CustomerNo
		}
	}
}

//--------------------
// HELPERS
//--------------------

// Order item type.
type OrderItem struct {
	ArticleNo    int
	Count        int
	UnitPrice    float64
	DiscountPerc float64
}

// Order type.
type Order struct {
	OrderNo    identifier.UUID
	CustomerNo int
	Items      []*OrderItem
}

func (o *Order) String() string {
	msg := "ON: %v / CN: %4v / I: %v"

	return fmt.Sprintf(msg, o.OrderNo, o.CustomerNo, len(o.Items))
}

// Order list.
type OrderList []*Order

func (l OrderList) Len() int {
	return len(l)
}

func (l OrderList) Less(i, j int) bool {
	return l[i].CustomerNo < l[j].CustomerNo
}

func (l OrderList) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

// generateTestOrders produces a list of random orders.
func generateTestOrders(count int) OrderList {
	articleMaxNo := 9999
	unitPrices := make([]float64, articleMaxNo+1)
	orderList := make(OrderList, count)

	for i := 0; i < articleMaxNo+1; i++ {
		unitPrices[i] = rand.Float64() * 100.0
	}

	for i := 0; i < count; i++ {
		order := new(Order)

		order.OrderNo = identifier.NewUUID()
		order.CustomerNo = rand.Intn(999) + 1
		order.Items = make([]*OrderItem, rand.Intn(9)+1)

		for j := 0; j < len(order.Items); j++ {
			articleNo := rand.Intn(articleMaxNo)

			order.Items[j] = &OrderItem{
				ArticleNo:    articleNo,
				Count:        rand.Intn(9) + 1,
				UnitPrice:    unitPrices[articleNo],
				DiscountPerc: rand.Float64() * 4.0,
			}
		}

		orderList[i] = order
	}

	return orderList
}

// EOF
