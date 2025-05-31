# Readwise to Newspaper

A Go application that fetches articles from your Readwise Reader using the [Readwise Reader API](https://readwise.io/reader_api) and generates a beautiful newspaper-style HTML document called "Daily Tech Digest".

## Features

- Fetches articles from Readwise Reader since a specified timestamp
- Includes HTML content, title, author, published date, and summary
- Generates a responsive newspaper-style layout with:
  - Professional typography using Times New Roman
  - Two-column layout (single column on mobile)
  - Clean formatting with proper spacing and typography
  - Embedded images from articles
  - Article summaries highlighted in blue boxes
  - Word count statistics
- Supports `.env` files for easy configuration

## Prerequisites

- Go 1.21 or later
- A Readwise account with Reader access
- Readwise API token (get yours at [readwise.io/access_token](https://readwise.io/access_token))

## Quick Start

1. Clone this repository:
```bash
git clone <repository-url>
cd readwise-to-newspaper
```

2. Install dependencies:
```bash
make deps
```

3. Create a `.env` file with your configuration:
```bash
make env-example
# Edit .env file with your actual token
```

4. Run the application:
```bash
# Generate newspaper with articles from the last 7 days
make week

# Or from yesterday
make yesterday

# Or from the last 30 days
make month
```

## Usage

### Using Make Commands (Recommended)

The easiest way to use the application is with the provided Make targets:

```bash
# Show all available commands
make help

# Quick runs with your token
READWISE_TOKEN=your_token_here make yesterday
READWISE_TOKEN=your_token_here make week
READWISE_TOKEN=your_token_here make month

# Using .env file (create it first with: make env-example)
make week  # Uses token from .env file
```

### Using Environment Variables

The application requires two environment variables:

- `READWISE_TOKEN`: Your Readwise API access token
- `CUTOFF_DATETIME`: Unix timestamp indicating the earliest date to fetch articles from

```bash
# Set your Readwise token
export READWISE_TOKEN="your_token_here"

# Set cutoff to fetch articles from the last 7 days
export CUTOFF_DATETIME=$(date -d "7 days ago" +%s)

# Run the application
go run main.go
```

### Using .env Files

Create a `.env` file in the project root:

```bash
# Create sample .env file
make env-example
```

Edit the `.env` file:
```env
READWISE_TOKEN=your_actual_token_here
CUTOFF_DATETIME=1748629267  # Unix timestamp
```

Then run:
```bash
make run
# or
go run main.go
```

### Getting Unix Timestamps

You can generate unix timestamps using various methods:

```bash
# Last 24 hours
date -d "1 day ago" +%s    # Linux
date -v-1d +%s             # macOS

# Last week
date -d "1 week ago" +%s   # Linux
date -v-7d +%s             # macOS

# Specific date (January 1, 2024)
date -d "2024-01-01" +%s   # Linux
date -j -f "%Y-%m-%d" "2024-01-01" "+%s"  # macOS
```

## Available Make Targets

- `make help` - Show all available commands
- `make yesterday` - Fetch articles from yesterday
- `make week` - Fetch articles from last 7 days
- `make month` - Fetch articles from last 30 days
- `make build` - Build the binary
- `make run` - Run with current environment variables
- `make env-example` - Create a sample .env file
- `make clean` - Clean up generated files
- `make test` - Run tests
- `make deps` - Install/update dependencies

## Output

The application generates a file called `daily-tech-digest.html` in the current directory. This file contains:

- A newspaper masthead with "Daily Tech Digest" title
- Current date and article statistics
- All articles formatted in newspaper columns
- Article metadata (author, published date, word count)
- Full HTML content including images
- Responsive design that works on desktop and mobile

## API Details

The application uses the Readwise Reader API v3 with the following endpoints:

- `GET /api/v3/list/` - Fetches documents with HTML content
- Includes pagination support for large collections
- Filters articles by `updatedAfter` timestamp
- Requests HTML content with `withHtmlContent=true`

## Error Handling

The application will exit with an error if:

- `READWISE_TOKEN` environment variable is not set
- `CUTOFF_DATETIME` environment variable is not set or invalid
- API requests fail (invalid token, network issues, etc.)
- File writing permissions are insufficient

## Rate Limiting

The Readwise API has rate limits of 20 requests per minute for most endpoints. The application handles pagination automatically and respects these limits.

## Security

- The application performs basic HTML sanitization by escaping script and style tags
- No external JavaScript is executed in the generated HTML
- Only uses the official Readwise API endpoints
- Supports `.env` files for secure token storage (remember to add `.env` to `.gitignore`)

## Contributing

Feel free to submit issues or pull requests to improve the newspaper layout, add features, or fix bugs.

## License

This project is open source. Please check the repository for license details.
