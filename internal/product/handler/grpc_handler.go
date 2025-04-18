package handler

import (
	"context"
	"go-bootiful-ordering/gen/product/v1"
	"go-bootiful-ordering/internal/product/domain"
	"go-bootiful-ordering/internal/product/service"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

// GRPCProductServer implements the ProductService gRPC server
type GRPCProductServer struct {
	productv1.UnimplementedProductServiceServer
	log     *zap.Logger
	factory *service.ProductFactory
}

// NewGRPCProductServer creates a new GRPCProductServer
func NewGRPCProductServer(log *zap.Logger, factory *service.ProductFactory) *GRPCProductServer {
	return &GRPCProductServer{
		log:     log,
		factory: factory,
	}
}

// CreateProduct implements the CreateProduct RPC method
func (s *GRPCProductServer) CreateProduct(ctx context.Context, req *productv1.CreateProductRequest) (*productv1.CreateProductResponse, error) {
	s.log.Info("GRPCProductServer_CreateProduct", 
		zap.String("name", req.Name),
		zap.String("category", req.Category))

	// Validate request
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}

	if req.Price <= 0 {
		return nil, status.Error(codes.InvalidArgument, "price must be greater than 0")
	}

	if req.Stock < 0 {
		return nil, status.Error(codes.InvalidArgument, "stock cannot be negative")
	}

	// Create product using the factory
	product, err := s.factory.CreateProduct(ctx, req.Name, req.Description, req.Price, req.Stock, req.Category)
	if err != nil {
		s.log.Error("Failed to create product", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to create product")
	}

	// Convert domain product to protobuf product
	return &productv1.CreateProductResponse{
		Product: domainToProtoProduct(product),
	}, nil
}

// GetProduct implements the GetProduct RPC method
func (s *GRPCProductServer) GetProduct(ctx context.Context, req *productv1.GetProductRequest) (*productv1.GetProductResponse, error) {
	s.log.Info("GRPCProductServer_GetProduct", zap.String("productID", req.ProductId))

	if req.ProductId == "" {
		return nil, status.Error(codes.InvalidArgument, "product_id is required")
	}

	// Get product using the factory
	product, err := s.factory.GetProduct(ctx, req.ProductId)
	if err != nil {
		s.log.Error("Failed to get product", zap.Error(err), zap.String("productID", req.ProductId))
		return nil, status.Error(codes.NotFound, "product not found")
	}

	// Convert domain product to protobuf product
	return &productv1.GetProductResponse{
		Product: domainToProtoProduct(product),
	}, nil
}

// ListProducts implements the ListProducts RPC method
func (s *GRPCProductServer) ListProducts(ctx context.Context, req *productv1.ListProductsRequest) (*productv1.ListProductsResponse, error) {
	s.log.Info("GRPCProductServer_ListProducts", 
		zap.String("category", req.Category),
		zap.Int32("pageSize", req.PageSize),
		zap.String("pageToken", req.PageToken))

	// List products using the factory
	products, nextPageToken, err := s.factory.ListProducts(ctx, req.Category, req.PageSize, req.PageToken)
	if err != nil {
		s.log.Error("Failed to list products", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to list products")
	}

	// Convert domain products to protobuf products
	protoProducts := make([]*productv1.Product, len(products))
	for i, product := range products {
		protoProducts[i] = domainToProtoProduct(product)
	}

	return &productv1.ListProductsResponse{
		Products:      protoProducts,
		NextPageToken: nextPageToken,
	}, nil
}

// UpdateProduct implements the UpdateProduct RPC method
func (s *GRPCProductServer) UpdateProduct(ctx context.Context, req *productv1.UpdateProductRequest) (*productv1.UpdateProductResponse, error) {
	s.log.Info("GRPCProductServer_UpdateProduct", 
		zap.String("productID", req.ProductId),
		zap.String("name", req.Name),
		zap.String("category", req.Category))

	if req.ProductId == "" {
		return nil, status.Error(codes.InvalidArgument, "product_id is required")
	}

	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}

	if req.Price <= 0 {
		return nil, status.Error(codes.InvalidArgument, "price must be greater than 0")
	}

	if req.Stock < 0 {
		return nil, status.Error(codes.InvalidArgument, "stock cannot be negative")
	}

	// Update product using the factory
	product, err := s.factory.UpdateProduct(ctx, req.ProductId, req.Name, req.Description, req.Price, req.Stock, req.Category)
	if err != nil {
		s.log.Error("Failed to update product", zap.Error(err), zap.String("productID", req.ProductId))
		return nil, status.Error(codes.Internal, "failed to update product")
	}

	// Convert domain product to protobuf product
	return &productv1.UpdateProductResponse{
		Product: domainToProtoProduct(product),
	}, nil
}

// DeleteProduct implements the DeleteProduct RPC method
func (s *GRPCProductServer) DeleteProduct(ctx context.Context, req *productv1.DeleteProductRequest) (*productv1.DeleteProductResponse, error) {
	s.log.Info("GRPCProductServer_DeleteProduct", zap.String("productID", req.ProductId))

	if req.ProductId == "" {
		return nil, status.Error(codes.InvalidArgument, "product_id is required")
	}

	// Delete product using the factory
	err := s.factory.DeleteProduct(ctx, req.ProductId)
	if err != nil {
		s.log.Error("Failed to delete product", zap.Error(err), zap.String("productID", req.ProductId))
		return nil, status.Error(codes.Internal, "failed to delete product")
	}

	return &productv1.DeleteProductResponse{
		Success: true,
	}, nil
}

// domainToProtoProduct converts a domain product to a protobuf product
func domainToProtoProduct(product *domain.Product) *productv1.Product {
	return &productv1.Product{
		Id:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		Stock:       product.Stock,
		Category:    product.Category,
		CreatedAt:   product.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   product.UpdatedAt.Format(time.RFC3339),
	}
}