.PHONY: build run clean install install-man test test-fast test-coverage test-bench lint fmt vet release

# Build the binary
build:
	go build -ldflags="-s -w" -o rekap ./cmd/rekap

# Run the application
run: build
	./rekap

# Clean build artifacts
clean:
	rm -f rekap
	rm -rf dist/

# Install to /usr/local/bin
install: build
	cp rekap /usr/local/bin/
	@echo "Binary installed to /usr/local/bin/rekap"
	@echo "To install man page, run: sudo make install-man"

# Install man page to /usr/local/share/man/man1
install-man:
	@mkdir -p /usr/local/share/man/man1
	cp rekap.1 /usr/local/share/man/man1/
	@echo "Man page installed to /usr/local/share/man/man1/rekap.1"
	@echo "You can now run: man rekap"

# Run all tests with verbose output
test:
	go test -v ./...

# Run tests quickly (cached)
test-fast:
	go test ./...

# Run tests with coverage report
test-coverage:
	go test -cover -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Run benchmarks
test-bench:
	go test -bench=. -benchmem ./...

# Run linter (requires golangci-lint)
lint:
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not found. Install it from https://golangci-lint.run/usage/install/"; \
		echo "Running basic checks instead..."; \
		go vet ./... && go fmt ./...; \
	fi

# Format code
fmt:
	go fmt ./...

# Run go vet
vet:
	go vet ./...

# Build for multiple architectures
release:
	@echo "Building for macOS arm64 and amd64..."
	@mkdir -p dist
	GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o dist/rekap-darwin-arm64 ./cmd/rekap
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o dist/rekap-darwin-amd64 ./cmd/rekap
	@echo "Binaries created in dist/"
	@ls -lh dist/

# Create universal binary
universal:
	@echo "Creating universal binary..."
	@mkdir -p dist
	$(MAKE) release
	lipo -create -output dist/rekap-universal dist/rekap-darwin-arm64 dist/rekap-darwin-amd64
	@echo "Universal binary created"
	@ls -lh dist/rekap-universal
