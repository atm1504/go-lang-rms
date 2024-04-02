package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Note struct {
	Id        primitive.ObjectID `bson:"_id" json:"_id"`
	Text      string             `bson:"text" json:"text"`
	Title     string             `bson:"title" json:"title"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
	NoteId    string             `bson:"note_id" json:"note_id"`
}
