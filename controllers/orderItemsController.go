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
		orderID := c.Param("order_id")

		allOrderItems, err := ItemsByOrder(c, orderID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while listing order items by order ID"})
			return
		}
		c.JSON(http.StatusOK, allOrderItems)
	}
}

func ItemsByOrder(c *gin.Context, orderID string) (OrderItems []primitive.M, err error) {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	dbConn, dbErr := db.DBInstanceSql()
	if ISEInjection(c, dbErr, "Error connecting to database") {
		return
	}
	defer dbConn.Close()

	const fetchQuery = `SELECT ot.id, ot.quantity, ot.food_id, ot.order_id , tt.table_number, tt.number_of_guests,  f.price, f.name, f.food_image
	FROM order_item ot
	JOIN food f on f.id=ot.food_id
	JOIN orders oo on oo.id=ot.order_id
	JOIN tables tt on tt.id=oo.id 
	WHERE ot.order_id=?`

	orderItemsByRow, err := dbConn.QueryContext(ctx, fetchQuery, orderID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching menus"})
		return
	}
	defer orderItemsByRow.Close()

	var items []primitive.M
	for orderItemsByRow.Next() {
		var id, foodID, orderID int64
		var quantity, name, foodImage *string
		var tableNumber, numberOfGuests int
		var price *float64

		err = orderItemsByRow.Scan(&id, &quantity, &foodID, &orderID, &tableNumber, &numberOfGuests, &price, &name, &foodImage)
		if ISEInjection(c, err, "Error scanning order items") {
			return nil, err
		}
		item := primitive.M{
			"id":               id,
			"quantity":         quantity,
			"food_id":          foodID,
			"order_id":         orderID,
			"table_number":     tableNumber,
			"number_of_guests": numberOfGuests,
			"price":            price,
			"name":             name,
			"food_image":       foodImage,
		}

		items = append(items, item)
	}
	return items, err
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
			orderItem, ok := item.(models.OrderItem)
			if !ok {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assert order item type"})
				tx.Rollback()
				defer cancel()
				return
			}
			_, execErr := stmt.ExecContext(ctx, orderItem.Quantity, *orderItem.UnitPrice, orderItem.CreatedAt, orderItem.UpdatedAt, orderItem.FoodID, orderItem.OrderID)
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
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var orderItem models.OrderItem
		orderItemID := c.Param("order_item_id")

		defer cancel()
		var dbConn, dbErr = db.DBInstanceSql()
		if ISEInjection(c, dbErr, "Error connecting to database") {
			return
		}

		if err := c.BindJSON(&orderItem); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		now := time.Now()
		query := "UPDATE order_item SET updated_at=? "
		values := []interface{}{now}

		if orderItem.UnitPrice != nil {
			var num = toFixed(*orderItem.UnitPrice, 2)
			query += ", unit_price=? "
			values = append(values, &num)
		}
		if orderItem.Quantity != nil {
			query += ", quantity=? "
			values = append(values, *orderItem.Quantity)
		}
		if orderItem.FoodID != 0 {
			query += ", food_id=? "
			values = append(values, orderItem.FoodID)
		}

		query += "WHERE id =?"
		values = append(values, orderItemID)
		result, err := dbConn.ExecContext(ctx, query, values...)
		if err != nil {
			ISEInjection(c, err, "Error in updating food")
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Order item updated successfully", "item": result})
	}
}
