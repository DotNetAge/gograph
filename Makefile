.PHONY: build
build:
	go build -ldflags "-X main.Version=$(shell git describe --tags --always)" -o bin/gograph ./cmd/gograph

.PHONY: build-all
build-all:
	@VERSION=$$(git describe --tags --always); \
	echo "Building for macOS (Intel)..."; \
	GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.Version=$$VERSION" -o bin/darwin-amd64/gograph ./cmd/gograph; \
	echo "Building for macOS (ARM)..."; \
	GOOS=darwin GOARCH=arm64 go build -ldflags "-X main.Version=$$VERSION" -o bin/darwin-arm64/gograph ./cmd/gograph; \
	echo "Building for Linux (Intel)..."; \
	GOOS=linux GOARCH=amd64 go build -ldflags "-X main.Version=$$VERSION" -o bin/linux-amd64/gograph ./cmd/gograph; \
	echo "Building for Linux (ARM)..."; \
	GOOS=linux GOARCH=arm64 go build -ldflags "-X main.Version=$$VERSION" -o bin/linux-arm64/gograph ./cmd/gograph

test: 
	go test ./...

.PHONY: coverage
coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

.PHONY: coverage-html
coverage-html:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

.PHONY: release
release:
	@VERSION=$$(git describe --tags --always); \
	if [ -z "$$VERSION" ]; then echo "No git tags found"; exit 1; fi; \
	echo "Releasing version $$VERSION"; \
	rm -rf dist; \
	mkdir -p dist/darwin-amd64 dist/darwin-arm64 dist/linux-amd64 dist/linux-arm64; \
	echo "Packaging darwin/amd64..."; \
	GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.Version=$$VERSION" -o dist/darwin-amd64/gograph ./cmd/gograph; \
	tar -czf dist/gograph-$$VERSION-darwin-amd64.tar.gz -C dist/darwin-amd64 gograph; \
	MAC_AMD64_SHA=$$(shasum -a 256 dist/gograph-$$VERSION-darwin-amd64.tar.gz | awk '{print $$1}'); \
	echo "Packaging darwin/arm64..."; \
	GOOS=darwin GOARCH=arm64 go build -ldflags "-X main.Version=$$VERSION" -o dist/darwin-arm64/gograph ./cmd/gograph; \
	tar -czf dist/gograph-$$VERSION-darwin-arm64.tar.gz -C dist/darwin-arm64 gograph; \
	MAC_ARM64_SHA=$$(shasum -a 256 dist/gograph-$$VERSION-darwin-arm64.tar.gz | awk '{print $$1}'); \
	echo "Packaging linux/amd64..."; \
	GOOS=linux GOARCH=amd64 go build -ldflags "-X main.Version=$$VERSION" -o dist/linux-amd64/gograph ./cmd/gograph; \
	tar -czf dist/gograph-$$VERSION-linux-amd64.tar.gz -C dist/linux-amd64 gograph; \
	LINUX_AMD64_SHA=$$(shasum -a 256 dist/gograph-$$VERSION-linux-amd64.tar.gz | awk '{print $$1}'); \
	echo "Packaging linux/arm64..."; \
	GOOS=linux GOARCH=arm64 go build -ldflags "-X main.Version=$$VERSION" -o dist/linux-arm64/gograph ./cmd/gograph; \
	tar -czf dist/gograph-$$VERSION-linux-arm64.tar.gz -C dist/linux-arm64 gograph; \
	LINUX_ARM64_SHA=$$(shasum -a 256 dist/gograph-$$VERSION-linux-arm64.tar.gz | awk '{print $$1}'); \
	sed -e "s/{{VERSION}}/$$VERSION/g" \
	    -e "s/{{MAC_AMD64_SHA256}}/$$MAC_AMD64_SHA/g" \
	    -e "s/{{MAC_ARM64_SHA256}}/$$MAC_ARM64_SHA/g" \
	    -e "s/{{LINUX_AMD64_SHA256}}/$$LINUX_AMD64_SHA/g" \
	    -e "s/{{LINUX_ARM64_SHA256}}/$$LINUX_ARM64_SHA/g" \
	    scripts/formula.rb.tmpl > scripts/gograph.rb; \
	echo "Homebrew formula generated in scripts/gograph.rb"
