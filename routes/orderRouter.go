package routes

import (
	"restaurant_management/controllers"
	"github.com/gin-gonic/gin"
)

func OrderRoutes(incomingRoutes *gin.Engine) {

	incomingRoutes.GET("/order", controllers.GetOrders())
	incomingRoutes.GET("/order/:order_id", controllers.GetOrder())
	incomingRoutes.POST("/order", controllers.CreateFood())
	incomingRoutes.PATCH("/order/:order_id", controllers.UpdateOrder())
}
