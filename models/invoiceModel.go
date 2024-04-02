package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Invoice struct {
	Id             primitive.ObjectID `bson:"_id" json:"_id"`
	InvoiceId      string             `bson:"invoice_id" json:"invoice_id"`
	OrderId        string             `bson:"order_id" json:"order_id"`
	PaymentMethod  *string            `bson:"payment_method" json:"payment_method" validate:"eq=CARD|eq=CASH|eq="`
	PaymentStatus  *string            `bson:"payment_status" json:"payment_status" validate:"required,eq=PENDING|eq=PAID"`
	PaymentDueDate time.Time          `bson:"payment_due_date" json:"payment_due_date"`
	CreatedAt      time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time          `bson:"updated_at" json:"updated_at"`
}
