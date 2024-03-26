package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type OrderItem struct {
	id          primitive.ObjectID `bson:"_id"`
	quantity    *string            `json:"quantity" validate:"required,eq=S|eq=M|eq=L"`
	unitPrice   *float64           `json:"unit_price" validate:"required"`
	createdAt   time.Time          `json:"created_at"`
	updatedAt   time.Time          `json:"updated_at"`
	foodId      *string            `json:"food_id" validate:"required"`
	orderItemId string             `json:"order_item_id"`
	orderId     string             `json:"order_id" validate:"required"`
}
