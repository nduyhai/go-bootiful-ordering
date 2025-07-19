package handler

import (
	"context"
	"go-bootiful-ordering/gen/order/v1"
	"go-bootiful-ordering/internal/order/domain"
	"go-bootiful-ordering/internal/order/service"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

// GRPCOrderServer implements the OrderService gRPC server
type GRPCOrderServer struct {
	orderv1.UnimplementedOrderServiceServer
	log     *zap.SugaredLogger
	service service.OrderService
}

// NewGRPCOrderServer creates a new GRPCOrderServer
func NewGRPCOrderServer(log *zap.SugaredLogger, service service.OrderService) *GRPCOrderServer {
	return &GRPCOrderServer{
		log:     log,
		service: service,
	}
}

// CreateOrder implements the CreateOrder RPC method
func (s *GRPCOrderServer) CreateOrder(ctx context.Context, req *orderv1.CreateOrderRequest) (*orderv1.CreateOrderResponse, error) {
	s.log.Infof("GRPCOrderServer_CreateOrder customerID=%s", req.CustomerId)

	if req.CustomerId == "" {
		return nil, status.Error(codes.InvalidArgument, "customer_id is required")
	}

	if len(req.Items) == 0 {
		return nil, status.Error(codes.InvalidArgument, "at least one item is required")
	}

	// Convert protobuf items to domain items
	items := make([]domain.OrderItem, len(req.Items))
	for i, item := range req.Items {
		items[i] = domain.OrderItem{
			ProductID: item.ProductId,
			Quantity:  item.Quantity,
			Price:     item.Price,
		}
	}

	// Create order using the service
	order, err := s.service.CreateOrder(ctx, req.CustomerId, items)
	if err != nil {
		s.log.Errorf("Failed to create order: %v", err)
		return nil, status.Error(codes.Internal, "failed to create order")
	}

	// Convert domain order to protobuf order
	return &orderv1.CreateOrderResponse{
		Order: domainToProtoOrder(order),
	}, nil
}

// GetOrder implements the GetOrder RPC method
func (s *GRPCOrderServer) GetOrder(ctx context.Context, req *orderv1.GetOrderRequest) (*orderv1.GetOrderResponse, error) {
	s.log.Infof("GRPCOrderServer_GetOrder orderID=%s", req.OrderId)

	if req.OrderId == "" {
		return nil, status.Error(codes.InvalidArgument, "order_id is required")
	}

	// Get order using the service
	order, err := s.service.GetOrder(ctx, req.OrderId)
	if err != nil {
		s.log.Errorf("Failed to get order: %v, orderID=%s", err, req.OrderId)
		return nil, status.Error(codes.NotFound, "order not found")
	}

	// Convert domain order to protobuf order
	return &orderv1.GetOrderResponse{
		Order: domainToProtoOrder(order),
	}, nil
}

// ListOrders implements the ListOrders RPC method
func (s *GRPCOrderServer) ListOrders(ctx context.Context, req *orderv1.ListOrdersRequest) (*orderv1.ListOrdersResponse, error) {
	s.log.Infof("GRPCOrderServer_ListOrders customerID=%s pageSize=%d pageToken=%s",
		req.CustomerId, req.PageSize, req.PageToken)

	if req.CustomerId == "" {
		return nil, status.Error(codes.InvalidArgument, "customer_id is required")
	}

	// List orders using the service
	orders, nextPageToken, err := s.service.ListOrders(ctx, req.CustomerId, req.PageSize, req.PageToken)
	if err != nil {
		s.log.Errorf("Failed to list orders: %v, customerID=%s", err, req.CustomerId)
		return nil, status.Error(codes.Internal, "failed to list orders")
	}

	// Convert domain orders to protobuf orders
	protoOrders := make([]*orderv1.Order, len(orders))
	for i, order := range orders {
		protoOrders[i] = domainToProtoOrder(order)
	}

	return &orderv1.ListOrdersResponse{
		Orders:        protoOrders,
		NextPageToken: nextPageToken,
	}, nil
}

// UpdateOrderStatus implements the UpdateOrderStatus RPC method
func (s *GRPCOrderServer) UpdateOrderStatus(ctx context.Context, req *orderv1.UpdateOrderStatusRequest) (*orderv1.UpdateOrderStatusResponse, error) {
	s.log.Infof("GRPCOrderServer_UpdateOrderStatus orderID=%s status=%d",
		req.OrderId, int32(req.Status))

	if req.OrderId == "" {
		return nil, status.Error(codes.InvalidArgument, "order_id is required")
	}

	// Convert protobuf status to domain status
	var orderStatus domain.OrderStatus
	switch req.Status {
	case orderv1.OrderStatus_ORDER_STATUS_PENDING:
		orderStatus = domain.OrderStatusPending
	case orderv1.OrderStatus_ORDER_STATUS_PROCESSING:
		orderStatus = domain.OrderStatusProcessing
	case orderv1.OrderStatus_ORDER_STATUS_SHIPPED:
		orderStatus = domain.OrderStatusShipped
	case orderv1.OrderStatus_ORDER_STATUS_DELIVERED:
		orderStatus = domain.OrderStatusDelivered
	case orderv1.OrderStatus_ORDER_STATUS_CANCELLED:
		orderStatus = domain.OrderStatusCancelled
	default:
		return nil, status.Error(codes.InvalidArgument, "invalid order status")
	}

	// Update order status using the service
	order, err := s.service.UpdateOrderStatus(ctx, req.OrderId, orderStatus)
	if err != nil {
		s.log.Errorf("Failed to update order status: %v, orderID=%s", err, req.OrderId)
		return nil, status.Error(codes.Internal, "failed to update order status")
	}

	// Convert domain order to protobuf order
	return &orderv1.UpdateOrderStatusResponse{
		Order: domainToProtoOrder(order),
	}, nil
}

// domainToProtoOrder converts a domain order to a protobuf order
func domainToProtoOrder(order *domain.Order) *orderv1.Order {
	// Convert domain items to protobuf items
	items := make([]*orderv1.OrderItem, len(order.Items))
	for i, item := range order.Items {
		items[i] = &orderv1.OrderItem{
			ProductId: item.ProductID,
			Quantity:  item.Quantity,
			Price:     item.Price,
		}
	}

	// Convert domain status to protobuf status
	var protoStatus orderv1.OrderStatus
	switch order.Status {
	case domain.OrderStatusPending:
		protoStatus = orderv1.OrderStatus_ORDER_STATUS_PENDING
	case domain.OrderStatusProcessing:
		protoStatus = orderv1.OrderStatus_ORDER_STATUS_PROCESSING
	case domain.OrderStatusShipped:
		protoStatus = orderv1.OrderStatus_ORDER_STATUS_SHIPPED
	case domain.OrderStatusDelivered:
		protoStatus = orderv1.OrderStatus_ORDER_STATUS_DELIVERED
	case domain.OrderStatusCancelled:
		protoStatus = orderv1.OrderStatus_ORDER_STATUS_CANCELLED
	default:
		protoStatus = orderv1.OrderStatus_ORDER_STATUS_UNSPECIFIED
	}

	return &orderv1.Order{
		Id:          order.ID,
		CustomerId:  order.CustomerID,
		Items:       items,
		Status:      protoStatus,
		TotalAmount: order.TotalAmount,
		CreatedAt:   order.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   order.UpdatedAt.Format(time.RFC3339),
	}
}
