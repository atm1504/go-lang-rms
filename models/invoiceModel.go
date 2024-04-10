package models

import (
	"time"
)

type Invoice struct {
	ID             int64     `bson:"id" json:"id"`
	OrderID        int64     `bson:"order_id" json:"order_id"`
	PaymentMethod  *string   `bson:"payment_method" json:"payment_method" validate:"eq=CARD|eq=CASH|eq="`
	PaymentStatus  *string   `bson:"payment_status" json:"payment_status" validate:"required,eq=PENDING|eq=PAID"`
	PaymentDueDate time.Time `bson:"payment_due_date" json:"payment_due_date"`
	CreatedAt      time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time `bson:"updated_at" json:"updated_at"`
}
