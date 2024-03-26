package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Table struct {
	id             primitive.ObjectID `bson:"_id"`
	numberOfGuests *int               `json:"number_of_guests" validate:"required"`
	tableNumber    *int               `json:"table_number" validate:"required"`
	createdAt      time.Time          `json:"created_at"`
	updatedAt      time.Time          `json:"updated_at"`
	tableId        string             `json:"table_id"`
}
