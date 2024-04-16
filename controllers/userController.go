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
	"golang.org/x/crypto/bcrypt"
)

// var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")

func GetUsers() gin.HandlerFunc {
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
			err := invoiceRows.Scan(&totalCount, &invoice.ID, &invoice.OrderID, &invoice.PaymentMethod, &invoice.PaymentStatus, &invoice.PaymentDueDate, &createdAtStr, &updatedAtStr)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error in fetching invoices"})
				return
			}
			createdAt, err3 := ParseTime(createdAtStr)
			updatedAt, err4 := ParseTime(updatedAtStr)

			if helper.ISEInjection(c, err3, "Error parsing time strings") || helper.ISEInjection(c, err4, "Error parsing time strings") {
				return
			}
			invoice.CreatedAt = createdAt
			invoice.UpdatedAt = updatedAt
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

func GetUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		var dbConn, dbErr = db.DBInstanceSql()
		if helper.ISEInjection(c, dbErr, "Error connecting to database") {
			return
		}
		defer dbConn.Close()
		var user models.User
		var createdAtStr string
		var updatedAtStr string

		userID := c.Param("user_id")

		row := dbConn.QueryRowContext(ctx, "SELECT id, first_name, last_name, password, email, avatar, phone, created_at, updated_at FROM user WHERE id = ?", userID)
		if err := row.Scan(&user.ID, &user.FirstName, &user.LastName, &user.Password, &user.Email, &user.Avatar, &user.Phone, &createdAtStr, &updatedAtStr); err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"message": "Menu not found"})
				return
			}
			fmt.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error in fetching menu details"})
			return
		}

		createdAt, err3 := ParseTime(createdAtStr)
		updatedAt, err4 := ParseTime(updatedAtStr)

		if helper.ISEInjection(c, err3, "Error parsing time strings") || helper.ISEInjection(c, err4, "Error parsing time strings") {
			return
		}
		user.CreatedAt = createdAt
		user.UpdatedAt = updatedAt
		c.JSON(http.StatusOK, user)
	}
}

func SignUp() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		var user models.User
		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validationErr := validate.Struct(user)
		if BadRequestInjection(c, validationErr, "User structure is incorrect") {
			return
		}
		var dbConn, dbErr = db.DBInstanceSql()
		if helper.ISEInjection(c, dbErr, "Error connecting to database") {
			return
		}
		defer cancel()

		var emailPhoneCountValue int
		emailPhoneCount := dbConn.QueryRowContext(ctx, "SELECT COUNT(*) as count FROM user WHERE email =? OR phone =?", user.Email, user.Phone)
		if err := emailPhoneCount.Scan(&emailPhoneCountValue); err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"message": "Menu not found"})
				return
			}
			helper.ISEInjection(c, err, "Error in fetching menu details")
			return
		}

		password := HashPassword(*user.Password)
		user.Password = &password

		var now = time.Now()
		user.CreatedAt = now
		user.UpdatedAt = now

		token, refreshToken, _ := helper.GenerateAllTokens(*user.Email, *user.FirstName, *user.LastName, *user.Phone)
		user.Token = &token
		user.RefreshToken = &refreshToken

		result, err := dbConn.ExecContext(ctx, "INSERT INTO user (first_name, last_name, password, email, avatar, phone, token, refresh_token, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
			user.FirstName, user.LastName, user.Password, user.Email, user.Avatar, user.Phone, user.Token, user.RefreshToken, user.CreatedAt, user.UpdatedAt)

		if helper.ISEInjection(c, err, "Failed to insert user item") {
			return
		}
		userID, err := result.LastInsertId()

		if helper.ISEInjection(c, err, "Failed to get user item ID") {
			return
		}
		user.ID = userID
		c.JSON(http.StatusOK, gin.H{"message": "user item created successfully", "item": user})
	}
}

func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var user models.User
		var foundUser models.User
		defer cancel()
		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		var dbConn, dbErr = db.DBInstanceSql()
		if helper.ISEInjection(c, dbErr, "Error connecting to database") {
			return
		}
		defer cancel()
		query := "SELECT id, first_name, last_name, password, email, phone FROM user WHERE email =?"
		row := dbConn.QueryRowContext(ctx, query, user.Email)

		if err := row.Scan(&foundUser.ID, &foundUser.FirstName, &foundUser.LastName, &foundUser.Password, &foundUser.Email, &foundUser.Phone); err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
				return
			}
			helper.ISEInjection(c, err, "Error in fetching user details")
			return
		}

		passwordIsValid, msg := VerifyPassword(*user.Password, *foundUser.Password)
		defer cancel()
		if !passwordIsValid {
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		token, refreshToken, _ := helper.GenerateAllTokens(*foundUser.Email, *foundUser.FirstName, *foundUser.LastName, *foundUser.Phone)
		helper.UpdateAllTokens(token, refreshToken, foundUser.ID)
		c.JSON(http.StatusOK, foundUser)
	}
}

func HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Panic(err)
	}

	return string(bytes)
}

func VerifyPassword(userPassword string, providedPassword string) (bool, string) {

	err := bcrypt.CompareHashAndPassword([]byte(providedPassword), []byte(userPassword))
	check := true
	msg := ""

	if err != nil {
		msg = "login or password is incorrect"
		check = false
	}
	return check, msg
}
