package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Note struct {
	id        primitive.ObjectID `bson:"_id"`
	text      string             `json:"text"`
	title     string             `json:"title"`
	createdAt time.Time          `json:"created_at"`
	updatedAt time.Time          `json:"updated_at"`
	noteId    string             `json:"note_id"`
}
