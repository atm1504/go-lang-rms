package controller

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"

	db "atm1504.in/rms/database"
	"atm1504.in/rms/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// var orderItemCollection *mongo.Collection = database.OpenCollection(database.Client, "orderItem")

type OrderItemPack struct {
	TableID    int64              `bson:"table_id" json:"table_id"`
	OrderItems []models.OrderItem `bson:"order_items" json:"order_items"`
}

func GetOrderItems() gin.HandlerFunc {
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
		dbConn, dbErr := db.DBInstanceSql()

		if ISEInjection(c, dbErr, "Error connecting to database") {
			return
		}
		defer dbConn.Close()
		query := `SELECT COUNT(*) OVER(), id, quantity, created_at, updated_at, unit_price, food_id, order_id  FROM order_item LIMIT ? OFFSET ?`

		fmt.Println(query)
		orderRows, err := dbConn.QueryContext(ctx, query, recordPerPage, startIndex)
		if ISEInjection(c, err, "Error fetching order items") {
			return
		}

		defer orderRows.Close()

		var totalCount int
		var orderItemsList []models.OrderItem
		for orderRows.Next() {
			var orderObj models.OrderItem
			var createdAtStr string
			var updatedAtStr string
			err := orderRows.Scan(&totalCount, &orderObj.ID, &orderObj.Quantity, &createdAtStr, &updatedAtStr, &orderObj.UnitPrice, &orderObj.FoodID, &orderObj.OrderID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error in fetching tables"})
				return
			}

			createdAt, err3 := ParseTime(createdAtStr)
			updatedAt, err4 := ParseTime(updatedAtStr)

			if ISEInjection(c, err3, "Error parsing time strings") || ISEInjection(c, err4, "Error parsing time strings") {
				return
			}

			orderObj.CreatedAt = createdAt
			orderObj.UpdatedAt = updatedAt

			orderItemsList = append(orderItemsList, orderObj)
		}

		if err = orderRows.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error during rows iteration"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"total_count": totalCount,
			"items":       orderItemsList,
		})
	}
}

func GetOrderItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		var orderItem models.OrderItem

		orderItemID := c.Param("order_item_id")
		fmt.Println("Order id is: ", orderItemID)

		var dbConn, dbErr = db.DBInstanceSql()
		if ISEInjection(c, dbErr, "Error connecting to database") {
			return
		}

		var createdAtStr string
		var updatedAtStr string

		row := dbConn.QueryRowContext(ctx, "SELECT id, quantity, created_at, updated_at, unit_price, food_id, order_id  FROM order_item WHERE id = ?", orderItemID)
		if err := row.Scan(&orderItem.ID, &orderItem.Quantity, &createdAtStr, &updatedAtStr, &orderItem.UnitPrice, &orderItem.FoodID, &orderItem.OrderID); err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"message": "Order item not found"})
				return
			}
			fmt.Println(err)
			ISEInjection(c, err, "Error in fetching order item details")
			return
		}

		createdAt, err3 := ParseTime(createdAtStr)
		updatedAt, err4 := ParseTime(updatedAtStr)

		if ISEInjection(c, err3, "Error parsing time strings") || ISEInjection(c, err4, "Error parsing time strings") {
			return
		}

		orderItem.CreatedAt = createdAt
		orderItem.UpdatedAt = updatedAt

		c.JSON(http.StatusOK, orderItem)
	}
}

func GetOrderItemsByOrder() gin.HandlerFunc {
	return func(c *gin.Context) {
		// orderID := c.Param("order_id")

		// allOrderItems, err := ItemsByOrder(orderID)

		// if err != nil {
		// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while listing order items by order ID"})
		// 	return
		// }
		// c.JSON(http.StatusOK, allOrderItems)
	}
}

func ItemsByOrder(id string) (OrderItems []primitive.M, err error) {
	// var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

	// matchStage := bson.D{{Key: "$match", Value: bson.D{{Key: "order_id", Value: id}}}}
	// lookupStage := bson.D{{Key: "$lookup", Value: bson.D{{Key: "from", Value: "food"}, {Key: "localField", Value: "food_id"}, {Key: "foreignField", Value: "food_id"}, {Key: "as", Value: "food"}}}}
	// unwindStage := bson.D{{Key: "$unwind", Value: bson.D{{Key: "path", Value: "$food"}, {Key: "preserveNullAndEmptyArrays", Value: true}}}}

	// lookupOrderStage := bson.D{{Key: "$lookup", Value: bson.D{{Key: "from", Value: "order"}, {Key: "localField", Value: "order_id"}, {Key: "foreignField", Value: "order_id"}, {Key: "as", Value: "order"}}}}
	// unwindOrderStage := bson.D{{Key: "$unwind", Value: bson.D{{Key: "path", Value: "$order"}, {Key: "preserveNullAndEmptyArrays", Value: true}}}}

	// lookupTableStage := bson.D{{Key: "$lookup", Value: bson.D{{Key: "from", Value: "table"}, {Key: "localField", Value: "order.table_id"}, {Key: "foreignField", Value: "table_id"}, {Key: "as", Value: "table"}}}}
	// unwindTableStage := bson.D{{Key: "$unwind", Value: bson.D{{Key: "path", Value: "$table"}, {Key: "preserveNullAndEmptyArrays", Value: true}}}}

	// projectStage := bson.D{
	// 	{Key: "$project", Value: bson.D{
	// 		{Key: "id", Value: 0},
	// 		{Key: "amount", Value: "$food.price"},
	// 		{Key: "total_count", Value: 1},
	// 		{Key: "food_name", Value: "$food.name"},
	// 		{Key: "food_image", Value: "$food.food_image"},
	// 		{Key: "table_number", Value: "$table.table_number"},
	// 		{Key: "table_id", Value: "$table.table_id"},
	// 		{Key: "order_id", Value: "$order.order_id"},
	// 		{Key: "price", Value: "$food.price"},
	// 		{Key: "quantity", Value: 1},
	// 	}}}

	// groupStage := bson.D{{Key: "$group", Value: bson.D{{Key: "_id", Value: bson.D{{Key: "order_id", Value: "$order_id"}, {Key: "table_id", Value: "$table_id"}, {Key: "table_number", Value: "$table_number"}}}, {Key: "payment_due", Value: bson.D{{Key: "$sum", Value: "$amount"}}}, {Key: "total_count", Value: bson.D{{Key: "$sum", Value: 1}}}, {Key: "order_items", Value: bson.D{{Key: "$push", Value: "$$ROOT"}}}}}}

	// projectStage2 := bson.D{
	// 	{Key: "$project", Value: bson.D{

	// 		{Key: "id", Value: 0},
	// 		{Key: "payment_due", Value: 1},
	// 		{Key: "total_count", Value: 1},
	// 		{Key: "table_number", Value: "$_id.table_number"},
	// 		{Key: "order_items", Value: 1},
	// 	}}}

	// result, err := orderItemCollection.Aggregate(ctx, mongo.Pipeline{
	// 	matchStage,
	// 	lookupStage,
	// 	unwindStage,
	// 	lookupOrderStage,
	// 	unwindOrderStage,
	// 	lookupTableStage,
	// 	unwindTableStage,
	// 	projectStage,
	// 	groupStage,
	// 	projectStage2})

	// if err != nil {
	// 	panic(err)
	// }

	// if err = result.All(ctx, &OrderItems); err != nil {
	// 	panic(err)
	// }

	// defer cancel()

	// return OrderItems, err
	return nil, nil
}

func CreateOrderItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		var dbConn, dbErr = db.DBInstanceSql()
		if ISEInjection(c, dbErr, "Error connecting to database") {
			return
		}
		var orderItemPack OrderItemPack
		var order models.Order

		if err := c.BindJSON(&orderItemPack); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			defer cancel()
			return
		}
		currentTime := time.Now()
		order.OrderDate = &currentTime
		orderItemsToBeInserted := []interface{}{}
		order.TableID = orderItemPack.TableID
		orderID := OrderItemOrderCreator(order, currentTime, dbConn, ctx)

		for _, orderItem := range orderItemPack.OrderItems {
			orderItem.OrderID = orderID
			validationErr := validate.Struct((orderItem))
			if validationErr != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
				defer cancel()
				return
			}
			orderItem.CreatedAt = currentTime
			orderItem.UpdatedAt = currentTime
			var num = toFixed(*orderItem.UnitPrice, 2)
			orderItem.UnitPrice = &num
			orderItemsToBeInserted = append(orderItemsToBeInserted, orderItem)
		}
		tx, txErr := dbConn.BeginTx(ctx, nil)
		if ISEInjection(c, txErr, "Error connecting while initiating a database transaction") {
			return
		}
		stmt, stmtErr := tx.PrepareContext(ctx, "INSERT INTO order_item (quantity, unit_price, created_at, updated_at, food_id, order_id) VALUES (?, ?, ?, ?, ?, ?)")

		if ISEInjection(c, stmtErr, "Error in transaction statement") {
			tx.Rollback()
			defer cancel()
			return
		}

		defer stmt.Close()

		for _, item := range orderItemsToBeInserted {
			// Assuming item is the struct you want to insert into the database
			orderItem, ok := item.(models.OrderItem) // Replace YourOrderItemType with the actual type of your order item struct
			if !ok {
				// Handle the case where the assertion fails
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assert order item type"})
				tx.Rollback()
				defer cancel()
				return
			}
			_, execErr := stmt.ExecContext(ctx, orderItem.Quantity, orderItem.UnitPrice, orderItem.CreatedAt, orderItem.UpdatedAt, orderItem.FoodID, orderItem.OrderID)
			if execErr != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": execErr.Error()})
				tx.Rollback()
				defer cancel()
				return
			}
		}

		if commitErr := tx.Commit(); commitErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": commitErr.Error()})
			tx.Rollback()
			defer cancel()
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Items inserted successfully"})
	}
}

func UpdateOrderItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		// var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		// var orderItem models.OrderItem
		// orderItemID := c.Param("order_item_id")

		// filter := bson.M{"order_item_id": orderItemID}
		// var updateObj primitive.D
		// if orderItem.UnitPrice != nil {
		// 	updateObj = append(updateObj, bson.E{Key: "unit_price", Value: orderItem.UnitPrice})
		// }
		// if orderItem.Quantity != nil {
		// 	updateObj = append(updateObj, bson.E{Key: "quantity", Value: orderItem.Quantity})
		// }
		// if orderItem.FoodID != nil {
		// 	updateObj = append(updateObj, bson.E{Key: "food_id", Value: orderItem.FoodID})
		// }

		// orderItem.UpdatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		// updateObj = append(updateObj, bson.E{Key: "updated_at", Value: orderItem.UpdatedAt})

		// upsert := true
		// opt := options.UpdateOptions{
		// 	Upsert: &upsert,
		// }

		// result, err := orderItemCollection.UpdateOne(ctx, filter, bson.D{
		// 	{Key: "$set", Value: updateObj},
		// }, &opt)
		// defer cancel()
		// if err != nil {
		// 	msg := "Order item update failed"
		// 	c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
		// 	return
		// }

		// c.JSON(http.StatusOK, result)

	}
}
