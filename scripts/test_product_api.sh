#!/bin/bash

# Test script for Product API
# This script tests the basic CRUD operations of the Product API

# Set the base URL
BASE_URL="http://localhost:8081"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Function to print success message
success() {
  echo -e "${GREEN}✓ $1${NC}"
}

# Function to print error message
error() {
  echo -e "${RED}✗ $1${NC}"
  exit 1
}

echo "Testing Product API..."

# Create a product
echo "Creating a product..."
CREATE_RESPONSE=$(curl -s -X POST -H "Content-Type: application/json" -d '{
  "name": "Test Product",
  "description": "This is a test product",
  "price": 1999,
  "stock": 100,
  "category": "test"
}' $BASE_URL/products)

# Extract product ID from response
PRODUCT_ID=$(echo $CREATE_RESPONSE | grep -o '"id":"[^"]*' | cut -d'"' -f4)

if [ -z "$PRODUCT_ID" ]; then
  error "Failed to create product"
else
  success "Product created with ID: $PRODUCT_ID"
fi

# Get the product
echo "Getting the product..."
GET_RESPONSE=$(curl -s -X GET $BASE_URL/products/$PRODUCT_ID)

if [[ $GET_RESPONSE == *"Test Product"* ]]; then
  success "Product retrieved successfully"
else
  error "Failed to get product"
fi

# Update the product
echo "Updating the product..."
UPDATE_RESPONSE=$(curl -s -X PUT -H "Content-Type: application/json" -d '{
  "name": "Updated Test Product",
  "description": "This is an updated test product",
  "price": 2999,
  "stock": 50,
  "category": "test-updated"
}' $BASE_URL/products/$PRODUCT_ID)

if [[ $UPDATE_RESPONSE == *"Updated Test Product"* ]]; then
  success "Product updated successfully"
else
  error "Failed to update product"
fi

# List products
echo "Listing products..."
LIST_RESPONSE=$(curl -s -X GET "$BASE_URL/products?category=test-updated")

if [[ $LIST_RESPONSE == *"Updated Test Product"* ]]; then
  success "Products listed successfully"
else
  error "Failed to list products"
fi

# Delete the product
echo "Deleting the product..."
DELETE_RESPONSE=$(curl -s -X DELETE -w "%{http_code}" $BASE_URL/products/$PRODUCT_ID -o /dev/null)

if [ "$DELETE_RESPONSE" -eq 204 ]; then
  success "Product deleted successfully"
else
  error "Failed to delete product"
fi

# Try to get the deleted product (should fail)
echo "Verifying deletion..."
GET_DELETED_RESPONSE=$(curl -s -X GET -w "%{http_code}" $BASE_URL/products/$PRODUCT_ID -o /dev/null)

if [ "$GET_DELETED_RESPONSE" -eq 404 ]; then
  success "Product deletion verified"
else
  error "Product still exists after deletion"
fi

echo "All tests passed!"