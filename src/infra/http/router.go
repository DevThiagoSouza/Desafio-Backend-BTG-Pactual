package httpapi

import "github.com/gin-gonic/gin"

func NewRouter(handler *OrderHandler) *gin.Engine {
	router := gin.Default()

	router.POST("/orders", handler.CreateOrder)
	router.GET("/customers/:customerCode/orders", handler.GetOrdersByCustomer)
	router.GET("/customers/:customerCode/orders/count", handler.GetOrderCountByCustomer)
	router.GET("/orders/:orderID/total", handler.GetOrderTotal)

	return router
}
