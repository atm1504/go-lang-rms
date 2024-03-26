package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Invoice struct {
	id             primitive.ObjectID `bson:"_id"`
	invoiceId      string             `json:"invoice_id"`
	orderId        string             `json:"order_id"`
	paymentMethod  *string            `json:"payment_method" validate:"eq=CARD|eq=CASH|eq="`
	paymentStatus  *string            `json:"payment_status" validate:"required,eq=PENDING|eq=PAID"`
	paymentDueDate time.Time          `json:"Payment_due_date"`
	createdAt      time.Time          `json:"created_at"`
	updatedAt      time.Time          `json:"updated_at"`
}
