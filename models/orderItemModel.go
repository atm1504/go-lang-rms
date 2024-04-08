package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type OrderItem struct {
	ID          primitive.ObjectID `bson:"_id" json:"_id"`
	Quantity    *string            `bson:"quantity" json:"quantity" validate:"required,eq=S|eq=M|eq=L"`
	UnitPrice   *float64           `bson:"unit_price" json:"unit_price" validate:"required"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
	FoodID      *string            `bson:"food_id" json:"food_id" validate:"required"`
	OrderItemID string             `bson:"order_item_id" json:"order_item_id"`
	OrderID     string             `bson:"order_id" json:"order_id" validate:"required"`
}
