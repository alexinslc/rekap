.PHONY: build run clean install test release

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

# Run tests
test:
	go test -v ./...

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
