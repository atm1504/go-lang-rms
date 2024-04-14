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
)

// var menuCollection *mongo.Collection = database.OpenCollection(database.Client, "menu")

// var dbConn, dbErr = db.DBInstanceSql()

func ParseTime(timeStr string) (time.Time, error) {
	// Parse the time string into a time.Time object
	parsedTime, err := time.Parse("2006-01-02 15:04:05", timeStr)
	if err != nil {
		return time.Time{}, err
	}
	return parsedTime, nil
}

// Function to handle database connection errors
func ISEInjection(c *gin.Context, err error, errorMessage string) bool {
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": errorMessage})
		return true
	}
	return false
}

// var validate = validator.New()
func GetMenus() gin.HandlerFunc {
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

		var totalCount int
		err = dbConn.QueryRowContext(ctx, "SELECT COUNT(*) FROM menu").Scan(&totalCount)
		if ISEInjection(c, dbErr, "Error fetching menu count") {
			return
		}

		rows, err := dbConn.QueryContext(ctx, "SELECT * FROM menu LIMIT ?, ?", startIndex, recordPerPage)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching menus"})
			return
		}
		defer rows.Close()
		var menus []models.Menu
		for rows.Next() {
			var menu models.Menu
			var startDateStr string
			var endDateStr string
			var createdAtStr string
			var updatedAtStr string
			if err := rows.Scan(&menu.ID, &menu.Name, &menu.Category, &startDateStr, &endDateStr, &createdAtStr, &updatedAtStr); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error scanning menu data"})
				return
			}
			startDate, err1 := ParseTime(startDateStr)
			endDate, err2 := ParseTime(endDateStr)
			createdAt, err3 := ParseTime(createdAtStr)
			updatedAt, err4 := ParseTime(updatedAtStr)
			if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error parsing time strings"})
				return
			}
			menu.StartDate = &startDate
			menu.EndDate = &endDate
			menu.CreatedAt = createdAt
			menu.UpdatedAt = updatedAt
			menus = append(menus, menu)
		}
		if err = rows.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error during rows iteration"})
			return
		}

		// Constructing the response
		c.JSON(http.StatusOK, gin.H{
			"total_count": totalCount,
			"menus":       menus,
		})
	}
}

func GetMenu() gin.HandlerFunc {
	return func(c *gin.Context) {

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		var dbConn, dbErr = db.DBInstanceSql()
		if ISEInjection(c, dbErr, "Error connecting to database") {
			return
		}
		// defer dbConn.Close()
		var menu models.Menu
		var startDateStr string
		var endDateStr string
		var createdAtStr string
		var updatedAtStr string

		menuID := c.Param("menu_id")
		fmt.Println("Menu id is: ", menuID)

		// err := menuCollection.FindOne(ctx, bson.M{"menu_id": menuID}).Decode(&menu)
		row := dbConn.QueryRowContext(ctx, "SELECT * FROM menu WHERE id = ?", menuID)
		if err := row.Scan(&menu.ID, &menu.Name, &menu.Category, &startDateStr, &endDateStr, &createdAtStr, &updatedAtStr); err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"message": "Menu not found"})
				return
			}
			fmt.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error in fetching menu details"})
			return
		}

		startDate, err := ParseTime(startDateStr)
		endDate, err := ParseTime(endDateStr)
		createdAt, err1 := ParseTime(createdAtStr)
		updatedAt, err1 := ParseTime(updatedAtStr)
		if err != nil || err1 != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error parsing time strings"})
			return
		}
		menu.StartDate = &startDate
		menu.EndDate = &endDate
		menu.CreatedAt = createdAt
		menu.UpdatedAt = updatedAt

		c.JSON(http.StatusOK, menu)
	}
}

func CreateMenu() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		var dbConn, dbErr = db.DBInstanceSql()
		if dbErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error connecting to database"})
			return
		}

		var menu models.Menu
		if err := c.BindJSON(&menu); err != nil {
			defer cancel()
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validationErr := validate.Struct(menu)
		if validationErr != nil {
			defer cancel()

			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}
		now := time.Now()
		menu.CreatedAt = now
		menu.UpdatedAt = now
		// Perform the insertion into the database
		result, err := dbConn.ExecContext(ctx, "INSERT INTO menu (name, category, start_date, end_date, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
			menu.Name, menu.Category, menu.StartDate, menu.EndDate, menu.CreatedAt, menu.UpdatedAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert food item"})
			return
		}

		menuID, err := result.LastInsertId()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get menu item ID"})
			return
		}
		menu.ID = menuID
		c.JSON(http.StatusOK, gin.H{"message": "Menu item created successfully", "menu": menu})
	}
}

func inTimeSpan(start, end, check time.Time) bool {
	return check.After(start) && check.Before(end)
}

func UpdateMenu() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		db, err := db.DBInstanceSql()
		if ISEInjection(c, err, "Error connecting to database") {
			return
		}
		defer db.Close()

		var menu models.Menu
		menuID := c.Param("menu_id")

		if err := c.BindJSON(&menu); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if menu.StartDate != nil && menu.EndDate != nil {
			if !inTimeSpan(*menu.StartDate, *menu.EndDate, time.Now()) {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Kindly enter valid time"})
				defer cancel()
				return
			}
			query := "UPDATE menu SET start_date=?, end_date=?"
			values := []interface{}{menu.StartDate, menu.EndDate}

			if menu.Name != "" {
				query += ", name=?"
				values = append(values, menu.Name)
			}

			if menu.Category != "" {
				query += ", category=?"
				values = append(values, menu.Category)
			}

			query += " WHERE id=?"
			values = append(values, menuID)

			// Execute the SQL UPDATE query
			_, err := db.ExecContext(ctx, query, values...)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error in updating menu"})
				return
			}

			c.JSON(http.StatusOK, gin.H{"message": "Menu updated successfully"})
			return
		}
		defer cancel()
		c.JSON(http.StatusBadRequest, gin.H{"error": "Start date and end date are required"})

	}
}

func connectDB() {
	panic("unimplemented")
}
