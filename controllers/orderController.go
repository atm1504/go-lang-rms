package controller

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"atm1504.in/rms/database"
	"atm1504.in/rms/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var orderCollection *mongo.Collection = database.OpenCollection(database.Client, "order")
var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

func GetOrders() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		recordPerPage, err := strconv.Atoi(c.DefaultQuery("recordPerPage", "10"))
		if err != nil || recordPerPage < 1 {
			recordPerPage = 10
		}

		page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
		if err != nil || page < 1 {
			page = 1
		}

		startIndex := (page - 1) * recordPerPage

		matchStage := bson.D{{Key: "$match", Value: bson.D{}}}
		groupStage := bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: nil},
			{Key: "total_count", Value: bson.D{{Key: "$sum", Value: 1}}},
			{Key: "data", Value: bson.D{{Key: "$push", Value: "$$ROOT"}}},
		}}}
		projectStage := bson.D{
			{Key: "$project", Value: bson.D{
				{Key: "_id", Value: 0},
				{Key: "total_count", Value: 1},
				{Key: "orders", Value: bson.D{{Key: "$slice", Value: []interface{}{"$data", startIndex, recordPerPage}}}},
			}},
		}

		result, err := orderCollection.Aggregate(ctx, mongo.Pipeline{matchStage, groupStage, projectStage})
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occurred while listing menus"})
			return
		}

		var allOrders []bson.M
		if err = result.All(ctx, &allOrders); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occurred while processing menus"})
			log.Fatal(err)
		}

		// Assuming you want to return the list of menus directly
		if len(allOrders) > 0 {
			c.JSON(http.StatusOK, allOrders[0])
		} else {
			c.JSON(http.StatusOK, []interface{}{}) // Return an empty array if no menus
		}
	}
}

func GetOrder() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var order models.Order

		orderID := c.Param("order_id")
		fmt.Println("Order id is: ", orderID)

		err := orderCollection.FindOne(ctx, bson.M{"order_id": orderID}).Decode(&order)
		defer cancel()
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusNotFound, gin.H{
					"message": "Order not found",
				})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error in fetching order details"})
			return
		}
		c.JSON(http.StatusOK, order)
	}
}

func CreateOrder() gin.HandlerFunc {
	return func(c *gin.Context) {
		// var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		// var table models.Table
		// var order models.Order

		// if err := c.BindJSON(&order); err != nil {
		// 	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		// 	defer cancel()
		// 	return
		// }

		// validationErr := validate.Struct(order)
		// if validationErr != nil {
		// 	c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
		// 	defer cancel()
		// 	return
		// }

		// if order.TableID != nil {
		// 	err := tableCollection.FindOne(context.Background(), bson.M{"table_id": order.TableID}).Decode(&table)
		// 	defer cancel()
		// 	if err != nil {
		// 		if err == mongo.ErrNoDocuments {
		// 			c.JSON(http.StatusNotFound, gin.H{
		// 				"message": "Table data not found",
		// 			})
		// 			return
		// 		}
		// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error in fetching table details"})
		// 		return
		// 	}
		// }

		// order.CreatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		// order.UpdatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

		// order.ID = primitive.NewObjectID()
		// order.OrderID = order.ID.Hex()

		// result, insertErr := orderCollection.InsertOne(ctx, order)
		// defer cancel()
		// if insertErr != nil {
		// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "order item was not created"})
		// 	return
		// }

		// defer cancel()
		// c.JSON(http.StatusOK, result)

	}
}

func UpdateOrder() gin.HandlerFunc {
	return func(c *gin.Context) {
		// var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		// var table models.Table
		// var order models.Order

		// var updateObj primitive.D

		// orderID := c.Param("order_id")

		// if err := c.BindJSON(&order); err != nil {
		// 	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		// 	defer cancel()
		// 	return
		// }

		// if order.TableID != nil {
		// 	err := tableCollection.FindOne(ctx, bson.M{"table_id": order.TableID}).Decode(&table)
		// 	defer cancel()
		// 	if err != nil {
		// 		if err == mongo.ErrNoDocuments {
		// 			c.JSON(http.StatusNotFound, gin.H{
		// 				"message": "Table data not found",
		// 			})
		// 			return
		// 		}
		// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error in fetching table details"})
		// 		return
		// 	}
		// 	updateObj = append(updateObj, bson.E{Key: "table_id", Value: order.TableID})
		// }

		// order.UpdatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		// updateObj = append(updateObj, bson.E{Key: "updated_at", Value: order.UpdatedAt})

		// upsert := true
		// filter := bson.M{"order_id": orderID}
		// opt := options.UpdateOptions{
		// 	Upsert: &upsert,
		// }

		// result, err := orderCollection.UpdateOne(
		// 	ctx, filter, bson.D{
		// 		{Key: "$set", Value: updateObj},
		// 	},
		// 	&opt,
		// )

		// defer cancel()
		// if err != nil {
		// 	msg := "order item update failed"
		// 	c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
		// 	return
		// }
		// c.JSON(http.StatusOK, result)

	}
}

func OrderItemOrderCreator(order models.Order) string {

	// order.CreatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	// order.UpdatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	// order.ID = primitive.NewObjectID()
	// order.OrderID = order.ID.Hex()

	// err, _ := orderCollection.InsertOne(ctx, order)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer cancel()

	// return order.OrderID
	return ""
}
