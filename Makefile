# Makefile for jfrog-client-go

.PHONY: $(MAKECMDGOALS)

# Default target
help:
	@echo "Available targets:"
	@echo "  update-all               - Update all JFrog dependencies to latest versions"
	@echo "  update-build-info-go     - Update build-info-go to latest main branch"
	@echo "  update-client-go         - Update client-go to latest main branch"
	@echo "  update-gofrog            - Update gofrog to latest main branch"
	@echo "  update-core              - Update jfrog-cli-core to latest main branch"
	@echo "  update-artifactory       - Update jfrog-cli-artifactory to latest main branch"
	@echo "  update-platform-services - Update jfrog-cli-platform-services to latest main branch"
	@echo "  update-security          - Update jfrog-cli-security to latest main branch"
	@echo "  update-apptrust          - Update jfrog-cli-application to latest main branch"
	@echo "  clean                    - Clean build artifacts"
	@echo "  test                     - Run tests"
	@echo "  build                    - Build the project"

# Update all JFrog dependencies
update-all: update-build-info-go update-client-go update-gofrog update-core update-artifactory update-platform-services update-security update-apptrust update-evidence
	@echo "All JFrog dependencies updated successfully!"
	@GOPROXY=direct go mod tidy

# Update build-info-go to latest main branch (using direct proxy to bypass Artifactory)
update-build-info-go:
	@echo "Updating build-info-go to latest main branch..."
	@GOPROXY=direct go get github.com/jfrog/build-info-go@main
	@echo "build-info-go updated successfully!"

# Update gofrog to latest main branch
update-client-go:
	@echo "Updating client-go to latest main branch..."
	@GOPROXY=direct go get github.com/jfrog/jfrog-client-go@master
	@echo "client-go updated successfully!"

# Update gofrog to latest main branch
update-gofrog:
	@echo "Updating gofrog to latest main branch..."
	@GOPROXY=direct go get github.com/jfrog/gofrog@master
	@echo "gofrog updated successfully!"

# Update jfrog-cli-core to latest main branch
update-core:
	@echo "Updating jfrog-cli-core to latest main branch..."
	@GOPROXY=direct go get github.com/jfrog/jfrog-cli-core/v2@master
	@echo "jfrog-cli-core updated successfully!"

# Update jfrog-cli-artifactory to latest main branch
update-artifactory:
	@echo "Updating jfrog-cli-artifactory to latest main branch..."
	@GOPROXY=direct go get github.com/jfrog/jfrog-cli-artifactory@main
	@echo "jfrog-cli-artifactory updated successfully!"

# Update jfrog-cli-platform-services to latest main branch
update-platform-services:
	@echo "Updating jfrog-cli-platform-services to latest main branch..."
	@GOPROXY=direct go get github.com/jfrog/jfrog-cli-platform-services@main
	@echo "jfrog-cli-platform-services updated successfully!"

# Update jfrog-cli-security to latest main branch
update-security:
	@echo "Updating jfrog-cli-security to latest main branch..."
	@GOPROXY=direct go get github.com/jfrog/jfrog-cli-security@main
	@echo "jfrog-cli-security updated successfully!"

# Update jfrog-cli-application to latest main branch
update-apptrust:
	@echo "Updating jfrog-cli-application to latest main branch..."
	@GOPROXY=direct go get github.com/jfrog/jfrog-cli-application@main
	@echo "jfrog-cli-application updated successfully!"

# Update jfrog-cli-evidence to latest main branch
update-evidence:
	@echo "Updating jfrog-cli-evidence to latest main branch..."
	@GOPROXY=direct go get github.com/jfrog/jfrog-cli-evidence@main
	@echo "jfrog-cli-evidence updated successfully!"


# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@go clean
	@go clean -cache
	@go clean -modcache

# Run tests
test:
	@echo "Running tests..."
	@go test ./...

# Build the project
build:
	@echo "Building project..."
	@go build ./...
