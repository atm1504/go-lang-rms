package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Menu struct {
	id        primitive.ObjectID `bson:"_id"`
	name      string             `json:"name" validate:"required"`
	category  string             `json:"category" validate:"required"`
	startDate *time.Time         `json:"start_date"`
	endDate   *time.Time         `json:"end_date"`
	createdAt time.Time          `json:"created_at"`
	updatedAt time.Time          `json:"updated_at"`
	menuId    string             `json:"food_id"`
}
