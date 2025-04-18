package handler

import (
	"github.com/gin-gonic/gin"
	"go-bootiful-ordering/internal/product/service"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

// CreateProductHandler handles requests to create products
type CreateProductHandler struct {
	log     *zap.Logger
	factory *service.ProductFactory
}

// NewCreateProductHandler creates a new CreateProductHandler
func NewCreateProductHandler(log *zap.Logger, factory *service.ProductFactory) *CreateProductHandler {
	return &CreateProductHandler{
		log:     log,
		factory: factory,
	}
}

// Pattern returns the URL pattern for this handler
func (h *CreateProductHandler) Pattern() string {
	return "/products"
}

// Register registers the handler with the router group
func (h *CreateProductHandler) Register(rg *gin.RouterGroup) {
	rg.POST("/products", h.CreateProduct)
}

// CreateProductRequest represents the request body for creating a product
type CreateProductRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       int64  `json:"price"`
	Stock       int32  `json:"stock"`
	Category    string `json:"category"`
}

// CreateProduct handles HTTP requests to create products
func (h *CreateProductHandler) CreateProduct(c *gin.Context) {
	var req CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Error("Failed to decode request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Validate request
	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name is required"})
		return
	}

	if req.Price <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Price must be greater than 0"})
		return
	}

	if req.Stock < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Stock cannot be negative"})
		return
	}

	// Create product
	product, err := h.factory.CreateProduct(c.Request.Context(), req.Name, req.Description, req.Price, req.Stock, req.Category)
	if err != nil {
		h.log.Error("Failed to create product", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create product"})
		return
	}

	c.JSON(http.StatusCreated, product)
}

// GetProductHandler handles requests to get products
type GetProductHandler struct {
	log     *zap.Logger
	factory *service.ProductFactory
}

// NewGetProductHandler creates a new GetProductHandler
func NewGetProductHandler(log *zap.Logger, factory *service.ProductFactory) *GetProductHandler {
	return &GetProductHandler{
		log:     log,
		factory: factory,
	}
}

// Pattern returns the URL pattern for this handler
func (h *GetProductHandler) Pattern() string {
	return "/products/"
}

// Register registers the handler with the router group
func (h *GetProductHandler) Register(rg *gin.RouterGroup) {
	rg.GET("/products/:id", h.GetProduct)
}

// GetProduct handles HTTP requests to get products
func (h *GetProductHandler) GetProduct(c *gin.Context) {
	productID := c.Param("id")
	if productID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Product ID is required"})
		return
	}

	product, err := h.factory.GetProduct(c.Request.Context(), productID)
	if err != nil {
		h.log.Error("Failed to get product", zap.Error(err), zap.String("productID", productID))
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	c.JSON(http.StatusOK, product)
}

// ListProductsHandler handles requests to list products
type ListProductsHandler struct {
	log     *zap.Logger
	factory *service.ProductFactory
}

// NewListProductsHandler creates a new ListProductsHandler
func NewListProductsHandler(log *zap.Logger, factory *service.ProductFactory) *ListProductsHandler {
	return &ListProductsHandler{
		log:     log,
		factory: factory,
	}
}

// Pattern returns the URL pattern for this handler
func (h *ListProductsHandler) Pattern() string {
	return "/products"
}

// Register registers the handler with the router group
func (h *ListProductsHandler) Register(rg *gin.RouterGroup) {
	rg.GET("/products", h.ListProducts)
}

// ListProducts handles HTTP requests to list products
func (h *ListProductsHandler) ListProducts(c *gin.Context) {
	category := c.Query("category")

	pageSizeStr := c.Query("page_size")
	pageSize := int32(10) // Default page size
	if pageSizeStr != "" {
		if size, err := strconv.ParseInt(pageSizeStr, 10, 32); err == nil {
			pageSize = int32(size)
		}
	}

	pageToken := c.Query("page_token")

	products, nextPageToken, err := h.factory.ListProducts(c.Request.Context(), category, pageSize, pageToken)
	if err != nil {
		h.log.Error("Failed to list products", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list products"})
		return
	}

	response := struct {
		Products      interface{} `json:"products"`
		NextPageToken string      `json:"next_page_token,omitempty"`
	}{
		Products:      products,
		NextPageToken: nextPageToken,
	}

	c.JSON(http.StatusOK, response)
}

// UpdateProductHandler handles requests to update products
type UpdateProductHandler struct {
	log     *zap.Logger
	factory *service.ProductFactory
}

// NewUpdateProductHandler creates a new UpdateProductHandler
func NewUpdateProductHandler(log *zap.Logger, factory *service.ProductFactory) *UpdateProductHandler {
	return &UpdateProductHandler{
		log:     log,
		factory: factory,
	}
}

// Pattern returns the URL pattern for this handler
func (h *UpdateProductHandler) Pattern() string {
	return "/products/"
}

// Register registers the handler with the router group
func (h *UpdateProductHandler) Register(rg *gin.RouterGroup) {
	rg.PUT("/products/:id", h.UpdateProduct)
	rg.PATCH("/products/:id", h.UpdateProduct)
}

// UpdateProductRequest represents the request body for updating a product
type UpdateProductRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       int64  `json:"price"`
	Stock       int32  `json:"stock"`
	Category    string `json:"category"`
}

// UpdateProduct handles HTTP requests to update products
func (h *UpdateProductHandler) UpdateProduct(c *gin.Context) {
	productID := c.Param("id")
	if productID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Product ID is required"})
		return
	}

	var req UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Error("Failed to decode request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Validate request
	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name is required"})
		return
	}

	if req.Price <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Price must be greater than 0"})
		return
	}

	if req.Stock < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Stock cannot be negative"})
		return
	}

	// Update product
	product, err := h.factory.UpdateProduct(c.Request.Context(), productID, req.Name, req.Description, req.Price, req.Stock, req.Category)
	if err != nil {
		h.log.Error("Failed to update product", zap.Error(err), zap.String("productID", productID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product"})
		return
	}

	c.JSON(http.StatusOK, product)
}

// DeleteProductHandler handles requests to delete products
type DeleteProductHandler struct {
	log     *zap.Logger
	factory *service.ProductFactory
}

// NewDeleteProductHandler creates a new DeleteProductHandler
func NewDeleteProductHandler(log *zap.Logger, factory *service.ProductFactory) *DeleteProductHandler {
	return &DeleteProductHandler{
		log:     log,
		factory: factory,
	}
}

// Pattern returns the URL pattern for this handler
func (h *DeleteProductHandler) Pattern() string {
	return "/products/"
}

// Register registers the handler with the router group
func (h *DeleteProductHandler) Register(rg *gin.RouterGroup) {
	rg.DELETE("/products/:id", h.DeleteProduct)
}

// DeleteProduct handles HTTP requests to delete products
func (h *DeleteProductHandler) DeleteProduct(c *gin.Context) {
	productID := c.Param("id")
	if productID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Product ID is required"})
		return
	}

	err := h.factory.DeleteProduct(c.Request.Context(), productID)
	if err != nil {
		h.log.Error("Failed to delete product", zap.Error(err), zap.String("productID", productID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete product"})
		return
	}

	c.Status(http.StatusNoContent)
}
