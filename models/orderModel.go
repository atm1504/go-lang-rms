package models

import (
	"time"
)

type Order struct {
	ID        int64     `bson:"id" json:"id"`
	OrderDate time.Time `bson:"order_date" json:"order_date" validate:"required"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
	TableID   int64     `bson:"table_id" json:"table_id" validate:"required"`
}
