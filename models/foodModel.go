package models

import (
	"time"
)

type Food struct {
	ID        int64     `bson:"id" json:"id"`
	Name      *string   `bson:"name" json:"name" validate:"required,min=2,max=100"`
	Price     *float64  `bson:"price" json:"price" validate:"required"`
	FoodImage *string   `bson:"food_image" json:"food_image" validate:"required"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
	MenuID    int64     `bson:"menu_id" json:"menu_id" validate:"required"`
}
