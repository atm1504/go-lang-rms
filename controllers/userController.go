package controller

import (
	"net/http"

	//"atm1504.in/rms/models"

	"github.com/gin-gonic/gin"
)

// var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")

func GetUsers() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, nil)

	}
}
