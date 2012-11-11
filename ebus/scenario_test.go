// Tideland Common Go Library - Event Bus - Scenario Unit Tests
//
// Copyright (C) 2010-2012 Frank Mueller / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed 
// by the new BSD license.

package ebus_test

//--------------------
// IMPORTS
//--------------------

import (
	"cgl.tideland.biz/applog"
	"cgl.tideland.biz/asserts"
	"cgl.tideland.biz/config"
	"cgl.tideland.biz/ebus"
	"cgl.tideland.biz/monitoring"
	"fmt"
	"math/rand"
	"testing"
	"time"
)

//--------------------
// ENTITIES
//--------------------

// OrderItem is one item of an order.
type OrderItem struct {
	OrderNo  int
	ItemNo   int
	Quantity int
}

// String returns an order item in a readable form.
func (oi OrderItem) String() string {
	return fmt.Sprintf("item(no: %d / quantity: %d)", oi.ItemNo, oi.Quantity)
}

// Order represents an order.
type Order struct {
	OrderNo    int
	OrderItems map[int]OrderItem
}

// String returns an order in a readable form.
func (o Order) String() string {
	return fmt.Sprintf("order(no: %d / items: %+v)", o.OrderNo, o.OrderItems)
}

// generateOrder creates an order with the given order number.
func generateOrder(orderNo int) Order {
	order := Order{
		OrderNo:    orderNo,
		OrderItems: make(map[int]OrderItem),
	}
	for i := 0; i < rand.Intn(25)+1; i++ {
		orderItem := OrderItem{
			OrderNo:  orderNo,
			ItemNo:   rand.Intn(100000),
			Quantity: rand.Intn(100) + 1,
		}
		order.OrderItems[orderItem.ItemNo] = orderItem
	}
	return order
}

// ShipmentItem is one item of a shipment.
type ShipmentItem struct {
	ShipmentNo int
	ItemNo     int
	Quantity   int
}

// Shipment represents a shipment.
type Shipment struct {
	ShipmentNo    int
	OrderNo       int
	ShipmentItems map[int]ShipmentItem
}

//--------------------
// AGENTS
//--------------------

// ShopAgent fetches new orders. It creates a new OrderAgent for each order.
type ShopAgent struct {
	id  string
	err error
}

func StartShopAgent() {
	a, _ := ebus.Register(&ShopAgent{"Shop", nil})
	ebus.Subscribe(a, "OrderReceived")
}

func (a *ShopAgent) Id() string {
	return a.id
}

func (a *ShopAgent) Process(event ebus.Event) error {
	var order Order
	a.err = event.Payload(&order)
	if a.err != nil {
		return a.err
	}
	a.err = StartOrderAgent(order)
	if a.err != nil {
		return a.err
	}
	return nil
}

func (a *ShopAgent) Recover(r interface{}, event ebus.Event) error {
	applog.Errorf("cannot process event %v: %v", event, r)
	return nil
}

func (a *ShopAgent) Stop() {}

func (a *ShopAgent) Err() error {
	return a.err
}

// OrderAgent is responsible for one order.
type OrderAgent struct {
	id           string
	order        Order
	openItems    map[int]bool
	shippedTopic string
	err          error
}

func StartOrderAgent(order Order) error {
	a := &OrderAgent{
		id:           ebus.Id("Order", order.OrderNo),
		order:        order,
		openItems:    make(map[int]bool),
		shippedTopic: ebus.Id("WarehouseShipped", order.OrderNo),
	}
	_, err := ebus.Register(a)
	if err != nil {
		return err
	}
	ebus.Subscribe(a, a.shippedTopic)
	// Emit a request for each order idem.
	for _, item := range order.OrderItems {
		a.openItems[item.ItemNo] = true
		ebus.Emit(item, "WarehouseItemOrdered")
	}
	return nil
}

func (a *OrderAgent) Id() string {
	return a.id
}

func (a *OrderAgent) Process(event ebus.Event) error {
	var shipment Shipment
	a.err = event.Payload(&shipment)
	if a.err != nil {
		return a.err
	}
	for itemNo := range shipment.ShipmentItems {
		delete(a.openItems, itemNo)
	}
	if len(a.openItems) == 0 {
		// All done, pack for delivery.
		shipment := Shipment{a.order.OrderNo, a.order.OrderNo, make(map[int]ShipmentItem)}
		for itemNo, item := range a.order.OrderItems {
			shipment.ShipmentItems[itemNo] = ShipmentItem{a.order.OrderNo, itemNo, item.Quantity}
		}
		ebus.Emit(shipment, "OrderPacked")
	}
	return nil
}

func (a *OrderAgent) Recover(r interface{}, event ebus.Event) error {
	applog.Errorf("cannot process event %v: %v", event, r)
	return nil
}

func (a *OrderAgent) Stop() {}

func (a *OrderAgent) Err() error {
	return a.err
}

// WarehouseAgent manages the warehouse.
type WarehouseAgent struct {
	id            string
	inventory     map[int]int
	openOrders    map[int]map[int]OrderItem
	shipmentNo    int
	orderedTopic  string
	shippedTopic  string
	preparedTopic string
	err           error
}

func StartWarehouseAgent() error {
	a := &WarehouseAgent{
		id:            ebus.Id("Warehouse"),
		inventory:     make(map[int]int),
		openOrders:    make(map[int]map[int]OrderItem),
		orderedTopic:  "WarehouseItemOrdered",
		shippedTopic:  "ManufacturingShipped",
		preparedTopic: "OrderItemsPrepared",
	}
	_, err := ebus.Register(a)
	if err != nil {
		return err
	}
	ebus.AddTicker("Warehouse", time.Second, a.preparedTopic)
	ebus.Subscribe(a, a.orderedTopic)
	ebus.Subscribe(a, a.shippedTopic)
	ebus.Subscribe(a, a.preparedTopic)
	return nil
}

func (a *WarehouseAgent) Id() string {
	return a.id
}

func (a *WarehouseAgent) Process(event ebus.Event) error {
	switch event.Topic() {
	case a.orderedTopic:
		var item OrderItem
		a.err = event.Payload(&item)
		if a.err != nil {
			return a.err
		}
		// Request item from manufacturer.
		if a.openOrders[item.ItemNo] == nil {
			a.openOrders[item.ItemNo] = make(map[int]OrderItem)
		}
		a.openOrders[item.ItemNo][item.OrderNo] = item
		ebus.Emit(item, "ManufacturerItemOrdered")
	case a.shippedTopic:
		// Manufactured items are shipped, add to inventory.
		var shipment Shipment
		a.err = event.Payload(&shipment)
		if a.err != nil {
			return a.err
		}
		for itemNo, item := range shipment.ShipmentItems {
			a.inventory[itemNo] += item.Quantity
		}
	case a.preparedTopic:
		// Order items are prepared and can now be shipped.
		for itemNo, orderItems := range a.openOrders {
			for orderNo, item := range orderItems {
				if a.inventory[itemNo] >= item.Quantity {
					a.shipmentNo++
					shipment := Shipment{a.shipmentNo, item.OrderNo, make(map[int]ShipmentItem)}
					shipment.ShipmentItems[itemNo] = ShipmentItem{a.shipmentNo, itemNo, item.Quantity}
					ebus.Emit(shipment, "WarehouseShipped", item.OrderNo)
					a.inventory[itemNo] -= item.Quantity
					delete(orderItems, orderNo)
				}
			}
		}
	}
	return nil
}

func (a *WarehouseAgent) Recover(r interface{}, event ebus.Event) error {
	applog.Errorf("cannot process event %v: %v", event, r)
	return nil
}

func (a *WarehouseAgent) Stop() {}

func (a *WarehouseAgent) Err() error {
	return a.err
}

// ManufacturerAgent manufactures group of items.
type ManufacturerAgent struct {
	id                string
	openOrders        map[int]int
	shipmentNo        int
	orderedTopic      string
	manufacturedTopic string
	err               error
}

func StartManufacturerAgent() error {
	a := &ManufacturerAgent{
		id:                ebus.Id("Manufacturer"),
		openOrders:        make(map[int]int),
		orderedTopic:      "ManufacturerItemOrdered",
		manufacturedTopic: "OrderedItemsManufactured",
	}
	_, err := ebus.Register(a)
	if err != nil {
		return err
	}
	ebus.AddTicker("Manufacturing", time.Second, a.manufacturedTopic)
	ebus.Subscribe(a, a.orderedTopic)
	ebus.Subscribe(a, a.manufacturedTopic)
	return nil
}

func (a *ManufacturerAgent) Id() string {
	return a.id
}

func (a *ManufacturerAgent) Process(event ebus.Event) error {
	switch event.Topic() {
	case a.orderedTopic:
		// Store a new manufacturing order.
		var item OrderItem
		a.err = event.Payload(&item)
		if a.err != nil {
			return a.err
		}
		a.openOrders[item.ItemNo] += item.Quantity
	case a.manufacturedTopic:
		// Deliver random numbers of each item.
		a.shipmentNo++
		shipment := Shipment{a.shipmentNo, 0, make(map[int]ShipmentItem)}
		for itemNo, quantity := range a.openOrders {
			shipmentQuantity := rand.Intn(quantity) + 1
			if quantity == shipmentQuantity {
				delete(a.openOrders, itemNo)
			} else {
				a.openOrders[itemNo] = quantity - shipmentQuantity
			}
			shipment.ShipmentItems[itemNo] = ShipmentItem{a.shipmentNo, itemNo, shipmentQuantity}
		}
		ebus.Emit(shipment, "ManufacturingShipped")
	}
	if len(a.openOrders)%10 == 0 {
		applog.Infof("manufacturer: topic %q open orders are %d", event.Topic(), len(a.openOrders))
	}
	return nil
}

func (a *ManufacturerAgent) Recover(r interface{}, event ebus.Event) error {
	applog.Errorf("cannot process event %v: %v", event, r)
	return nil
}

func (a *ManufacturerAgent) Stop() {}

func (a *ManufacturerAgent) Err() error {
	return a.err
}

// DeliveryAgent delivers done orders to the customers.
type DeliveryAgent struct {
	id            string
	openOrders    map[int]bool
	receivedTopic string
	packedTopic   string
	err           error
}

func StartDeliveryAgent() error {
	a := &DeliveryAgent{
		id:            ebus.Id("Delivery"),
		openOrders:    make(map[int]bool),
		receivedTopic: "OrderReceived",
		packedTopic:   "OrderPacked",
	}
	_, err := ebus.Register(a)
	if err != nil {
		return err
	}
	ebus.Subscribe(a, a.receivedTopic)
	ebus.Subscribe(a, a.packedTopic)
	return nil
}

func (a *DeliveryAgent) Id() string {
	return a.id
}

func (a *DeliveryAgent) Process(event ebus.Event) error {
	switch event.Topic() {
	case a.receivedTopic:
		// Received a new order.
		var order Order
		a.err = event.Payload(&order)
		if a.err != nil {
			return a.err
		}
		a.openOrders[order.OrderNo] = true
	case a.packedTopic:
		// Deliver the packed order.
		var shipment Shipment
		a.err = event.Payload(&shipment)
		if a.err != nil {
			return a.err
		}
		delete(a.openOrders, shipment.OrderNo)
		ebus.Emit(len(a.openOrders), "OrderDelivered")
	}
	if len(a.openOrders)%10 == 0 {
		applog.Infof("delivery: topic %q open orders are %d", event.Topic(), len(a.openOrders))
	}
	return nil
}

func (a *DeliveryAgent) Recover(r interface{}, event ebus.Event) error {
	applog.Errorf("cannot process event %v: %v", event, r)
	return nil
}

func (a *DeliveryAgent) Stop() {}

func (a *DeliveryAgent) Err() error {
	return a.err
}

// WaitAgent waits that everything is done.
type WaitAgent struct {
	id             string
	done           chan bool
	deliveredTopic string
	err            error
}

func StartWaitAgent(done chan bool) error {
	a := &WaitAgent{
		id:             ebus.Id("Wait"),
		done:           done,
		deliveredTopic: "OrderDelivered",
	}
	_, err := ebus.Register(a)
	if err != nil {
		return err
	}
	ebus.Subscribe(a, a.deliveredTopic)
	return nil
}

func (a *WaitAgent) Id() string {
	return a.id
}

func (a *WaitAgent) Process(event ebus.Event) error {
	var rest int
	a.err = event.Payload(&rest)
	if a.err != nil {
		return a.err
	}
	applog.Infof("wait: %d orders remaining", rest)
	if rest == 0 {
		a.done <- true
	}
	return nil
}

func (a *WaitAgent) Recover(r interface{}, event ebus.Event) error {
	applog.Errorf("cannot process event %v: %v", event, r)
	return nil
}

func (a *WaitAgent) Stop() {}

func (a *WaitAgent) Err() error {
	return a.err
}

//--------------------
// HELPERS
//--------------------

// scenarioParam describes the parameters for a scenario test.
type scenarioParam struct {
	Id      string
	Orders  int
	Timeout time.Duration
}

func (p scenarioParam) String() string {
	return fmt.Sprintf("scenario %q with %d orders", p.Id, p.Orders)
}

// runScenarioTest runs a test with the given number of orders
func runScenarioTest(assert *asserts.Asserts, param scenarioParam) {
	applog.Infof(param.String())
	monitoring.Reset()

	// Prepare test.
	done := make(chan bool)
	provider := config.NewMapConfigurationProvider()
	config := config.New(provider)

	config.Set("backend", "single")

	assert.Nil(ebus.Init(config), "single node backend started")

	StartShopAgent()
	StartWarehouseAgent()
	StartManufacturerAgent()
	StartDeliveryAgent()
	StartWaitAgent(done)

	// Run orders.
	for on := 0; on < param.Orders; on++ {
		order := generateOrder(on)
		err := ebus.Emit(order, "OrderReceived")
		assert.Nil(err, "order emitted")
	}

	select {
	case <-done:
		applog.Infof("order processing done")
	case <-time.After(param.Timeout):
		assert.Fail("timeout during wait for processed orders")
	}

	// Finalize test.
	err := ebus.Stop()
	assert.Nil(err, "stopped the bus")
	time.Sleep(time.Second)
	monitoring.MeasuringPointsPrintAll()
}

//--------------------
// TESTS
//--------------------

var scenarioTestParams = []scenarioParam{
	{"scenario-s", 1000, 30 * time.Second},
	// {"scenario-m", 10000, 300 * time.Second},
	// {"scenario-l", 100000, 1000 * time.Second},
}

func TestScenarios(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	for _, param := range scenarioTestParams {
		runScenarioTest(assert, param)
	}
}

// EOF
