package repository

import (
	"context"
	"encoding/json"
	"github.com/redis/go-redis/v9"
	"go-bootiful-ordering/internal/product/domain"
	"time"
)

const (
	// Default cache expiration time
	defaultCacheTTL = 30 * time.Minute

	// Key prefixes for Redis
	productKeyPrefix  = "product:"
	categoryKeyPrefix = "category:"
)

// RedisProductRepository implements ProductRepository using Redis for caching
// and delegates to another ProductRepository for persistence
type RedisProductRepository struct {
	redis      *redis.Client
	repository ProductRepository // The underlying repository for persistence
}

// NewRedisProductRepository creates a new RedisProductRepository
func NewRedisProductRepository(redis *redis.Client, repository ProductRepository) *RedisProductRepository {
	return &RedisProductRepository{
		redis:      redis,
		repository: repository,
	}
}

// productKey generates a Redis key for a product
func productKey(productID string) string {
	return productKeyPrefix + productID
}

// categoryKey generates a Redis key for a category
func categoryKey(category string, pageSize int32, pageToken string) string {
	return categoryKeyPrefix + category + ":" + string(pageSize) + ":" + pageToken
}

// CreateProduct persists a new product and invalidates cache
func (r *RedisProductRepository) CreateProduct(ctx context.Context, product *domain.Product) (*domain.Product, error) {
	// Delegate to the underlying repository
	createdProduct, err := r.repository.CreateProduct(ctx, product)
	if err != nil {
		return nil, err
	}

	// Cache the created product
	productJSON, err := json.Marshal(createdProduct)
	if err != nil {
		return createdProduct, nil // Return the product even if caching fails
	}

	// Store in Redis with expiration
	err = r.redis.Set(ctx, productKey(createdProduct.ID), productJSON, defaultCacheTTL).Err()
	if err != nil {
		return createdProduct, nil // Return the product even if caching fails
	}

	return createdProduct, nil
}

// GetProduct retrieves a product by ID, using cache if available
func (r *RedisProductRepository) GetProduct(ctx context.Context, productID string) (*domain.Product, error) {
	// Try to get from cache first
	productJSON, err := r.redis.Get(ctx, productKey(productID)).Bytes()
	if err == nil {
		// Cache hit
		var product domain.Product
		if err := json.Unmarshal(productJSON, &product); err == nil {
			return &product, nil
		}
		// If unmarshaling fails, fall through to get from repository
	}

	// Cache miss or error, get from repository
	product, err := r.repository.GetProduct(ctx, productID)
	if err != nil {
		return nil, err
	}

	// Cache the product for future requests
	productJSON, err = json.Marshal(product)
	if err != nil {
		return product, nil // Return the product even if caching fails
	}

	// Store in Redis with expiration
	err = r.redis.Set(ctx, productKey(product.ID), productJSON, defaultCacheTTL).Err()
	if err != nil {
		return product, nil // Return the product even if caching fails
	}

	return product, nil
}

// ListProducts retrieves a list of products with pagination, using cache if available
func (r *RedisProductRepository) ListProducts(ctx context.Context, category string, pageSize int32, pageToken string) ([]*domain.Product, string, error) {
	// Generate cache key for this query
	cacheKey := categoryKey(category, pageSize, pageToken)

	// Try to get from cache first
	cacheData, err := r.redis.Get(ctx, cacheKey).Bytes()
	if err == nil {
		// Cache hit
		var cacheResult struct {
			Products      []*domain.Product
			NextPageToken string
		}
		if err := json.Unmarshal(cacheData, &cacheResult); err == nil {
			return cacheResult.Products, cacheResult.NextPageToken, nil
		}
		// If unmarshaling fails, fall through to get from repository
	}

	// Cache miss or error, get from repository
	products, nextPageToken, err := r.repository.ListProducts(ctx, category, pageSize, pageToken)
	if err != nil {
		return nil, "", err
	}

	// Cache the results for future requests
	cacheResult := struct {
		Products      []*domain.Product
		NextPageToken string
	}{
		Products:      products,
		NextPageToken: nextPageToken,
	}

	cacheData, err = json.Marshal(cacheResult)
	if err != nil {
		return products, nextPageToken, nil // Return the products even if caching fails
	}

	// Store in Redis with expiration
	err = r.redis.Set(ctx, cacheKey, cacheData, defaultCacheTTL).Err()
	if err != nil {
		return products, nextPageToken, nil // Return the products even if caching fails
	}

	return products, nextPageToken, nil
}

// UpdateProduct updates a product and invalidates cache
func (r *RedisProductRepository) UpdateProduct(ctx context.Context, product *domain.Product) (*domain.Product, error) {
	// Delegate to the underlying repository
	updatedProduct, err := r.repository.UpdateProduct(ctx, product)
	if err != nil {
		return nil, err
	}

	// Invalidate the cache for this product
	err = r.redis.Del(ctx, productKey(updatedProduct.ID)).Err()
	if err != nil {
		// Log the error but continue
		// In a real implementation, you might want to log this error
	}

	// Cache the updated product
	productJSON, err := json.Marshal(updatedProduct)
	if err != nil {
		return updatedProduct, nil // Return the product even if caching fails
	}

	// Store in Redis with expiration
	err = r.redis.Set(ctx, productKey(updatedProduct.ID), productJSON, defaultCacheTTL).Err()
	if err != nil {
		return updatedProduct, nil // Return the product even if caching fails
	}

	return updatedProduct, nil
}

// DeleteProduct deletes a product and invalidates cache
func (r *RedisProductRepository) DeleteProduct(ctx context.Context, productID string) error {
	// Delegate to the underlying repository
	err := r.repository.DeleteProduct(ctx, productID)
	if err != nil {
		return err
	}

	// Invalidate the cache for this product
	err = r.redis.Del(ctx, productKey(productID)).Err()
	if err != nil {
		// Log the error but continue
		// In a real implementation, you might want to log this error
	}

	return nil
}
