// Tideland Common Go Library - Cells - Scenario Unit Tests
//
// Copyright (C) 2010-2012 Frank Mueller / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed 
// by the new BSD license.

package cells

//--------------------
// IMPORTS
//--------------------

import (
	"cgl.tideland.biz/applog"
	"cgl.tideland.biz/asserts"
	"cgl.tideland.biz/monitoring"
	"fmt"
	"math/rand"
	"testing"
	"time"
)

//--------------------
// ENTITIES
//--------------------

// Order represents an order.
type Order struct {
	OrderNo    int
	OrderItems map[int]*OrderItem
}

// String returns on order in a readable form.
func (o *Order) String() string {
	return fmt.Sprintf("order(no: %d / items: %+v)", o.OrderNo, o.OrderItems)
}

// OrderItem is one item of an order.
type OrderItem struct {
	OrderNo  int
	ItemNo   int
	Quantity int
}

// generateOrder creates an order with the given order number.
func generateOrder(orderNo, items int) *Order {
	order := &Order{
		OrderNo:    orderNo,
		OrderItems: make(map[int]*OrderItem),
	}
	for i := 0; i < rand.Intn(20)+1; i++ {
		orderItem := &OrderItem{
			OrderNo:  orderNo,
			ItemNo:   rand.Intn(items),
			Quantity: rand.Intn(items/100) + 1,
		}
		order.OrderItems[orderItem.ItemNo] = orderItem
	}
	return order
}

// Shipment represents a shipment of items by the manufacturer or the shop.
type Shipment struct {
	ShipmentNo    int
	ShipmentItems map[int]*ShipmentItem
}

// ShipmentItem is one item of a shipment.
type ShipmentItem struct {
	ShipmentNo int
	ItemNo     int
	Quantity   int
}

//--------------------
// SHOP BEHAVIOR
//--------------------

// shopBehavior is the entry point for new orders
// by customers.
type shopBehavior struct {
	env *Environment
	oas *OrdersAndShipments
}

// NewShopBehaviorFactory is the factory for a shop behavior.
func NewShopBehaviorFactory(oas *OrdersAndShipments) BehaviorFactory {
	return func() Behavior { return &shopBehavior{nil, oas} }
}

// Init the behavior.
func (b *shopBehavior) Init(env *Environment, id Id) error {
	b.env = env
	return nil
}

// ProcessEvent processes an event.
func (b *shopBehavior) ProcessEvent(e Event, emitter EventEmitter) {
	switch e.Topic() {
	case "order":
		// A new order has been placed.
		order := e.Payload().(*Order)
		b.oas.OrderChan <- order
		orderCellId := NewId("order", order.OrderNo)
		b.env.AddCell(orderCellId, NewOrderBehaviorFactory(order))
		b.env.Subscribe(orderCellId, "distribution")
		b.env.EmitSimple(orderCellId, "order", true)
	}
}

// Recover from an error.
func (b *shopBehavior) Recover(err interface{}, e Event) {}

// Stop the behavior.
func (b *shopBehavior) Stop() {}

//--------------------
// DISTRIBUTION BEHAVIOR
//--------------------

// distributionBehavior represents a center for the distribution
// of an order to a customer.
type distributionBehavior struct {
	oas *OrdersAndShipments
}

// NewDistributionBehaviorFactory creates a factory for a distribution behavior.
func NewDistributionBehaviorFactory(oas *OrdersAndShipments) BehaviorFactory {
	return func() Behavior { return &distributionBehavior{oas} }
}

// Init the behavior.
func (b *distributionBehavior) Init(env *Environment, id Id) error {
	return nil
}

// ProcessEvent processes an event.
func (b *distributionBehavior) ProcessEvent(e Event, emitter EventEmitter) {
	switch e.Topic() {
	case "shipment":
		// A shipment has been ordered.
		shipment := e.Payload().(*Shipment)
		b.oas.ShipmentChan <- shipment
	}
}

// Recover from an error.
func (b *distributionBehavior) Recover(err interface{}, e Event) {}

// Stop the behavior.
func (b *distributionBehavior) Stop() {}

// PoolConfig returns the pool size and that the distribution behavior is
// not stateful.
func (b *distributionBehavior) PoolConfig() (poolSize int, stateful bool) {
	return 100, false
}

//--------------------
// ORDER BEHAVIOR
//--------------------

// orderBehavior manages one order.
type orderBehavior struct {
	env                *Environment
	id                 Id
	orderNo            int
	distributionCenter int
	openItems          map[int]*OrderItem
	providedItems      map[int]*OrderItem
}

// NewOrderBehaviorFactory creates the factory for an order behavior.
func NewOrderBehaviorFactory(order *Order) BehaviorFactory {
	return func() Behavior {
		b := &orderBehavior{
			orderNo:       order.OrderNo,
			openItems:     order.OrderItems,
			providedItems: make(map[int]*OrderItem),
		}
		return b
	}
}

// Init the behavior.
func (b *orderBehavior) Init(env *Environment, id Id) error {
	b.env = env
	b.id = id
	return nil
}

// ProcessEvent processes an event.
func (b *orderBehavior) ProcessEvent(e Event, emitter EventEmitter) {
	switch e.Topic() {
	case "order":
		// Command to perform order received from shop.
		for _, orderItem := range b.openItems {
			stockCellId := NewId("stock", orderItem.ItemNo)
			b.env.Subscribe(stockCellId, b.id)
			b.env.EmitSimple(stockCellId, "order-item", orderItem)
		}
	case "order-item":
		// Item received from stock.
		orderItem := e.Payload().(*OrderItem)
		b.providedItems[orderItem.ItemNo] = orderItem
		delete(b.openItems, orderItem.ItemNo)
		// Check for open items. If none start delivery.
		if len(b.openItems) == 0 {
			// Deliver.
			shipment := &Shipment{b.orderNo, make(map[int]*ShipmentItem)}
			for itemNo, orderItem := range b.providedItems {
				shipment.ShipmentItems[itemNo] = &ShipmentItem{
					ShipmentNo: b.orderNo,
					ItemNo:     orderItem.ItemNo,
					Quantity:   orderItem.Quantity,
				}
			}
			emitter.EmitSimple("shipment", shipment)
			// Remove cell, it's needed anymore.
			b.env.RemoveCell(b.id)
		}
	}
}

// Recover from an error.
func (b *orderBehavior) Recover(err interface{}, e Event) {}

// Stop the behavior.
func (b *orderBehavior) Stop() {
	for itemNo, item := range b.openItems {
		applog.Infof("order %d has item %d open, needs %d", b.orderNo, itemNo, item.Quantity)
	}
}

//--------------------
// STOCK ITEM BEHAVIOR
//--------------------

// stockItemBehavior manages an item, its stock and the open
// orders for this item.
type stockItemBehavior struct {
	env        *Environment
	itemNo     int
	quantity   int
	orderItems []*OrderItem
}

// NewStockItemBehaviorFactory creates a stock behavior.
func NewStockItemBehaviorFactory(itemNo int) BehaviorFactory {
	return func() Behavior { return &stockItemBehavior{nil, itemNo, 0, []*OrderItem{}} }
}

// Init the behavior.
func (b *stockItemBehavior) Init(env *Environment, id Id) error {
	b.env = env
	return nil
}

// ProcessEvent processes an event. In case of an order item it's added to the open
// order items, if it's a shippment item the stock quantity will be increased. In
// both cases a delivery will be started.
func (b *stockItemBehavior) ProcessEvent(e Event, emitter EventEmitter) {
	switch e.Topic() {
	case "order-item":
		// Add order item to the list of orders for this item.
		orderItem := e.Payload().(*OrderItem)
		b.orderItems = append(b.orderItems, orderItem)
	case "shipment-item":
		// Add shipped item to the stock.
		shippmentItem := e.Payload().(*ShipmentItem)
		b.quantity += shippmentItem.Quantity
	}
	b.deliver(emitter)
}

// Recover from an error.
func (b *stockItemBehavior) Recover(err interface{}, e Event) {}

// Stop the behavior.
func (b *stockItemBehavior) Stop() {
	quantity := 0
	for _, orderItem := range b.orderItems {
		quantity += orderItem.Quantity
	}
	if quantity > 0 {
		applog.Infof("stock item %d needs %d items, has %d", b.itemNo, quantity, b.quantity)
	}
}

// deliver compares stock and orders. Orders will be delivered
// as long as possibe. If needed a supply order is emitted.
func (b *stockItemBehavior) deliver(emitter EventEmitter) {
	// Deliver to orders.
	for len(b.orderItems) > 0 && b.quantity >= b.orderItems[0].Quantity {
		emitter.EmitSimple("order-item", b.orderItems[0])
		b.quantity -= b.orderItems[0].Quantity
		b.orderItems = b.orderItems[1:]
	}
	// If more items needed call supply.
	if len(b.orderItems) > 0 {
		quantity := 0
		for _, orderItem := range b.orderItems {
			quantity += orderItem.Quantity
		}
		if quantity > 0 {
			b.env.EmitSimple("supply", "order-item", &OrderItem{0, b.itemNo, quantity})
		}
	}
}

//--------------------
// SUPPLY BEHAVIOR
//--------------------

// supplyBehavior manages the supply of items into the stock.
type supplyBehavior struct {
	env            *Environment
	orderNo        int
	itemQuantities map[int]int
}

// SupplyBehaviorFactory is the factory for a stock behavior.
func SupplyBehaviorFactory() Behavior {
	return &supplyBehavior{nil, 0, make(map[int]int)}
}

// Init the behavior.
func (b *supplyBehavior) Init(env *Environment, id Id) error {
	b.env = env
	return nil
}

// ProcessEvent processes an event.
func (b *supplyBehavior) ProcessEvent(e Event, emitter EventEmitter) {
	switch e.Topic() {
	case "order-item":
		// Order item by a stock cell, add the quantity.
		orderItem := e.Payload().(*OrderItem)
		b.itemQuantities[orderItem.ItemNo] += orderItem.Quantity
		// More than 10 ordered items, so let produce them.
		if len(b.itemQuantities) > 10 {
			b.manufacture(emitter)
		}
	case "ticker(manufacturing)":
		// Let the manufactures produce the collected items.
		b.manufacture(emitter)
	case "shipment":
		// Shipment of an order by the manufacturers, distribute the
		// items to the stock cells.
		shipment := e.Payload().(*Shipment)
		for _, shipmentItem := range shipment.ShipmentItems {
			stockCellId := NewId("stock", shipmentItem.ItemNo)
			b.env.EmitSimple(stockCellId, "shipment-item", shipmentItem)
		}
	}
}

// Recover from an error.
func (b *supplyBehavior) Recover(err interface{}, e Event) {}

// Stop the behavior.
func (b *supplyBehavior) Stop() {
	for itemNo, quantity := range b.itemQuantities {
		if quantity > 0 {
			applog.Infof("supply needs still %d pieces of item %d", quantity, itemNo)
		}
	}
}

// Manufacture the ordered items.
func (b *supplyBehavior) manufacture(emitter EventEmitter) {
	b.orderNo++
	order := &Order{b.orderNo, make(map[int]*OrderItem)}
	for itemNo, quantity := range b.itemQuantities {
		order.OrderItems[itemNo] = &OrderItem{b.orderNo, itemNo, quantity}
	}
	emitter.EmitSimple("order", order)
	b.itemQuantities = make(map[int]int)
}

//--------------------
// MANUFACTURER BEHAVIOR
//--------------------

// manufacturerBehavior manages one manufacturer for a set of items.
type manufacturerBehavior struct {
	id                   Id
	shipmentNo           int
	itemNoLow            int
	itemNoHigh           int
	packagingSize        int
	manufacturedItems    map[int]*ShipmentItem
	manufacturedQuantity int
}

// ManufacturerBehaviorFactory is the factory for a manufacturer behavior. It
// works a bit faked, all manufacturer receive order containing also items they
// don't produce. The filter their items matching to the range. Different item 
// ranges for the manufacturers result in different timespans until ordered 
// items are shipped (simulated bundeling).
func NewManufacturerBehaviorFactory(lo, hi, ps int) BehaviorFactory {
	return func() Behavior { return &manufacturerBehavior{"", 0, lo, hi, ps, make(map[int]*ShipmentItem), 0} }
}

// Init the behavior.
func (b *manufacturerBehavior) Init(env *Environment, id Id) error {
	b.id = id
	return nil
}

// ProcessEvent processes an event.
func (b *manufacturerBehavior) ProcessEvent(e Event, emitter EventEmitter) {
	switch e.Topic() {
	case "order":
		// Received an order from supply.
		order := e.Payload().(*Order)
		for itemNo, orderItem := range order.OrderItems {
			if itemNo >= b.itemNoLow && itemNo <= b.itemNoHigh {
				shipmentItem, ok := b.manufacturedItems[itemNo]
				if !ok {
					shipmentItem = &ShipmentItem{0, itemNo, 0}
					b.manufacturedItems[itemNo] = shipmentItem
				}
				shipmentItem.Quantity += orderItem.Quantity
				b.manufacturedQuantity += orderItem.Quantity
			}
		}
		// Check number of manufactured items.
		if b.manufacturedQuantity > b.packagingSize {
			// Packaging size reached, so ship.
			b.ship(emitter)
		}
	case "ticker(shipment)":
		// Ticker event for shipment.
		b.ship(emitter)
	}
}

// Recover from an error.
func (b *manufacturerBehavior) Recover(err interface{}, e Event) {}

// Stop the behavior.
func (b *manufacturerBehavior) Stop() {
	if b.manufacturedQuantity > 0 {
		applog.Infof("manufacturer for %d to %d has %d unshipped items", b.itemNoLow, b.itemNoHigh, b.manufacturedQuantity)
	}
}

// ship ships the manufactured items.
func (b *manufacturerBehavior) ship(emitter EventEmitter) {
	b.shipmentNo++
	shipment := &Shipment{b.shipmentNo, b.manufacturedItems}
	emitter.EmitSimple("shipment", shipment)
	b.manufacturedItems = make(map[int]*ShipmentItem)
	b.manufacturedQuantity = 0
}

//--------------------
// HELPERS
//--------------------

// OrdersAndShipments collects orders and shipments for tests.
type OrdersAndShipments struct {
	Count        int
	Orders       map[int]*Order
	Shipments    map[int]*Shipment
	OrderChan    chan *Order
	ShipmentChan chan *Shipment
	DoneChan     chan bool
}

// newOrdersAndShipments create a collector.
func newOrdersAndShipments(count int) *OrdersAndShipments {
	oas := &OrdersAndShipments{
		Count:        count,
		Orders:       make(map[int]*Order),
		Shipments:    make(map[int]*Shipment),
		OrderChan:    make(chan *Order, 4096),
		ShipmentChan: make(chan *Shipment, 4096),
		DoneChan:     make(chan bool, 1),
	}
	go oas.loop()
	return oas
}

// Stop ends the collecting.
func (oas *OrdersAndShipments) Wait(orders int) {
	timeout := time.Duration(10*orders) * time.Millisecond
	applog.Infof("waiting for max %v", timeout)
	select {
	case <-oas.DoneChan:
		return
	case <-time.After(timeout):
		return
	}
}

// Lengths returns the number or orders and shipments.
func (oas *OrdersAndShipments) Lengths() (int, int) {
	return len(oas.Orders), len(oas.Shipments)
}

// OrdersAreShipped checks if all orders are shipped.
func (oas *OrdersAndShipments) Compare() bool {
	if len(oas.Orders) != len(oas.Shipments) {
		return false
	}
	for orderNo := range oas.Orders {
		if _, ok := oas.Shipments[orderNo]; !ok {
			return false
		}
	}
	return true
}

// loop is the backend loop for the collector.
func (oas *OrdersAndShipments) loop() {
	for {
		select {
		case order := <-oas.OrderChan:
			oas.Orders[order.OrderNo] = order
		case shipment := <-oas.ShipmentChan:
			oas.Shipments[shipment.ShipmentNo] = shipment
		}
		if len(oas.Shipments) == oas.Count {
			applog.Infof("all shipments are done (%d = %d)", len(oas.Shipments), oas.Count)
			oas.DoneChan <- true
			return
		}
	}
}

// scenarioParam describes the parameters for a scenario test.
type scenarioParam struct {
	Id     string
	Items  int
	Orders int
}

// String returns the scenario parameters as string.
func (p scenarioParam) String() string {
	return fmt.Sprintf("scenario %q with max. %d items and %d orders", p.Id, p.Items, p.Orders)
}

// setUpEnvironment sets the environment for the tests up.
func setUpEnvironment(param scenarioParam) (*Environment, *OrdersAndShipments) {
	applog.Infof("starting set-up of environment")
	rand.Seed(time.Now().Unix())

	env := NewEnvironment(Id(param.Id))
	oas := newOrdersAndShipments(param.Orders)
	bfm := BehaviorFactoryMap{
		"shop":         NewShopBehaviorFactory(oas),
		"supply":       SupplyBehaviorFactory,
		"distribution": NewDistributionBehaviorFactory(oas),
		"manufacturer": BroadcastBehaviorFactory,
	}
	manufacturerRange := param.Items / 10
	for i := 1; i < 11; i++ {
		manufacturerId := NewId("manufacturer", i-1)
		itemNoLow := (i - 1) * manufacturerRange
		itemNoHigh := i*manufacturerRange - 1
		packagingSize := (rand.Intn(99) + 1) * 10
		bfm[manufacturerId] = NewManufacturerBehaviorFactory(itemNoLow, itemNoHigh, packagingSize)

	}
	for itemNo := 0; itemNo <= param.Items; itemNo++ {
		stockCellId := NewId("stock", itemNo)
		bfm[stockCellId] = NewStockItemBehaviorFactory(itemNo)
	}
	sm := SubscriptionMap{
		"supply": {"manufacturer"},
		"manufacturer": {
			"manufacturer:0", "manufacturer:1", "manufacturer:2", "manufacturer:3", "manufacturer:4",
			"manufacturer:5", "manufacturer:6", "manufacturer:7", "manufacturer:8", "manufacturer:9",
		},
		"manufacturer:0": {"supply"},
		"manufacturer:1": {"supply"},
		"manufacturer:2": {"supply"},
		"manufacturer:3": {"supply"},
		"manufacturer:4": {"supply"},
		"manufacturer:5": {"supply"},
		"manufacturer:6": {"supply"},
		"manufacturer:7": {"supply"},
		"manufacturer:8": {"supply"},
		"manufacturer:9": {"supply"},
	}
	applog.Infof("adding cells, subscriptions and ticker")
	env.AddCells(bfm)
	env.SubscribeAll(sm)
	env.AddTicker("manufacturing", "supply", 500*time.Millisecond)
	env.AddTicker("shipment", "manufacturer", 100*time.Millisecond)
	applog.Infof("set-up of environment done")

	return env, oas
}

// runScenarioTest runs a test with the given number of orders
func runScenarioTest(assert *asserts.Asserts, param scenarioParam) {
	applog.Infof(param.String())
	monitoring.Reset()

	env, oas := setUpEnvironment(param)

	for on := 0; on < param.Orders; on++ {
		env.EmitSimple("shop", "order", generateOrder(on, param.Items))
	}

	oas.Wait(param.Orders)

	ol, sl := oas.Lengths()

	assert.Equal(ol, param.Orders, "All orders have been placed.")
	assert.Equal(sl, param.Orders, "All orders have been delivered.")
	assert.Equal(ol, sl, "The number of orders and shipments is equal.")

	time.Sleep(2 * time.Second)
	env.Shutdown()
	time.Sleep(2 * time.Second)

	monitoring.MeasuringPointsPrintAll()
	monitoring.StaySetVariablesPrintAll()
}

//--------------------
// TESTS
//--------------------

var scenarioTestParams = []scenarioParam{
	{"scenario-s", 1000, 1000},
	{"scenario-m", 10000, 10000},
	{"scenario-l", 100000, 100000},
}

func TestScenarios(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	for _, param := range scenarioTestParams {
		runScenarioTest(assert, param)
	}
}

// EOF
