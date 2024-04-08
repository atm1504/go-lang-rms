package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Order struct {
	ID        primitive.ObjectID `bson:"_id" json:"_id"`
	OrderDate time.Time          `bson:"order_date" json:"order_date" validate:"required"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
	OrderID   string             `bson:"order_id" json:"order_id"`
	TableID   *string            `bson:"table_id" json:"table_id" validate:"required"`
}
