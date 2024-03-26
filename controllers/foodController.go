package controller

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"atm1504.in/rms/database"
	"atm1504.in/rms/models"

	//"atm1504.in/rms/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var foodCollection *mongo.Collection = database.OpenCollection(database.Client, "food")

func GetFood() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		foodId := c.Param("food_id")
		var food models.Food
		err := foodCollection.FindOne(ctx, bson.M{"food_id": foodId}).Decode(&food)
		// fmt.Println(food)
		defer cancel()
		if err != nil {
			fmt.Println(err)
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusNotFound, gin.H{
					"message": "Food not found",
				})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"erroe": "Error in fetching product"})
			return
		}
		c.JSON(http.StatusOK, food)
	}
}
