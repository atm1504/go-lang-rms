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

func GetTables() gin.HandlerFunc {
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
		query := `SELECT COUNT(*) OVER(), id, number_of_guests, table_number, created_at, updated_at FROM tables LIMIT ? OFFSET ?`

		fmt.Println(query)
		tableRows, err := dbConn.QueryContext(ctx, query, recordPerPage, startIndex)
		if ISEInjection(c, err, "Error fetching tables") {
			return
		}
		defer tableRows.Close()

		var totalCount int
		var tableList []models.Table
		for tableRows.Next() {
			var table models.Table
			var createdAtStr string
			var updatedAtStr string
			err := tableRows.Scan(&totalCount, &table.ID, &table.NumberOfGuests, &table.TableNumber, &createdAtStr, &updatedAtStr)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error in fetching tables"})
				return
			}

			createdAt, err3 := ParseTime(createdAtStr)
			updatedAt, err4 := ParseTime(updatedAtStr)

			if ISEInjection(c, err3, "Error parsing time strings") || ISEInjection(c, err4, "Error parsing time strings") {
				return
			}

			table.CreatedAt = createdAt
			table.UpdatedAt = updatedAt

			tableList = append(tableList, table)
		}

		if err = tableRows.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error during rows iteration"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"total_count": totalCount,
			"items":       tableList,
		})

	}
}

func GetTable() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		tableID := c.Param("table_id")

		var dbConn, dbErr = db.DBInstanceSql()
		if ISEInjection(c, dbErr, "Error connecting to database") {
			return
		}

		var table models.Table
		var createdAtStr string
		var updatedAtStr string
		row := dbConn.QueryRowContext(ctx, "SELECT id, number_of_guests, table_number, created_at, updated_at  FROM tables WHERE id = ?", tableID)
		if err := row.Scan(&table.ID, &table.NumberOfGuests, &table.TableNumber, &createdAtStr, &updatedAtStr); err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"message": "Table not found"})
				return
			}
			fmt.Println(err)
			ISEInjection(c, err, "Error in fetching table details")
			return
		}

		createdAt, err3 := ParseTime(createdAtStr)
		updatedAt, err4 := ParseTime(updatedAtStr)
		if ISEInjection(c, err3, "Error parsing time strings") || ISEInjection(c, err4, "Error parsing time strings") {
			return
		}

		table.CreatedAt = createdAt
		table.UpdatedAt = updatedAt

		c.JSON(http.StatusOK, table)
	}
}

func CreateTable() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		var dbConn, dbErr = db.DBInstanceSql()
		if ISEInjection(c, dbErr, "Error connecting to database") {
			return
		}

		var table models.Table

		if err := c.BindJSON(&table); err != nil {
			defer cancel()
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validationErr := validate.Struct(table)
		if BadRequestInjection(c, validationErr, "Error in create table payload") {
			return
		}

		now := time.Now()
		table.CreatedAt = now
		table.UpdatedAt = now

		result, err := dbConn.ExecContext(ctx, "INSERT INTO tables (number_of_guests, table_number, created_at, updated_at) VALUES (?, ?, ?, ?)",
			table.NumberOfGuests, table.TableNumber, table.CreatedAt, table.UpdatedAt)

		if ISEInjection(c, err, "Failed to insert table item") {
			return
		}
		tableID, err := result.LastInsertId()

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get table item ID"})
			return
		}

		table.ID = tableID
		c.JSON(http.StatusOK, gin.H{"message": "table item created successfully", "table": table})
	}
}

func UpdateTable() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		var dbConn, dbErr = db.DBInstanceSql()
		if ISEInjection(c, dbErr, "Error connecting to database") {
			return
		}
		var table models.Table
		tableID := c.Param("table_id")

		if err := c.BindJSON(&table); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

			return
		}

		updateQuery := "UPDATE tables SET updated_at=? "
		updateValues := []interface{}{time.Now()}

		if table.NumberOfGuests != nil {
			updateQuery += ", number_of_guests =? "
			updateValues = append(updateValues, table.NumberOfGuests)
		}
		if table.TableNumber != nil {
			updateQuery += ", table_number =? "
			updateValues = append(updateValues, table.TableNumber)
		}

		updateQuery += "WHERE id =?"
		updateValues = append(updateValues, tableID)

		_, err := dbConn.ExecContext(ctx, updateQuery, updateValues...)
		if err != nil {
			ISEInjection(c, err, "Error in updating table")
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Table updated successfully"})
	}
}
