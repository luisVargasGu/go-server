.PHONY: dev prod clean

# Development environment
dev:
	@echo "Running in development mode..."
	@export GO_ENV=development
	@go run main.go

# Production environment
prod:
	@echo "Running in production mode..."
	@export GO_ENV=production
	@go run main.go

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -f go-server

# Build the application
build:
	@echo "Building the application..."
	@go build -o go-server main.go

# Run tests
test:
	@echo "Running tests..."
	@go test ./... 