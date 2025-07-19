package handler

import (
	"github.com/gin-gonic/gin"
	"go-bootiful-ordering/internal/order/domain"
	"go-bootiful-ordering/internal/order/service"
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
	log     *zap.SugaredLogger
	service service.OrderService
}

// NewCreateOrderHandler creates a new CreateOrderHandler
func NewCreateOrderHandler(log *zap.SugaredLogger, service service.OrderService) *CreateOrderHandler {
	return &CreateOrderHandler{
		log:     log,
		service: service,
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
	var request struct {
		CustomerID string             `json:"customer_id"`
		Items      []domain.OrderItem `json:"items"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		h.log.Errorf("Failed to decode request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	order, err := h.service.CreateOrder(c.Request.Context(), request.CustomerID, request.Items)
	if err != nil {
		h.log.Errorf("Failed to create order: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order"})
		return
	}

	c.JSON(http.StatusCreated, order)
}

// GetOrderHandler handles requests to get an order by ID
type GetOrderHandler struct {
	log     *zap.SugaredLogger
	service service.OrderService
}

// NewGetOrderHandler creates a new GetOrderHandler
func NewGetOrderHandler(log *zap.SugaredLogger, service service.OrderService) *GetOrderHandler {
	return &GetOrderHandler{
		log:     log,
		service: service,
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

	order, err := h.service.GetOrder(c.Request.Context(), orderID)
	if err != nil {
		h.log.Errorf("Failed to get order: %v, orderID=%s", err, orderID)
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	c.JSON(http.StatusOK, order)
}

// ListOrdersHandler handles requests to list orders
type ListOrdersHandler struct {
	log     *zap.SugaredLogger
	service service.OrderService
}

// NewListOrdersHandler creates a new ListOrdersHandler
func NewListOrdersHandler(log *zap.SugaredLogger, service service.OrderService) *ListOrdersHandler {
	return &ListOrdersHandler{
		log:     log,
		service: service,
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

	orders, nextPageToken, err := h.service.ListOrders(c.Request.Context(), customerID, pageSize, pageToken)
	if err != nil {
		h.log.Errorf("Failed to list orders: %v, customerID=%s", err, customerID)
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
	log     *zap.SugaredLogger
	service service.OrderService
}

// NewUpdateOrderStatusHandler creates a new UpdateOrderStatusHandler
func NewUpdateOrderStatusHandler(log *zap.SugaredLogger, service service.OrderService) *UpdateOrderStatusHandler {
	return &UpdateOrderStatusHandler{
		log:     log,
		service: service,
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
		h.log.Errorf("Failed to decode request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	order, err := h.service.UpdateOrderStatus(c.Request.Context(), orderID, request.Status)
	if err != nil {
		h.log.Errorf("Failed to update order status: %v, orderID=%s", err, orderID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update order status"})
		return
	}

	c.JSON(http.StatusOK, order)
}
