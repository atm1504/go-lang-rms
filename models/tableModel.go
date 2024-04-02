package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Table struct {
	Id             primitive.ObjectID `bson:"_id" json:"_id"`
	NumberOfGuests *int               `bson:"number_of_guests" json:"number_of_guests" validate:"required"`
	TableNumber    *int               `bson:"table_number" json:"table_number" validate:"required"`
	CreatedAt      time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time          `bson:"updated_at" json:"updated_at"`
	TableId        string             `bson:"table_id" json:"table_id"`
}
