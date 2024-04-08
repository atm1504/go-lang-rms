package controller

import "time"

type InvoiceViewFormat struct {
	InvoiceID      string
	PaymentMethod  string
	OrderID        string
	PaymentStatus  *string
	PaymentSue     interface{}
	TableNumber    interface{}
	PaymentDueDate time.Time
	OrderDetails   interface{}
}
