package handler

import (
	"github.com/gin-gonic/gin"
	"go-bootiful-ord
	"go-bootiful-ordering/internal/order/domain"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

// Route interface defines a HTTP route handler
type Route interface {
	Register(*gin.RouterGroup)
	Pattern() string
}

// CreateOrderHandler handles order creation requests
type CreateOrderHandler struct {
	log     *zap.Logger
	factory *service.OrderFactory
}

// NewCreateOrderHandler creates a new CreateOrderHandler
func NewCreateOrderHandler(log *zap.Logger, factory *service.OrderFactory) *CreateOrderHandler {
	return &CreateOrderHandler{
		log:     log,
		factory: factory,
	}
}

// Pattern returns the URL pattern for this handler
func (h *CreateOrderHandler) Pattern() string {
	return "/orders"
}

// Register registers the handler with the router group
func (h *CreateOrderHandler) Register(rg *gin.RouterGroup) {
	rg.POST("/orders", h.CreateOrder)
}

// CreateOrder handles HTTP requests to create orders
func (h *CreateOrderHandler) CreateOrder(c *gin.Context) {
		CustomerID string             `json:"customer_id"`
		CustomerID string            `json:"customer_id"`
		Items      []domain.OrderItem `json:"items"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		h.log.Error("Failed to decode request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	order, err := h.factory.CreateOrder(c.Request.Context(), request.CustomerID, request.Items)
	if err != nil {
		h.log.Error("Failed to create order", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order"})
		return
	}

	c.JSON(http.StatusCreated, order)
}

// GetOrderHandler handles requests to get an order by ID
type GetOrderHandler struct {
	log     *zap.Logger
	factory *service.OrderFactory
}

// NewGetOrderHandler creates a new GetOrderHandler
func NewGetOrderHandler(log *zap.Logger, factory *service.OrderFactory) *GetOrderHandler {
	return &GetOrderHandler{
		log:     log,
		factory: factory,
	}
}

// Pattern returns the URL pattern for this handler
func (h *GetOrderHandler) Pattern() string {
	return "/orders/"
}

// Register registers the handler with the router group
func (h *GetOrderHandler) Register(rg *gin.RouterGroup) {
	rg.GET("/orders/:id", h.GetOrder)
}

// GetOrder handles HTTP requests to get orders
func (h *GetOrderHandler) GetOrder(c *gin.Context) {
	orderID := c.Param("id")
	if orderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order ID is required"})
		return
	}

	order, err := h.factory.GetOrder(c.Request.Context(), orderID)
	if err != nil {
		h.log.Error("Failed to get order", zap.Error(err), zap.String("orderID", orderID))
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	c.JSON(http.StatusOK, order)
}

// ListOrdersHandler handles requests to list orders
type ListOrdersHandler struct {
	log     *zap.Logger
	factory *service.OrderFactory
}

// NewListOrdersHandler creates a new ListOrdersHandler
func NewListOrdersHandler(log *zap.Logger, factory *service.OrderFactory) *ListOrdersHandler {
	return &ListOrdersHandler{
		log:     log,
		factory: factory,
	}
}

// Pattern returns the URL pattern for this handler
func (h *ListOrdersHandler) Pattern() string {
	return "/orders"
}

// Register registers the handler with the router group
func (h *ListOrdersHandler) Register(rg *gin.RouterGroup) {
	rg.GET("/orders", h.ListOrders)
}

// ListOrders handles HTTP requests to list orders
func (h *ListOrdersHandler) ListOrders(c *gin.Context) {
	customerID := c.Query("customer_id")
	if customerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Customer ID is required"})
		return
	}

	pageSizeStr := c.Query("page_size")
	pageSize := int32(10) // Default page size
	if pageSizeStr != "" {
		if size, err := strconv.ParseInt(pageSizeStr, 10, 32); err == nil {
			pageSize = int32(size)
		}
	}

	pageToken := c.Query("page_token")

	orders, nextPageToken, err := h.factory.ListOrders(c.Request.Context(), customerID, pageSize, pageToken)
	if err != nil {
		h.log.Error("Failed to list orders", zap.Error(err), zap.String("customerID", customerID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list orders"})
		return
	}

	response := struct {
		Orders        []*domain.Order `json:"orders"`
		NextPageToken string          `json:"next_page_token,omitempty"`
	}{
		Orders:        orders,
		NextPageToken: nextPageToken,
	}

	c.JSON(http.StatusOK, response)
}

// UpdateOrderStatusHandler handles requests to update an order's status
type UpdateOrderStatusHandler struct {
	log     *zap.Logger
	factory *service.OrderFactory
}

// NewUpdateOrderStatusHandler creates a new UpdateOrderStatusHandler
func NewUpdateOrderStatusHandler(log *zap.Logger, factory *service.OrderFactory) *UpdateOrderStatusHandler {
	return &UpdateOrderStatusHandler{
		log:     log,
		factory: factory,
	}
}

// Pattern returns the URL pattern for this handler
func (h *UpdateOrderStatusHandler) Pattern() string {
	return "/orders/"
}

// Register registers the handler with the router group
func (h *UpdateOrderStatusHandler) Register(rg *gin.RouterGroup) {
	rg.PATCH("/orders/:id", h.UpdateOrderStatus)
}

// UpdateOrderStatus handles HTTP requests to update order status
func (h *UpdateOrderStatusHandler) UpdateOrderStatus(c *gin.Context) {
	orderID := c.Param("id")
	if orderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order ID is required"})
		return
	}

	var request struct {
		Status domain.OrderStatus `json:"status"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		h.log.Error("Failed to decode request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	order, err := h.factory.UpdateOrderStatus(c.Request.Context(), orderID, request.Status)
	if err != nil {
		h.log.Error("Failed to update order status", zap.Error(err), zap.String("orderID", orderID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update order status"})
		return
	}

	c.JSON(http.StatusOK, order)
}
