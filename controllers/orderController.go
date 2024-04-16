package controller

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	db "atm1504.in/rms/database"
	helper "atm1504.in/rms/helpers"
	"atm1504.in/rms/models"
	"github.com/gin-gonic/gin"
)

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
		dbConn, dbErr := db.DBInstanceSql()

		if helper.ISEInjection(c, dbErr, "Error connecting to database") {
			return
		}
		defer dbConn.Close()
		query := `SELECT COUNT(*) OVER(), id, order_date, created_at, updated_at, table_id FROM orders LIMIT ? OFFSET ?`

		fmt.Println(query)
		orderRows, err := dbConn.QueryContext(ctx, query, recordPerPage, startIndex)
		if helper.ISEInjection(c, err, "Error fetching tables") {
			return
		}

		defer orderRows.Close()

		var totalCount int
		var orderList []models.Order
		for orderRows.Next() {
			var orderObj models.Order
			var createdAtStr string
			var updatedAtStr string
			var orderDateAtStr string
			err := orderRows.Scan(&totalCount, &orderObj.ID, &orderDateAtStr, &createdAtStr, &updatedAtStr, &orderObj.TableID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error in fetching tables"})
				return
			}

			createdAt, err3 := ParseTime(createdAtStr)
			updatedAt, err4 := ParseTime(updatedAtStr)
			orderDate, err5 := ParseTime(orderDateAtStr)

			if helper.ISEInjection(c, err3, "Error parsing time strings") || helper.ISEInjection(c, err4, "Error parsing time strings") || helper.ISEInjection(c, err5, "Error parsing time strings") {
				return
			}

			orderObj.CreatedAt = createdAt
			orderObj.UpdatedAt = updatedAt
			orderObj.OrderDate = &orderDate

			orderList = append(orderList, orderObj)
		}

		if err = orderRows.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error during rows iteration"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"total_count": totalCount,
			"items":       orderList,
		})

	}
}

func GetOrder() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		var order models.Order

		orderID := c.Param("order_id")
		fmt.Println("Order id is: ", orderID)

		var dbConn, dbErr = db.DBInstanceSql()
		if helper.ISEInjection(c, dbErr, "Error connecting to database") {
			return
		}

		var createdAtStr string
		var orderDateStr string
		var updatedAtStr string

		row := dbConn.QueryRowContext(ctx, "SELECT id, order_date, created_at, updated_at, table_id  FROM orders WHERE id = ?", orderID)
		if err := row.Scan(&order.ID, &orderDateStr, &createdAtStr, &updatedAtStr, &order.TableID); err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"message": "Order not found"})
				return
			}
			fmt.Println(err)
			helper.ISEInjection(c, err, "Error in fetching order details")
			return
		}

		createdAt, err3 := ParseTime(createdAtStr)
		updatedAt, err4 := ParseTime(updatedAtStr)
		orderDate, err5 := ParseTime(orderDateStr)

		if helper.ISEInjection(c, err3, "Error parsing time strings") || helper.ISEInjection(c, err4, "Error parsing time strings") || helper.ISEInjection(c, err5, "Error parsing time strings") {
			return
		}

		order.CreatedAt = createdAt
		order.UpdatedAt = updatedAt
		order.OrderDate = &orderDate

		c.JSON(http.StatusOK, order)
	}
}

func CreateOrder() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		var dbConn, dbErr = db.DBInstanceSql()
		if helper.ISEInjection(c, dbErr, "Error connecting to database") {
			return
		}
		var table models.Table
		var order models.Order

		if err := c.BindJSON(&order); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "msg": "I failed while destructuring"})
			return
		}

		validationErr := validate.Struct(order)
		if BadRequestInjection(c, validationErr, "Error in create order payload") {
			return
		}

		if order.TableID != 0 {
			tableDetails := dbConn.QueryRowContext(ctx, "SELECT id FROM tables WHERE id = ?", order.TableID)
			if err := tableDetails.Scan(&table.ID); err != nil {
				if err == sql.ErrNoRows {
					c.JSON(http.StatusNotFound, gin.H{"message": "Table not found"})
					return
				}
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error in fetching table details", "err": err.Error()})
				return
			}
		}
		now := time.Now()
		order.CreatedAt = now
		order.UpdatedAt = now
		result, err := dbConn.ExecContext(ctx, "INSERT INTO orders (order_date, created_at, updated_at, table_id) VALUES (?, ?, ?, ?)",
			order.OrderDate, order.CreatedAt, order.UpdatedAt, order.TableID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert order", "err": err.Error()})
			return
		}
		orderID, err := result.LastInsertId()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get order ID", "err": err.Error()})
			return
		}
		order.ID = orderID
		c.JSON(http.StatusOK, gin.H{"message": "Food item created successfully", "item": order})
	}
}

func UpdateOrder() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		var table models.Table
		var order models.Order

		orderID := c.Param("order_id")
		defer cancel()
		var dbConn, dbErr = db.DBInstanceSql()
		if helper.ISEInjection(c, dbErr, "Error connecting to database") {
			return
		}

		if err := c.BindJSON(&order); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if order.TableID != 0 {
			tableDetails := dbConn.QueryRowContext(ctx, "SELECT id FROM tables WHERE id = ?", order.TableID)
			if err := tableDetails.Scan(&table.ID); err != nil {
				if err == sql.ErrNoRows {
					c.JSON(http.StatusNotFound, gin.H{"message": "Table not found"})
					return
				}
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error in fetching table details", "err": err.Error()})
				return
			}
		}

		now := time.Now()
		query := "UPDATE orders SET updated_at=? "
		values := []interface{}{now}
		if order.TableID != 0 {
			query += ", table_id=? "
			values = append(values, order.TableID)
		}

		if order.OrderDate != nil {
			query += ", order_date=? "
			values = append(values, order.OrderDate)
		}

		query += "WHERE id =?"
		values = append(values, orderID)
		result, err := dbConn.ExecContext(ctx, query, values...)
		if err != nil {
			helper.ISEInjection(c, err, "Error in updating food")
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Food updated successfully", "item": result})
	}
}

func OrderItemOrderCreator(order models.Order, currentTime time.Time, dbConn *sql.DB, ctx context.Context) int64 {
	order.CreatedAt = currentTime
	order.UpdatedAt = currentTime
	result, err := dbConn.ExecContext(ctx, "INSERT INTO orders (order_date, created_at, updated_at, table_id) VALUES (?, ?, ?, ?)",
		order.OrderDate, order.CreatedAt, order.UpdatedAt, order.TableID)
	if err != nil {
		log.Fatal(err)
	}
	orderID, err := result.LastInsertId()
	if err != nil {
		log.Fatal(err)
	}
	order.ID = orderID
	return orderID
}
