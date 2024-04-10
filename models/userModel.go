package models

import (
	"time"
)

type User struct {
	ID           int64     `bson:"id" json:"id"`
	FirstName    *string   `bson:"first_name" json:"first_name" validate:"required,min=2,max=100"`
	LastName     *string   `bson:"last_name" json:"last_name" validate:"required,min=2,max=100"`
	Password     *string   `bson:"password" json:"password" validate:"required,min=6"`
	Email        *string   `bson:"email" json:"email" validate:"email,required"`
	Avatar       *string   `bson:"avatar" json:"avatar"`
	Phone        *string   `bson:"phone" json:"phone" validate:"required"`
	Token        *string   `bson:"token" json:"token"`
	RefreshToken *string   `bson:"refresh_token" json:"refresh_token"`
	CreatedAt    time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time `bson:"updated_at" json:"updated_at"`
}
