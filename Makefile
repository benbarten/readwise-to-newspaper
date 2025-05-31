.PHONY: help build run clean test yesterday week month env-example env-check

# Load .env file if it exists
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

# Default target
help:
	@echo "Readwise to Newspaper - Available commands:"
	@echo ""
	@echo "  make build         - Build the binary"
	@echo "  make run           - Run the application (requires READWISE_TOKEN and CUTOFF_DATETIME)"
	@echo "  make yesterday     - Run with articles from yesterday (requires READWISE_TOKEN)"
	@echo "  make week          - Run with articles from last 7 days (requires READWISE_TOKEN)"
	@echo "  make month         - Run with articles from last 30 days (requires READWISE_TOKEN)"
	@echo "  make clean         - Clean up generated files"
	@echo "  make test          - Run tests"
	@echo "  make env-example   - Create a sample .env file"
	@echo "  make env-check     - Show current environment configuration"
	@echo ""
	@echo "Environment variables:"
	@echo "  READWISE_TOKEN     - Your Readwise API token (get from https://readwise.io/access_token)"
	@echo "  CUTOFF_DATETIME    - Unix timestamp for earliest article date"
	@echo ""
	@echo "Examples:"
	@echo "  READWISE_TOKEN=your_token make yesterday"
	@echo "  make env-example && vi .env && make week"
	@echo ""
	@if [ -f .env ]; then \
		echo "✅ .env file found - environment variables will be loaded automatically"; \
	else \
		echo "💡 No .env file found - create one with 'make env-example'"; \
	fi

# Check current environment configuration
env-check:
	@echo "Current environment configuration:"
	@echo ""
	@if [ -f .env ]; then \
		echo "📁 .env file: Found"; \
		echo "Contents:"; \
		sed 's/^/  /' .env; \
	else \
		echo "📁 .env file: Not found"; \
	fi
	@echo ""
	@if [ -n "$(READWISE_TOKEN)" ]; then \
		echo "🔑 READWISE_TOKEN: Set ($(shell echo $(READWISE_TOKEN) | cut -c1-8)...)"; \
	else \
		echo "🔑 READWISE_TOKEN: Not set"; \
	fi
	@if [ -n "$(CUTOFF_DATETIME)" ]; then \
		echo "📅 CUTOFF_DATETIME: $(CUTOFF_DATETIME) ($(shell date -r $(CUTOFF_DATETIME) 2>/dev/null || date -d '@$(CUTOFF_DATETIME)' 2>/dev/null || echo 'Invalid timestamp'))"; \
	else \
		echo "📅 CUTOFF_DATETIME: Not set"; \
	fi

# Build the binary
build:
	go build -o readwise-to-newspaper main.go

# Run the application with current environment variables
run:
	@if [ -z "$(READWISE_TOKEN)" ]; then \
		echo "Error: READWISE_TOKEN is required"; \
		echo "Set it with: READWISE_TOKEN=your_token make run"; \
		echo "Or create a .env file with: make env-example"; \
		exit 1; \
	fi
	@if [ -z "$(CUTOFF_DATETIME)" ]; then \
		echo "Error: CUTOFF_DATETIME is required"; \
		echo "Use: make yesterday, make week, or set CUTOFF_DATETIME=timestamp"; \
		exit 1; \
	fi
	go run main.go

# Run with articles from yesterday
yesterday:
	@if [ -z "$(READWISE_TOKEN)" ]; then \
		echo "Error: READWISE_TOKEN is required"; \
		echo "Usage: READWISE_TOKEN=your_token make yesterday"; \
		echo "Or create a .env file with: make env-example"; \
		exit 1; \
	fi
	READWISE_TOKEN="$(READWISE_TOKEN)" CUTOFF_DATETIME=$$(date -v-1d +%s 2>/dev/null || date -d "1 day ago" +%s) go run main.go

# Run with articles from last 7 days
week:
	@if [ -z "$(READWISE_TOKEN)" ]; then \
		echo "Error: READWISE_TOKEN is required"; \
		echo "Usage: READWISE_TOKEN=your_token make week"; \
		echo "Or create a .env file with: make env-example"; \
		exit 1; \
	fi
	READWISE_TOKEN="$(READWISE_TOKEN)" CUTOFF_DATETIME=$$(date -v-7d +%s 2>/dev/null || date -d "7 days ago" +%s) go run main.go

# Run with articles from last 30 days
month:
	@if [ -z "$(READWISE_TOKEN)" ]; then \
		echo "Error: READWISE_TOKEN is required"; \
		echo "Usage: READWISE_TOKEN=your_token make month"; \
		echo "Or create a .env file with: make env-example"; \
		exit 1; \
	fi
	READWISE_TOKEN="$(READWISE_TOKEN)" CUTOFF_DATETIME=$$(date -v-30d +%s 2>/dev/null || date -d "30 days ago" +%s) go run main.go

# Create a sample .env file
env-example:
	@if [ -f .env ]; then \
		echo ".env file already exists. Remove it first if you want to recreate it."; \
		echo "Current .env contents:"; \
		sed 's/^/  /' .env; \
	else \
		echo "# Readwise to Newspaper Configuration" > .env; \
		echo "# Get your token from: https://readwise.io/access_token" >> .env; \
		echo "READWISE_TOKEN=your_token_here" >> .env; \
		echo "" >> .env; \
		echo "# Unix timestamp for earliest article date" >> .env; \
		echo "# Examples:" >> .env; \
		echo "#   Yesterday: $$(date -v-1d +%s 2>/dev/null || date -d '1 day ago' +%s)" >> .env; \
		echo "#   Last week: $$(date -v-7d +%s 2>/dev/null || date -d '7 days ago' +%s)" >> .env; \
		echo "#   Last month: $$(date -v-30d +%s 2>/dev/null || date -d '30 days ago' +%s)" >> .env; \
		echo "CUTOFF_DATETIME=$$(date -v-7d +%s 2>/dev/null || date -d '7 days ago' +%s)" >> .env; \
		echo ""; \
		echo "✅ Created .env file with sample configuration"; \
		echo "📝 Edit .env file with your actual Readwise token"; \
		echo ""; \
		echo "Generated .env contents:"; \
		sed 's/^/  /' .env; \
	fi

# Run tests
test:
	go test -v ./...

# Clean up generated files
clean:
	rm -f readwise-to-newspaper
	rm -f daily-tech-digest.html

# Install dependencies
deps:
	go mod download
	go mod tidy 