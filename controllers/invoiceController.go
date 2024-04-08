package controller

import (
	"context"
	"log"
	"net/http"
	"time"

	"atm1504.in/rms/database"
	"atm1504.in/rms/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type InvoiceViewFormat struct {
	InvoiceID      string      `bson:"invoice_id" json:"order_items"`
	PaymentMethod  string      `bson:"payment_method" json:"payment_method"`
	OrderID        string      `bson:"order_id" json:"order_id"`
	PaymentStatus  *string     `bson:"payment_status" json:"payment_status"`
	PaymentDue     interface{} `bson:"payment_due" json:"payment_due"`
	TableNumber    interface{} `bson:"table_number" json:"table_number"`
	PaymentDueDate time.Time   `bson:"payment_due_date" json:"payment_due_date"`
	OrderDetails   interface{} `bson:"order_details" json:"order_details"`
}

var invoiceCollection *mongo.Collection = database.OpenCollection(database.Client, "invoice")

func GetInvoices() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		result, err := invoiceCollection.Find(context.TODO(), bson.M{})
		defer cancel()
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusNotFound, gin.H{
					"message": "Menu not found",
				})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while listing invoice items"})
			return
		}

		var allInvoices []bson.M
		if err = result.All(ctx, &allInvoices); err != nil {
			log.Fatal(err)
		}
		c.JSON(http.StatusOK, allInvoices)
	}
}

func GetInvoice() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		invoiceID := c.Param("invoice_id")

		var invoice models.Invoice
		err := invoiceCollection.FindOne(ctx, bson.M{"invoice_id": invoiceID}).Decode(&invoice)
		if err != nil {
			defer cancel()
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusNotFound, gin.H{
					"message": "Invoice not found",
				})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while listing invoice items"})
			return
		}

		var invoiceView InvoiceViewFormat
		allOrderItems, err := ItemsByOrder(invoice.OrderID)

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

		invoiceView.InvoiceID = invoice.InvoiceID
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

		if err := c.BindJSON(&invoice); err != nil {
			defer cancel()
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var order models.Order

		err := orderCollection.FindOne(ctx, bson.M{"order_id": invoice.OrderID}).Decode(&order)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "message: Order was not found"})
			return
		}
		status := "PENDING"
		if invoice.PaymentStatus == nil {
			invoice.PaymentStatus = &status
		}

		invoice.PaymentDueDate, _ = time.Parse(time.RFC3339, time.Now().AddDate(0, 0, 1).Format(time.RFC3339))
		invoice.CreatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		invoice.UpdatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		invoice.ID = primitive.NewObjectID()
		invoice.InvoiceID = invoice.ID.Hex()

		validationErr := validate.Struct(invoice)
		if validationErr != nil {
			defer cancel()
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		result, insertErr := invoiceCollection.InsertOne(ctx, invoice)
		defer cancel()
		if insertErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invoice item was not created"})
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func UpdateInvoice() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var invoice models.Invoice
		invoiceID := c.Param("invoice_id")

		if err := c.BindJSON(&invoice); err != nil {
			defer cancel()
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		filter := bson.M{"invoice_id": invoiceID}
		var updateObj primitive.D
		if invoice.PaymentMethod != nil {
			updateObj = append(updateObj, bson.E{Key: "payment_method", Value: invoice.PaymentMethod})
		}

		if invoice.PaymentStatus != nil {
			updateObj = append(updateObj, bson.E{Key: "payment_status", Value: invoice.PaymentStatus})
		}

		invoice.UpdatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		updateObj = append(updateObj, bson.E{Key: "updated_at", Value: invoice.UpdatedAt})

		upsert := true
		opt := options.UpdateOptions{
			Upsert: &upsert,
		}

		status := "PENDING"
		if invoice.PaymentStatus == nil {
			invoice.PaymentStatus = &status
		}

		result, err := invoiceCollection.UpdateOne(
			ctx,
			filter,
			bson.D{
				{Key: "$set", Value: updateObj},
			},
			&opt,
		)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invoice item update failed"})
			return
		}
		c.JSON(http.StatusOK, result)
	}
}
