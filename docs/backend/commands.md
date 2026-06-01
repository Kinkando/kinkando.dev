# Backend Commands

```bash
# Run the server
go run ./cmd/main.go

# Build
go build -o server ./cmd/main.go

# Build Docker image
docker build -t personal-dashboard .

# Run tests
go test ./...

# Run a single package's tests
go test ./internal/finance/...

# Lint (requires golangci-lint)
golangci-lint run
```
