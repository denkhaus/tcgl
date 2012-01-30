// Tideland Common Go Library - Map/Reduce - Unit Tests
//
// Copyright (C) 2009-2012 Frank Mueller / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed 
// by the new BSD license.

package mapreduce

//--------------------
// IMPORTS
//--------------------

import (
	"code.google.com/p/tcgl/identifier"
	"fmt"
	"math/rand"
	"strconv"
	"testing"
)

//--------------------
// TESTS
//--------------------

// Test the MapReduce function.
func TestMapReduce(t *testing.T) {
	// Start data producer.
	orderChan := generateTestOrders(200000)

	// Define map and reduce functions.
	mapFunc := func(in *KeyValue, mapEmitChan KeyValueChan) {
		o := in.Value.(*Order)

		// Emit analysis data for each item.

		for _, i := range o.Items {
			unitDiscount := (i.UnitPrice / 100.0) * i.DiscountPerc
			totalDiscount := unitDiscount * float64(i.Count)
			totalAmount := (i.UnitPrice - unitDiscount) * float64(i.Count)
			analysis := &OrderItemAnalysis{i.ArticleNo, i.Count, totalAmount, totalDiscount}
			articleNo := strconv.Itoa(i.ArticleNo)

			mapEmitChan <- &KeyValue{articleNo, analysis}
		}
	}

	reduceFunc := func(inChan KeyValueChan, reduceEmitChan KeyValueChan) {
		memory := make(map[string]*OrderItemAnalysis)

		// Collect emitted analysis data.
		for kv := range inChan {
			analysis := kv.Value.(*OrderItemAnalysis)

			if existing, ok := memory[kv.Key]; ok {
				existing.Quantity += analysis.Quantity
				existing.Amount += analysis.Amount
				existing.Discount += analysis.Discount
			} else {
				memory[kv.Key] = analysis
			}
		}

		// Emit it to map/reduce caller.
		for articleNo, analysis := range memory {
			reduceEmitChan <- &KeyValue{articleNo, analysis}
		}
	}

	// Now call MapReduce.
	for result := range SortedMapReduce(orderChan, mapFunc, 100, reduceFunc, 20, KeyLessFunc) {
		t.Logf("%v\n", result.Value)
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

// Order item analysis type.
type OrderItemAnalysis struct {
	ArticleNo int
	Quantity  int
	Amount    float64
	Discount  float64
}

func (oia *OrderItemAnalysis) String() string {
	msg := "AN: %5v / Q: %4v / A: %10.2f € / D: %10.2f €"

	return fmt.Sprintf(msg, oia.ArticleNo, oia.Quantity, oia.Amount, oia.Discount)
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

// Generate test order and push them into a channel.
func generateTestOrders(count int) KeyValueChan {
	articleMaxNo := 9999
	unitPrices := make([]float64, articleMaxNo+1)

	for i := 0; i < articleMaxNo+1; i++ {
		unitPrices[i] = rand.Float64() * 100.0
	}

	kvc := make(KeyValueChan)

	go func() {
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

			kvc <- &KeyValue{order.OrderNo.String(), order}
		}

		close(kvc)
	}()

	return kvc
}

// EOF
