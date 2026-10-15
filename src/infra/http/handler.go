package httpapi

import (
	"net/http"

	portsin "github.com/DevThiagoSouza/src/application/ports/in"
	"github.com/DevThiagoSouza/src/domain"
	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	orderQuery  portsin.OrderQuery
	orderCreator portsin.OrderCreator
}

func NewOrderHandler(orderQuery portsin.OrderQuery, orderCreator portsin.OrderCreator) *OrderHandler {
	return &OrderHandler{
		orderQuery:  orderQuery,
		orderCreator: orderCreator,
	}
}

func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var order domain.Order
	if err := c.ShouldBindJSON(&order); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.orderCreator.CreateOrder(c.Request.Context(), order); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "order created"})
}

func (h *OrderHandler) GetOrdersByCustomer(c *gin.Context) {
	customerCode := c.Param("customerCode")
	orders, err := h.orderQuery.GetOrdersByCustomer(c.Request.Context(), customerCode)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"orders": orders})
}

func (h *OrderHandler) GetOrderTotal(c *gin.Context) {
	orderID := c.Param("orderID")
	total, err := h.orderQuery.GetOrderTotal(c.Request.Context(), orderID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"order_id": orderID, "total": total})
}

func (h *OrderHandler) GetOrderCountByCustomer(c *gin.Context) {
	customerCode := c.Param("customerCode")
	count, err := h.orderQuery.GetOrderCountByCustomer(c.Request.Context(), customerCode)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"customer_code": customerCode, "count": count})
}
