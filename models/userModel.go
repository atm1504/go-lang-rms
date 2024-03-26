package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	id            primitive.ObjectID `bson:"_id"`
	firstName    *string            `json:"first_name" validate:"required,min=2,max=100"`
	lastName     *string            `json:"last_name" validate:"required,min=2,max=100"`
	password      *string            `json:"Password" validate:"required,min=6"`
	email         *string            `json:"email" validate:"email,required"`
	avatar        *string            `json:"avatar"`
	phone         *string            `json:"phone" validate:"required"`
	token         *string            `json:"token"`
	refreshToken *string            `json:"refresh_token"`
	createdAt    time.Time          `json:"created_at"`
	updatedAt    time.Time          `json:"updated_at"`
	userId       string             `json:"user_id"`
}
