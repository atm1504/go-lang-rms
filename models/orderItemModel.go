package models

import (
	"time"
)

type OrderItem struct {
	ID        int64     `bson:"id" json:"id"`
	Quantity  *string   `bson:"quantity" json:"quantity" validate:"required,eq=S|eq=M|eq=L"`
	UnitPrice *float64  `bson:"unit_price" json:"unit_price" validate:"required"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
	FoodID    int64     `bson:"food_id" json:"food_id" validate:"required"`
	OrderID   int64     `bson:"order_id" json:"order_id" validate:"required"`
}
