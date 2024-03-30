package controller

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"time"

	"atm1504.in/rms/database"
	"atm1504.in/rms/models"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var foodCollection *mongo.Collection = database.OpenCollection(database.Client, "food")
var menuCollection *mongo.Collection = database.OpenCollection(database.Client, "menu")

var validate = validator.New()

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

func CreateFood() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var menu models.Menu
		var food models.Food

		if err := c.BindJSON(&food); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validationErr := validate.Struct(food)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		err := menuCollection.FindOne(ctx, bson.M{"menu_id": food.MenuId}).Decode(&menu)
		defer cancel()
		if err != nil {
			fmt.Println(err)
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusNotFound, gin.H{
					"message": "Menu not found",
				})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"erroe": "Error in fetching menu details"})
			return
		}

		food.CreatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		food.UpdatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		food.Id = primitive.NewObjectID()
		food.FoodId = food.Id.Hex()
		var num = toFixed(*food.Price, 2)
		food.Price = &num
		result, inserErr := foodCollection.InsertOne(ctx, food)

		if inserErr != nil {
			msg := fmt.Sprintf("Food item was not created")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		defer cancel()
		c.JSON(http.StatusOK, result)

	}
}

func round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

func toFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(round(num*output)) / output
}

func UpdateFood() gin.HandlerFunc {
	return func(c *gin.Context) {

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var menu models.Menu
		var food models.Food

		foodId := c.Param("food_id")

		if err := c.BindJSON(&food); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			defer cancel()
			return
		}

		var updateObj primitive.D
		if food.Name != nil {
			updateObj = append(updateObj, bson.E{Key: "name", Value: food.Name})
		}
		if food.Price != nil {
			updateObj = append(updateObj, bson.E{Key: "price", Value: food.Price})
		}

		if food.FoodImage != nil {
			updateObj = append(updateObj, bson.E{Key: "food_image", Value: food.Price})
		}

		if food.MenuId != nil {
			err := menuCollection.FindOne(ctx, bson.M{"menu_id": food.MenuId}).Decode(&menu)
			defer cancel()
			if err != nil {

				if err == mongo.ErrNoDocuments {
					msg := "message:Menu was not found"
					c.JSON(http.StatusNotFound, gin.H{"error": msg})
					return
				}

				msg := fmt.Sprintf("Internal Server Error occurred: %m", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
				return
			}
			updateObj = append(updateObj, bson.E{Key: "menu", Value: food.Price})
		}
		food.UpdatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		updateObj = append(updateObj, bson.E{Key: "updated_at", Value: food.UpdatedAt})

		upsert := true
		filter := bson.M{"food_id": foodId}
		opt := options.UpdateOptions{
			Upsert: &upsert,
		}

		result, err := foodCollection.UpdateOne(
			ctx, filter, bson.D{
				{Key: "$set", Value: updateObj},
			},
			&opt,
		)

		if err != nil {
			msg := "food item update failed"
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			defer cancel()
			return
		}
		c.JSON(http.StatusOK, result)
	}
}
