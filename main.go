package main

import (
	"os"

	"log"

	routes "atm1504.in/rms/routes"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// var foodCollection *mongo.Collection = database.OpenCollection(database.Client, "food")

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	router := gin.New()
	router.Use(gin.Logger())
	routes.UserRoutes(router)
	routes.FoodRoutes(router)
	routes.MenuRoutes(router)
	routes.OrderRoutes(router)
	routes.TableRoutes(router)
	routes.OrderItemRoutes(router)
	// router.Use(middleware.Authentication())

	runErr := router.Run(":" + port)
	if runErr != nil {
		// Handle the error, for example, log it or return it
		log.Fatalf("Error starting server: %v", runErr)
	}
}
