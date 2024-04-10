package models

import (
	"time"
)

type Note struct {
	ID        int64     `bson:"id" json:"id"`
	Text      string    `bson:"text" json:"text"`
	Title     string    `bson:"title" json:"title"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}
