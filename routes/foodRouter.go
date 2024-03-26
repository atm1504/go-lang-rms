package routes

import (
	controller "atm1504.in/rms/controllers"
	"github.com/gin-gonic/gin"
)

func FoodRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.GET("/foods/:food_id", controller.GetFood())
}
