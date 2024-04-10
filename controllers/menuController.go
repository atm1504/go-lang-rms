package controller

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"atm1504.in/rms/database"
	db "atm1504.in/rms/database"
	"atm1504.in/rms/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var menuCollection *mongo.Collection = database.OpenCollection(database.Client, "menu")

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
func HandleDBError(c *gin.Context, err error, errorMessage string) bool {
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
				{Key: "menus", Value: bson.D{{Key: "$slice", Value: []interface{}{"$data", startIndex, recordPerPage}}}},
			}},
		}

		result, err := menuCollection.Aggregate(ctx, mongo.Pipeline{matchStage, groupStage, projectStage})
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occurred while listing menus"})
			return
		}

		var allMenus []bson.M
		if err = result.All(ctx, &allMenus); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occurred while processing menus"})
			log.Fatal(err)
		}

		// Assuming you want to return the list of menus directly
		if len(allMenus) > 0 {
			c.JSON(http.StatusOK, allMenus[0])
		} else {
			c.JSON(http.StatusOK, []interface{}{}) // Return an empty array if no menus
		}
	}
}

func GetMenu() gin.HandlerFunc {
	return func(c *gin.Context) {

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		var dbConn, dbErr = db.DBInstanceSql()
		if dbErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error connecting to database"})
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
		// defer dbConn.Close()

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
		var menu models.Menu

		menuID := c.Param("menu_id")
		filter := bson.M{"menu_id": menuID}
		if err := c.BindJSON(&menu); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			defer cancel()
			return
		}

		var updateObj primitive.D
		if menu.StartDate != nil && menu.EndDate != nil {
			if !inTimeSpan(*menu.StartDate, *menu.EndDate, time.Now()) {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Kindly enter valid time"})
				defer cancel()
				return
			}
			updateObj = append(updateObj, bson.E{Key: "start_date", Value: menu.StartDate})
			updateObj = append(updateObj, bson.E{Key: "end_date", Value: menu.EndDate})

			if menu.Name != "" {
				updateObj = append(updateObj, bson.E{Key: "name", Value: menu.Name})
			}

			if menu.Category != "" {
				updateObj = append(updateObj, bson.E{Key: "category", Value: menu.Category})
			}

			upsert := true
			opt := options.UpdateOptions{Upsert: &upsert}

			result, err := menuCollection.UpdateOne(
				ctx,
				filter,
				bson.D{
					{Key: "$set", Value: updateObj},
				},
				&opt,
			)
			defer cancel()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error in updating menu"})
				return
			}
			c.JSON(http.StatusOK, result)
		}
		defer cancel()
	}
}
