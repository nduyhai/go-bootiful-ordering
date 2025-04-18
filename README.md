# Go Bootiful Ordering

A clean architecture implementation of an ordering system in Go, with persistence using GORM and PostgreSQL.

## Project Structure

The project follows clean architecture principles:

- `cmd/`: Contains the application entry points
- `internal/`: Contains the application code
  - `order/`: Order service implementation
    - `domain/`: Domain models and business logic
    - `repository/`: Data access layer
    - `service/`: Business logic layer
    - `handler/`: API handlers
    - `config/`: Configuration
  - `product/`: Product service implementation

## Prerequisites

- Go 1.24 or later
- PostgreSQL 12 or later

## Setup

1. Clone the repository
2. Install dependencies:
   ```
   go mod download
   ```
3. Set up PostgreSQL:
   ```
   # Create a database
   createdb orders

   # Or use Docker
   docker run --name postgres -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=orders -p 5432:5432 -d postgres

   # Or use Docker Compose (recommended)
   docker-compose up -d
   ```

## Database Migrations

The project uses [Atlas](https://atlasgo.io/) for database migrations. Atlas is a modern database schema migration tool that supports versioned migrations, validation, and more.

### Running Migrations

To run migrations, use the following commands:

```bash
# Install Atlas CLI (if not already installed)
curl -sSf https://atlasgo.sh | sh

# Run migrations for the order service
make migrate-order

# Run migrations for the product service
make migrate-product
```

### Migration Files

Migration files are stored in the `migrations` directory, with subdirectories for each service:

- `migrations/order/`: Contains migration files for the order service
- `migrations/product/`: Contains migration files for the product service

Each migration file is written in HCL (HashiCorp Configuration Language) and defines the database schema for the service.

## Configuration

The application uses environment variables for configuration:

- `DB_HOST`: PostgreSQL host (default: localhost)
- `DB_PORT`: PostgreSQL port (default: 5432)
- `DB_USER`: PostgreSQL user (default: postgres)
- `DB_PASSWORD`: PostgreSQL password (default: postgres)
- `DB_NAME`: PostgreSQL database name (default: orders)
- `DB_SSL_MODE`: PostgreSQL SSL mode (default: disable)

For the migration tool, you can use service-specific environment variables:

- `ORDER_DB_HOST`, `ORDER_DB_PORT`, etc.: For the order service database
- `PRODUCT_DB_HOST`, `PRODUCT_DB_PORT`, etc.: For the product service database

## Docker Compose

The project includes a Docker Compose configuration for setting up the required PostgreSQL database. This is the recommended way to set up the development environment.

### Using Docker Compose

1. Start the PostgreSQL container:
   ```
   docker-compose up -d
   ```

2. The Docker Compose setup creates:
   - A PostgreSQL container with the following configuration:
     - User: myuser
     - Password: secret
     - Default database: order
     - Additional database: products
   - Port 5432 is exposed to the host machine
   - Data is persisted in a Docker volume

3. To stop the container:
   ```
   docker-compose down
   ```

4. To stop the container and remove the volume (this will delete all data):
   ```
   docker-compose down -v
   ```

## Running the Application

```
go run cmd/order/main.go
```

The application will start an HTTP server on port 8080.

## Testing

A test script is provided to test the API endpoints:

```
chmod +x scripts/test_order_api.sh
./scripts/test_order_api.sh
```

## API Endpoints

- `POST /orders`: Create a new order
- `GET /orders/{id}`: Get an order by ID
- `GET /orders?customer_id={id}&page_size={size}&page_token={token}`: List orders for a customer
- `PATCH /orders/{id}`: Update an order's status

## Implementation Details

### Clean Architecture

The project follows clean architecture principles:

- Domain layer: Contains the business models and logic
- Repository layer: Handles data persistence
- Service layer: Implements business logic
- Handler layer: Handles HTTP requests and responses

### GORM and PostgreSQL

The application uses GORM as an ORM to interact with PostgreSQL. The repository layer implements the OrderRepository interface using GORM.

### Dependency Injection

The application uses Uber FX for dependency injection, making it easy to swap out implementations of interfaces.
