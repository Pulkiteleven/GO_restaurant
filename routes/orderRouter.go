package routes

import(
	"github.com/gin-gonic/gin"
	
	controller "golang-restaurant-management/controllers"
)

func OrderRoutes(incomingRoutes *gin.Engine){

	incomingRoutes.GET("/orders", controller.GetOrders())
	incomingRoutes.GET("/orders/:order_id", controller.GetOrder())
	incomingRoutes.POST("/orders", controller.CreateOrder())
	incomingRoutes.PATCH("/orders/:order_id", controller.UpdateOrder())
	incomingRoutes.DELETE("/order/:order_id", controller.DeleteOrder())
	incomingRoutes.DELETE("/allorders", controller.DeleteAllOrders())
}