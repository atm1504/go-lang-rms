package controller

import (
	"context"
	"log"
	"net/http"
	"time"

	"atm1504.in/rms/database"
	helper "atm1504.in/rms/helpers"
	"atm1504.in/rms/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")

func GetUsers() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, nil)

	}
}

func GetUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		userID := c.Param("user_id")
		var user models.User

		err := userCollection.FindOne(ctx, bson.M{"user": userID}).Decode(&user)

		defer cancel()

		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusNotFound, gin.H{
					"message": "User not found",
				})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while fetching the users"})
			return
		}
		c.JSON(http.StatusOK, user)
	}
}

func SignUp() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		var user models.User
		if err := c.BindJSON(&user); err != nil {
			defer cancel()
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validationErr := validate.Struct(user)
		if validationErr != nil {
			defer cancel()
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		emailCount, err := userCollection.CountDocuments(ctx, bson.M{"email": user.Email})
		if err != nil {
			defer cancel()
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if emailCount > 0 {
			defer cancel()
			c.JSON(http.StatusBadRequest, gin.H{"error": "Email already exists"})
			return
		}

		phoneCount, err := userCollection.CountDocuments(ctx, bson.M{"phone": user.Phone})
		if err != nil {
			defer cancel()
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if phoneCount > 0 {
			defer cancel()
			c.JSON(http.StatusBadRequest, gin.H{"error": "Phone already exists"})
			return
		}

		password := HashPassword(*user.Password)
		user.Password = &password

		user.CreatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.UpdatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.ID = primitive.NewObjectID()
		user.UserID = user.ID.Hex()

		token, refreshToken, _ := helper.GenerateAllTokens(*user.Email, *user.FirstName, *user.LastName, user.UserID)
		user.Token = &token
		user.RefreshToken = &refreshToken
		//if all ok, then you insert this new user into the user collection

		resultInsertionNumber, insertErr := userCollection.InsertOne(ctx, user)
		defer cancel()
		if insertErr != nil {
			msg := "User item was not created"
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		c.JSON(http.StatusOK, resultInsertionNumber)

	}
}

func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, nil)

	}
}

func HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Panic(err)
	}

	return string(bytes)
}

func VerifyPassword(userPassword string, providedPassword string) (bool, string) {

	err := bcrypt.CompareHashAndPassword([]byte(providedPassword), []byte(userPassword))
	check := true
	msg := ""

	if err != nil {
		msg = "login or password is incorrect"
		check = false
	}
	return check, msg
}
