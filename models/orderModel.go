package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Order struct {
	id        primitive.ObjectID `bson:"_id"`
	orderDate time.Time          `json:"order_date" validate:"required"`
	createdAt time.Time          `json:"created_at"`
	updatedAt time.Time          `json:"updated_at"`
	orderId   string             `json:"order_id"`
	tableId   *string            `json:"table_id" validate:"required"`
}
