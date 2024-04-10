package models

import (
	"time"
)

type Menu struct {
	ID        int64      `bson:"id" json:"id"`
	Name      string     `bson:"name" json:"name" validate:"required"`
	Category  string     `bson:"category" json:"category" validate:"required"`
	StartDate *time.Time `bson:"start_date" json:"start_date"`
	EndDate   *time.Time `bson:"end_date" json:"end_date"`
	CreatedAt time.Time  `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time  `bson:"updated_at" json:"updated_at"`
}
