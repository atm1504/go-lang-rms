package controller

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"

	"atm1504.in/rms/database"
	db "atm1504.in/rms/database"
	"atm1504.in/rms/models"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var foodCollection *mongo.Collection = database.OpenCollection(database.Client, "food")

var validate = validator.New()

// var dbConn, dbErr = db.DBInstanceSql()

func GetFoods() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		recordPerPage, err := strconv.Atoi(c.Query("recordPerPage"))
		if err != nil || recordPerPage < 1 {
			recordPerPage = 10
		}

		page, err := strconv.Atoi(c.Query("page"))
		if err != nil || page < 1 {
			page = 1
		}

		startIndex := (page - 1) * recordPerPage
		// startIndex, _ := strconv.Atoi(c.Query("startIndex"))

		matchStage := bson.D{{Key: "$match", Value: bson.D{{}}}}
		groupStage := bson.D{{Key: "$group", Value: bson.D{{Key: "_id", Value: bson.D{{Key: "_id", Value: "null"}}}, {Key: "total_count", Value: bson.D{{Key: "$sum", Value: 1}}}, {Key: "data", Value: bson.D{{Key: "$push", Value: "$$ROOT"}}}}}}
		projectStage := bson.D{
			{
				Key: "$project", Value: bson.D{
					{Key: "_id", Value: 0},
					{Key: "total_count", Value: 1},
					{Key: "food_items", Value: bson.D{{Key: "$slice", Value: []interface{}{"$data", startIndex, recordPerPage}}}},
				}}}

		result, err := foodCollection.Aggregate(ctx, mongo.Pipeline{
			matchStage, groupStage, projectStage})
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while listing food items"})
		}
		var allFoods []bson.M
		if err = result.All(ctx, &allFoods); err != nil {
			log.Fatal(err)
		}
		// c.JSON(http.StatusOK, allFoods[0])
		// Assuming you want to return the list of menus directly
		if len(allFoods) > 0 {
			c.JSON(http.StatusOK, allFoods[0])
		} else {
			c.JSON(http.StatusOK, []interface{}{}) // Return an empty array if no menus
		}

	}
}

func GetFood() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		foodID := c.Param("food_id")
		var food models.Food
		err := foodCollection.FindOne(ctx, bson.M{"food_id": foodID}).Decode(&food)
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

		var dbConn, dbErr = db.DBInstanceSql()
		if dbErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error connecting to database"})
			return
		}
		var menu models.Menu
		var food models.Food

		if err := c.BindJSON(&food); err != nil {
			defer cancel()
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validationErr := validate.Struct(food)
		if validationErr != nil {
			defer cancel()
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		// err := menuCollection.FindOne(ctx, bson.M{"menu_id": food.MenuID}).Decode(&menu)
		row := dbConn.QueryRowContext(ctx, "SELECT * FROM menu WHERE menu_id = ?", food.MenuID)

		// Scan the query result into a menu struct
		if err := row.Scan(&menu.ID, &menu.Name, &menu.Category, &menu.StartDate, &menu.EndDate, &menu.CreatedAt, &menu.UpdatedAt); err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"message": "Menu not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error in fetching menu details"})
			return
		}

		// Set the created and updated timestamps
		now := time.Now()
		food.CreatedAt = now
		food.UpdatedAt = now

		// Perform the insertion into the database
		result, err := dbConn.ExecContext(ctx, "INSERT INTO food (name, menu_id, price, created_at, updated_at) VALUES (?, ?, ?, ?, ?)",
			food.Name, food.MenuID, food.Price, food.CreatedAt, food.UpdatedAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert food item"})
			return
		}

		foodID, err := result.LastInsertId()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get food item ID"})
			return
		}

		// Set the food ID and send the response
		food.ID = foodID
		c.JSON(http.StatusOK, gin.H{"message": "Food item created successfully", "food": food})
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
		// var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		// var menu models.Menu
		// var food models.Food

		// foodID := c.Param("food_id")

		// if err := c.BindJSON(&food); err != nil {
		// 	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		// 	defer cancel()
		// 	return
		// }

		// var updateObj primitive.D
		// if food.Name != nil {
		// 	updateObj = append(updateObj, bson.E{Key: "name", Value: food.Name})
		// }
		// if food.Price != nil {
		// 	updateObj = append(updateObj, bson.E{Key: "price", Value: food.Price})
		// }

		// if food.FoodImage != nil {
		// 	updateObj = append(updateObj, bson.E{Key: "food_image", Value: food.FoodImage})
		// }

		// if food.MenuID != nil {
		// 	err := menuCollection.FindOne(ctx, bson.M{"menu_id": food.MenuID}).Decode(&menu)
		// 	defer cancel()
		// 	if err != nil {

		// 		if err == mongo.ErrNoDocuments {
		// 			msg := "message:Menu was not found"
		// 			c.JSON(http.StatusNotFound, gin.H{"error": msg})
		// 			return
		// 		}

		// 		msg := fmt.Sprintf("Internal Server Error occurred: %m", err)
		// 		c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
		// 		return
		// 	}
		// 	updateObj = append(updateObj, bson.E{Key: "menu_id", Value: food.MenuID})
		// }
		// food.UpdatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		// updateObj = append(updateObj, bson.E{Key: "updated_at", Value: food.UpdatedAt})

		// upsert := true
		// filter := bson.M{"food_id": foodID}
		// opt := options.UpdateOptions{
		// 	Upsert: &upsert,
		// }

		// result, err := foodCollection.UpdateOne(
		// 	ctx, filter, bson.D{
		// 		{Key: "$set", Value: updateObj},
		// 	},
		// 	&opt,
		// )
		// defer cancel()
		// if err != nil {
		// 	msg := "food item update failed"
		// 	c.JSON(http.StatusInternalServerError, gin.H{"error": msg})

		// 	return
		// }
		// c.JSON(http.StatusOK, result)
	}
}
