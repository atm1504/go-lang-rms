package models

import (
	"time"
)

type Table struct {
	ID             int64     `bson:"id" json:"id"`
	NumberOfGuests *int      `bson:"number_of_guests" json:"number_of_guests" validate:"required"`
	TableNumber    *int      `bson:"table_number" json:"table_number" validate:"required"`
	CreatedAt      time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time `bson:"updated_at" json:"updated_at"`
}
