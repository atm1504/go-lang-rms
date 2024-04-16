package controller

import (
	"context"
	"database/sql"
	"net/http"
	"strconv"
	"time"

	db "atm1504.in/rms/database"
	helper "atm1504.in/rms/helpers"

	"atm1504.in/rms/models"
	"github.com/gin-gonic/gin"
)

type InvoiceViewFormat struct {
	InvoiceID      int64       `bson:"invoice_id" json:"order_items"`
	PaymentMethod  string      `bson:"payment_method" json:"payment_method"`
	OrderID        int64       `bson:"order_id" json:"order_id"`
	PaymentStatus  *string     `bson:"payment_status" json:"payment_status"`
	PaymentDue     interface{} `bson:"payment_due" json:"payment_due"`
	TableNumber    interface{} `bson:"table_number" json:"table_number"`
	PaymentDueDate time.Time   `bson:"payment_due_date" json:"payment_due_date"`
	OrderDetails   interface{} `bson:"order_details" json:"order_details"`
}

func GetInvoices() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
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

		query := `
				SELECT COUNT(*) OVER(), id, order_id, payment_method, payment_status, payment_due_date, created_at, updated_at
				FROM invoice
				LIMIT ? OFFSET ?
			`
		invoiceRows, err := dbConn.QueryContext(ctx, query, recordPerPage, startIndex)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching menus"})
			return
		}
		defer invoiceRows.Close()

		var totalCount int
		var invoiceList []models.Invoice
		for invoiceRows.Next() {
			var invoice models.Invoice
			var createdAtStr string
			var updatedAtStr string
			var paymentDueDateStr string
			err := invoiceRows.Scan(&totalCount, &invoice.ID, &invoice.OrderID, &invoice.PaymentMethod, &invoice.PaymentStatus, &paymentDueDateStr, &createdAtStr, &updatedAtStr)

			if helper.ISEInjection(c, err, "Error in fetching invoices") {
				return
			}
			createdAt, err3 := ParseTime(createdAtStr)
			updatedAt, err4 := ParseTime(updatedAtStr)
			paymentDueDate, err2 := ParseTime(paymentDueDateStr)

			if helper.ISEInjection(c, err3, "Error parsing time strings") || helper.ISEInjection(c, err4, "Error parsing time strings") || helper.ISEInjection(c, err2, "Error parsing time strings") {
				return
			}
			invoice.CreatedAt = createdAt
			invoice.UpdatedAt = updatedAt
			invoice.PaymentDueDate = paymentDueDate
			invoiceList = append(invoiceList, invoice)
		}
		if err = invoiceRows.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error during rows iteration"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"total_count": totalCount,
			"items":       invoiceList,
		})
	}
}

func GetInvoice() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		invoiceID := c.Param("invoice_id")
		defer cancel()

		dbConn, dbErr := db.DBInstanceSql()
		if helper.ISEInjection(c, dbErr, "Error connecting to database") {
			return
		}
		defer dbConn.Close()

		query := `SELECT id, order_id, payment_method, payment_status, payment_due_date, created_at, updated_at FROM invoice where id = ?`

		var invoice models.Invoice
		var createdAtStr string
		var updatedAtStr string
		var paymentDueDateStr string
		row := dbConn.QueryRowContext(ctx, query, invoiceID)
		if err := row.Scan(&invoice.ID, &invoice.OrderID, &invoice.PaymentMethod, &invoice.PaymentStatus, &paymentDueDateStr, &createdAtStr, &updatedAtStr); err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"message": "Menu not found"})
				return
			}
			helper.ISEInjection(c, err, "Error in fetching menu details")
			return
		}
		createdAt, err3 := ParseTime(createdAtStr)
		updatedAt, err4 := ParseTime(updatedAtStr)
		paymentDueDate, err2 := ParseTime(paymentDueDateStr)
		if helper.ISEInjection(c, err3, "Error parsing time strings") || helper.ISEInjection(c, err4, "Error parsing time strings") || helper.ISEInjection(c, err2, "Error parsing time strings") {
			return
		}

		invoice.CreatedAt = createdAt
		invoice.UpdatedAt = updatedAt
		invoice.PaymentDueDate = paymentDueDate

		var invoiceView InvoiceViewFormat
		allOrderItems, err := ItemsByOrder(c, invoice.OrderID)

		if err != nil {
			defer cancel()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while listing order items by order ID"})
			return
		}
		invoiceView.OrderID = invoice.OrderID
		invoiceView.PaymentDueDate = invoice.PaymentDueDate
		invoiceView.PaymentMethod = "null"
		if invoice.PaymentMethod != nil {
			invoiceView.PaymentMethod = *invoice.PaymentMethod
		}

		invoiceView.InvoiceID = invoice.ID
		invoiceView.PaymentStatus = invoice.PaymentStatus
		invoiceView.PaymentDue = allOrderItems[0]["payment_due"]
		invoiceView.TableNumber = allOrderItems[0]["table_number"]
		invoiceView.OrderDetails = allOrderItems[0]["order_items"]
		defer cancel()
		c.JSON(http.StatusOK, invoiceView)
	}
}

func CreateInvoice() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var invoice models.Invoice
		defer cancel()
		if err := c.BindJSON(&invoice); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var dbConn, dbErr = db.DBInstanceSql()
		if helper.ISEInjection(c, dbErr, "Error connecting to database") {
			return
		}
		defer cancel()

		var order models.Order

		orderDetails := dbConn.QueryRowContext(ctx, "SELECT id FROM orders WHERE id = ?", invoice.OrderID)
		if err := orderDetails.Scan(&order.ID); err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"message": "Order not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error in fetching order details", "err": err.Error()})
			return
		}

		status := "PENDING"
		if invoice.PaymentStatus == nil {
			invoice.PaymentStatus = &status
		}

		invoice.PaymentDueDate = time.Now().AddDate(0, 0, 1)
		invoice.CreatedAt = time.Now()
		invoice.UpdatedAt = time.Now()

		validationErr := validate.Struct(invoice)
		if validationErr != nil {
			defer cancel()
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		result, err := dbConn.ExecContext(ctx, "INSERT INTO invoice (order_id, payment_method, payment_status, payment_due_date, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
			invoice.OrderID, invoice.PaymentMethod, invoice.PaymentStatus, invoice.PaymentDueDate, invoice.CreatedAt, invoice.UpdatedAt)
		if helper.ISEInjection(c, err, "Failed to insert invoice item") {
			return
		}
		invoiceID, err := result.LastInsertId()
		if helper.ISEInjection(c, err, "Failed to get invoice item ID") {
			return
		}

		invoice.ID = invoiceID
		c.JSON(http.StatusOK, gin.H{"message": "Food item created successfully", "items": invoice})
	}
}

func UpdateInvoice() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var invoice models.Invoice
		invoiceID := c.Param("invoice_id")
		defer cancel()

		if err := c.BindJSON(&invoice); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var dbConn, dbErr = db.DBInstanceSql()
		if helper.ISEInjection(c, dbErr, "Error connecting to database") {
			return
		}
		defer cancel()

		query := "UPDATE invoice SET updated_at=? "
		values := []interface{}{time.Now()}

		if invoice.PaymentMethod != nil {
			query += ", payment_method =? "
			values = append(values, invoice.PaymentMethod)
		}

		if invoice.PaymentStatus != nil {
			query += ", payment_status =? "
			values = append(values, invoice.PaymentStatus)
		}

		query += "WHERE id =?"
		values = append(values, invoiceID)

		result, err := dbConn.ExecContext(ctx, query, values...)
		if err != nil {
			helper.ISEInjection(c, err, "Error in updating invoice")
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Invoice updated successfully", "items": result})
	}
}
