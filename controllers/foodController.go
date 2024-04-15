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
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

func GetFoods() gin.HandlerFunc {
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

		if ISEInjection(c, dbErr, "Error connecting to database") {
			return
		}
		defer dbConn.Close()

		// Fetch total count and menus in one go
		query := `
				SELECT COUNT(*) OVER(), id, name, price, food_image, menu_id, created_at, updated_at
				FROM food
				LIMIT ? OFFSET ?
			`
		foodRows, err := dbConn.QueryContext(ctx, query, recordPerPage, startIndex)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching menus"})
			return
		}
		defer foodRows.Close()

		var totalCount int
		var foodList []models.Food
		for foodRows.Next() {
			var food models.Food
			var createdAtStr string
			var updatedAtStr string
			err := foodRows.Scan(&totalCount, &food.ID, &food.Name, &food.Price, &food.FoodImage, &food.MenuID, &createdAtStr, &updatedAtStr)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error in fetching foods"})
				return
			}

			createdAt, err3 := ParseTime(createdAtStr)
			updatedAt, err4 := ParseTime(updatedAtStr)

			if ISEInjection(c, err3, "Error parsing time strings") || ISEInjection(c, err4, "Error parsing time strings") {
				return
			}

			food.CreatedAt = createdAt
			food.UpdatedAt = updatedAt

			foodList = append(foodList, food)
		}

		if err = foodRows.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error during rows iteration"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"total_count": totalCount,
			"items":       foodList,
		})

	}

}

func GetFood() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		foodID := c.Param("food_id")

		var dbConn, dbErr = db.DBInstanceSql()
		if ISEInjection(c, dbErr, "Error connecting to database") {
			return
		}

		var food models.Food
		var createdAtStr string
		var updatedAtStr string
		row := dbConn.QueryRowContext(ctx, "SELECT * FROM food WHERE id = ?", foodID)
		if err := row.Scan(&food.ID, &food.Name, &food.Price, &food.FoodImage, &createdAtStr, &updatedAtStr, &food.MenuID); err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"message": "Menu not found"})
				return
			}
			fmt.Println(err)
			ISEInjection(c, err, "Error in fetching menu details")
			return
		}
		createdAt, err3 := ParseTime(createdAtStr)
		updatedAt, err4 := ParseTime(updatedAtStr)
		if ISEInjection(c, err3, "Error parsing time strings") || ISEInjection(c, err4, "Error parsing time strings") {
			return
		}

		food.CreatedAt = createdAt
		food.UpdatedAt = updatedAt
		c.JSON(http.StatusOK, food)
	}
}

func CreateFood() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		var dbConn, dbErr = db.DBInstanceSql()
		if ISEInjection(c, dbErr, "Error connecting to database") {
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

		menuDetails := dbConn.QueryRowContext(ctx, "SELECT id, name FROM menu WHERE id = ?", food.MenuID)
		if err := menuDetails.Scan(&menu.ID, &menu.Name); err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"message": "Menu not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error in fetching menu details", "err": err.Error()})
			return
		}
		now := time.Now()
		food.CreatedAt = now
		food.UpdatedAt = now
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

		food.ID = foodID
		c.JSON(http.StatusOK, gin.H{"message": "Food item created successfully", "food": food})
	}
}

// func round(num float64) int {
// 	return int(num + math.Copysign(0.5, num))
// }

// func toFixed(num float64, precision int) float64 {
// 	output := math.Pow(10, float64(precision))
// 	return float64(round(num*output)) / output
// }

func UpdateFood() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		defer cancel()
		var dbConn, dbErr = db.DBInstanceSql()
		if ISEInjection(c, dbErr, "Error connecting to database") {
			return
		}

		var menu models.Menu
		var food models.Food
		foodID := c.Param("food_id")
		if err := c.BindJSON(&food); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		query := "UPDATE food SET updated_at=? "
		values := []interface{}{time.Now()}
		if food.Name != nil {
			query += ", name =? "
			values = append(values, food.Name)
		}

		if food.Price != nil {
			query += ", price =? "
			values = append(values, food.Price)
		}

		if food.FoodImage != nil {
			query += ", food_image =? "
			values = append(values, food.FoodImage)
		}

		if food.MenuID > 0 {

			menuDetails := dbConn.QueryRowContext(ctx, "SELECT id, name FROM menu WHERE id = ?", food.MenuID)
			if err := menuDetails.Scan(&menu.ID, &menu.Name); err != nil {
				if err == sql.ErrNoRows {
					c.JSON(http.StatusNotFound, gin.H{"message": "Menu not found"})
					return
				}
				ISEInjection(c, err, "Error in fetching menu details")
				return
			}
			query += ", menu_id =? "
			values = append(values, food.MenuID)
		}

		query += "WHERE id =?"
		values = append(values, foodID)

		_, err := dbConn.ExecContext(ctx, query, values...)
		if err != nil {
			ISEInjection(c, err, "Error in updating food")
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Food updated successfully"})
	}
}
