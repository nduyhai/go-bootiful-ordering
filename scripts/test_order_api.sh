#!/bin/bash

# Test script for order API
# This script tests the basic CRUD operations for orders

# Set the base URL
BASE_URL="http://localhost:8080"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Function to print success message
success() {
  echo -e "${GREEN}SUCCESS: $1${NC}"
}

# Function to print error message
error() {
  echo -e "${RED}ERROR: $1${NC}"
}

# Test creating an order
echo "Testing order creation..."
CREATE_RESPONSE=$(curl -s -X POST "${BASE_URL}/orders" \
  -H "Content-Type: application/json" \
  -d '{
    "customer_id": "customer123",
    "items": [
      {
        "product_id": "product456",
        "quantity": 2,
        "price": 1000
      },
      {
        "product_id": "product789",
        "quantity": 1,
        "price": 1500
      }
    ]
  }')

# Check if order creation was successful
if [[ $CREATE_RESPONSE == *"id"* ]]; then
  success "Order created successfully"
  # Extract order ID from response
  ORDER_ID=$(echo $CREATE_RESPONSE | grep -o '"id":"[^"]*' | cut -d'"' -f4)
  echo "Order ID: $ORDER_ID"
else
  error "Failed to create order: $CREATE_RESPONSE"
  exit 1
fi

# Test getting an order
echo "Testing get order..."
GET_RESPONSE=$(curl -s -X GET "${BASE_URL}/orders/${ORDER_ID}")

# Check if get order was successful
if [[ $GET_RESPONSE == *"$ORDER_ID"* ]]; then
  success "Order retrieved successfully"
else
  error "Failed to retrieve order: $GET_RESPONSE"
  exit 1
fi

# Test listing orders
echo "Testing list orders..."
LIST_RESPONSE=$(curl -s -X GET "${BASE_URL}/orders?customer_id=customer123&page_size=10")

# Check if list orders was successful
if [[ $LIST_RESPONSE == *"orders"* ]]; then
  success "Orders listed successfully"
else
  error "Failed to list orders: $LIST_RESPONSE"
  exit 1
fi

# Test updating order status
echo "Testing update order status..."
UPDATE_RESPONSE=$(curl -s -X PATCH "${BASE_URL}/orders/${ORDER_ID}" \
  -H "Content-Type: application/json" \
  -d '{
    "status": 2
  }')

# Check if update order status was successful
if [[ $UPDATE_RESPONSE == *"status"* && $UPDATE_RESPONSE == *"2"* ]]; then
  success "Order status updated successfully"
else
  error "Failed to update order status: $UPDATE_RESPONSE"
  exit 1
fi

echo "All tests completed successfully!"