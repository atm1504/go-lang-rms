package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Food struct {
	id        primitive.ObjectID `bson:"_id"`
	name      *string            `json:"name" validate:"required,min=2,max=100"`
	price     *float64           `json:"price" validate:"required"`
	foodImage *string            `json:"food_image" validate:"required"`
	createdAt time.Time          `json:"created_at"`
	updatedAt time.Time          `json:"updated_at"`
	foodId    string             `json:"food_id"`
	menuId    *string            `json:"menu_id" validate:"required"`
}
