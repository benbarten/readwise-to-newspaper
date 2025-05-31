package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Document represents a Readwise document
type Document struct {
	ID            string      `json:"id"`
	Title         string      `json:"title"`
	Author        string      `json:"author"`
	Summary       string      `json:"summary"`
	PublishedDate interface{} `json:"published_date"`
	HTMLContent   string      `json:"html_content"`
	ImageURL      string      `json:"image_url"`
	WordCount     int         `json:"word_count"`
}

// ListResponse represents the API response for listing documents
type ListResponse struct {
	Count          int        `json:"count"`
	NextPageCursor string     `json:"nextPageCursor"`
	Results        []Document `json:"results"`
}

// ReadwiseClient handles API interactions
type ReadwiseClient struct {
	token   string
	baseURL string
}

// NewReadwiseClient creates a new Readwise API client
func NewReadwiseClient(token string) *ReadwiseClient {
	return &ReadwiseClient{
		token:   token,
		baseURL: "https://readwise.io/api/v3",
	}
}

// FetchDocuments retrieves all documents since the given timestamp
func (r *ReadwiseClient) FetchDocuments(updatedAfter time.Time) ([]Document, error) {
	var allDocuments []Document
	nextPageCursor := ""

	for {
		// Format date in UTC and URL encode it
		formattedDate := updatedAfter.UTC().Format("2006-01-02T15:04:05Z")
		baseURL := fmt.Sprintf("%s/list/", r.baseURL)

		// Build URL with parameters
		params := url.Values{}
		params.Add("withHtmlContent", "true")
		params.Add("updatedAfter", formattedDate)
		params.Add("location", "new")
		if nextPageCursor != "" {
			params.Add("pageCursor", nextPageCursor)
		}

		fullURL := baseURL + "?" + params.Encode()

		req, err := http.NewRequest("GET", fullURL, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Authorization", "Token "+r.token)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to make request: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
		}

		var listResp ListResponse
		if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}

		allDocuments = append(allDocuments, listResp.Results...)

		if listResp.NextPageCursor == "" {
			break
		}
		nextPageCursor = listResp.NextPageCursor
	}

	return allDocuments, nil
}

// NewspaperData represents the data for the newspaper template
type NewspaperData struct {
	Title          string
	Date           string
	Articles       []Article
	TotalWordCount int
}

// Article represents a processed article for the newspaper
type Article struct {
	Title         string
	Author        string
	Summary       string
	PublishedDate string
	Content       template.HTML
	WordCount     int
}

// ProcessDocuments converts Readwise documents to newspaper articles
func ProcessDocuments(documents []Document) NewspaperData {
	var articles []Article
	totalWordCount := 0

	for _, doc := range documents {
		if doc.Title == "" || doc.HTMLContent == "" {
			continue // Skip documents without title or content
		}

		// Clean and process the HTML content
		content := cleanHTML(doc.HTMLContent)

		// Format published date
		publishedDate := formatDate(doc.PublishedDate)

		article := Article{
			Title:         doc.Title,
			Author:        doc.Author,
			Summary:       doc.Summary,
			PublishedDate: publishedDate,
			Content:       template.HTML(content),
			WordCount:     doc.WordCount,
		}

		articles = append(articles, article)
		totalWordCount += doc.WordCount
	}

	return NewspaperData{
		Title:          "Daily Tech Digest",
		Date:           time.Now().Format("Monday, January 2, 2006"),
		Articles:       articles,
		TotalWordCount: totalWordCount,
	}
}

// cleanHTML removes dangerous tags and cleans up the HTML for display
func cleanHTML(html string) string {
	// Basic HTML cleaning - remove script and style tags
	html = strings.ReplaceAll(html, "<script", "&lt;script")
	html = strings.ReplaceAll(html, "</script>", "&lt;/script&gt;")
	html = strings.ReplaceAll(html, "<style", "&lt;style")
	html = strings.ReplaceAll(html, "</style>", "&lt;/style&gt;")
	return html
}

// formatDate formats a date string for display
func formatDate(dateValue interface{}) string {
	switch date := dateValue.(type) {
	case string:
		if date == "" {
			return ""
		}

		// Try to parse various date formats
		formats := []string{
			"2006-01-02T15:04:05Z07:00",
			"2006-01-02T15:04:05Z",
			"2006-01-02",
			time.RFC3339,
		}

		for _, format := range formats {
			if t, err := time.Parse(format, date); err == nil {
				return t.Format("January 2, 2006")
			}
		}

		return date
	case float64:
		return time.Unix(int64(date), 0).Format("January 2, 2006")
	case nil:
		return ""
	default:
		return ""
	}
}

// generatePDF creates a PDF from the HTML file using Chrome's headless mode
func generatePDF(htmlPath, pdfPath string) error {
	// Get absolute path for the HTML file
	absHTMLPath, err := filepath.Abs(htmlPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for HTML file: %w", err)
	}

	// Chrome executable path for macOS
	chromePath := "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"

	// Check if Chrome exists
	if _, err := os.Stat(chromePath); os.IsNotExist(err) {
		return fmt.Errorf("Google Chrome not found at %s", chromePath)
	}

	// Chrome arguments for PDF generation optimized for landscape printing with columns
	args := []string{
		"--headless",
		"--disable-gpu",
		"--no-sandbox",
		"--disable-dev-shm-usage",
		"--disable-extensions",
		"--disable-plugins",
		"--print-to-pdf=" + pdfPath,
		"--print-to-pdf-no-header",
		"--enable-print-preview",
		"--run-all-compositor-stages-before-draw",
		"--virtual-time-budget=15000", // Wait 15 seconds for page to load and render columns
		"--disable-background-timer-throttling",
		"--disable-renderer-backgrounding",
		"--disable-backgrounding-occluded-windows",
		"--force-color-profile=srgb",
		"--media=print", // Force print media queries
		"file://" + absHTMLPath,
	}

	// Execute Chrome command
	cmd := exec.Command(chromePath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to generate PDF: %w\nOutput: %s", err, string(output))
	}

	return nil
}

const newspaperTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}} - {{.Date}}</title>
    <style>
        body {
            font-family: 'Times New Roman', Times, serif;
            line-height: 1.4;
            margin: 0;
            padding: 10px;
            background-color: #fafafa;
            color: #333;
            font-size: 14px;
        }
        
        .newspaper {
            max-width: 1200px;
            margin: 0 auto;
            background: white;
            box-shadow: 0 0 20px rgba(0,0,0,0.1);
            padding: 20px;
        }
        
        .masthead {
            text-align: center;
            border-bottom: 2px solid #333;
            margin-bottom: 15px;
            padding-bottom: 10px;
        }
        
        .masthead h1 {
            font-size: 2.2em;
            font-weight: bold;
            margin: 0;
            letter-spacing: 1px;
            text-transform: uppercase;
        }
        
        .masthead .date {
            font-size: 1em;
            margin-top: 5px;
            font-style: italic;
        }
        
        .masthead .stats {
            font-size: 0.8em;
            margin-top: 3px;
            color: #666;
        }
        
        .articles {
            column-count: 3;
            column-gap: 25px;
            column-rule: 1px solid #ddd;
        }
        
        .article {
            break-inside: avoid;
            margin-bottom: 20px;
            padding-bottom: 10px;
            border-bottom: 1px solid #eee;
        }
        
        .article:last-child {
            border-bottom: none;
        }
        
        .article-header {
            margin-bottom: 8px;
        }
        
        .article-title {
            font-size: 1.3em;
            font-weight: bold;
            margin: 0 0 5px 0;
            line-height: 1.2;
        }
        
        .article-meta {
            font-size: 0.75em;
            color: #666;
            margin-bottom: 5px;
        }
        
        .article-meta .author {
            font-weight: bold;
        }
        
        .article-meta .date {
            margin-left: 8px;
        }
        
        .article-meta .word-count {
            margin-left: 8px;
            font-style: italic;
        }
        
        .article-summary {
            font-style: italic;
            margin-bottom: 8px;
            padding: 5px 8px;
            background-color: #f8f9fa;
            border-left: 3px solid #007bff;
            font-size: 0.9em;
            line-height: 1.3;
        }
        
        .article-content {
            text-align: justify;
            font-size: 0.85em;
            line-height: 1.3;
        }
        
        /* Normalize all headings and text within article content */
        .article-content h1,
        .article-content h2,
        .article-content h3,
        .article-content h4,
        .article-content h5,
        .article-content h6 {
            margin-top: 12px;
            margin-bottom: 5px;
            font-weight: bold;
            font-size: 1.1em !important; /* Force consistent heading size */
        }
        
        /* Reset any inherited font sizes for consistency */
        .article-content * {
            font-size: inherit !important;
        }
        
        /* Specific overrides for particular elements */
        .article-content h1 { font-size: 1.1em !important; }
        .article-content h2 { font-size: 1.05em !important; }
        .article-content h3 { font-size: 1.0em !important; }
        .article-content h4,
        .article-content h5,
        .article-content h6 { font-size: 0.95em !important; }
        
        .article-content p {
            margin-bottom: 6px;
            text-indent: 10px;
            font-size: 0.85em !important;
        }
        
        .article-content blockquote {
            margin: 10px 0;
            padding: 5px 10px;
            border-left: 3px solid #333;
            background-color: #f9f9f9;
            font-style: italic;
            font-size: 0.8em !important;
        }
        
        .article-content code {
            background-color: #f4f4f4;
            padding: 1px 3px;
            border-radius: 2px;
            font-family: 'Courier New', monospace;
            font-size: 0.8em !important;
        }
        
        .article-content pre {
            background-color: #f4f4f4;
            padding: 8px;
            border-radius: 3px;
            overflow-x: auto;
            margin: 8px 0;
            font-size: 0.75em !important;
        }
        
        .article-content ul,
        .article-content ol {
            margin: 8px 0;
            padding-left: 20px;
        }
        
        .article-content li {
            margin-bottom: 2px;
            font-size: 0.85em !important;
        }
        
        .article-content a {
            font-size: inherit !important;
        }
        
        .article-content span,
        .article-content div,
        .article-content em,
        .article-content strong,
        .article-content i,
        .article-content b {
            font-size: inherit !important;
        }
        
        .article-content img {
            max-width: 100%;
            height: auto;
            margin: 8px 0;
            border: 1px solid #ddd;
        }
        
        @media print {
            * {
                -webkit-print-color-adjust: exact !important;
                color-adjust: exact !important;
            }
            body { 
                background-color: white;
                padding: 0 !important;
                margin: 0 !important;
                font-size: 11px;
            }
            .newspaper { 
                box-shadow: none;
                padding: 2px !important;
                margin: 0 !important;
                width: 100%;
            }
            .masthead {
                margin: 0 !important;
                padding: 0 0 1px 0 !important;
                border-bottom: 1px solid #333;
            }
            .masthead h1 {
                font-size: 1.2em;
                margin: 0 !important;
                padding: 0 !important;
                letter-spacing: 0.5px;
            }
            .masthead .date {
                font-size: 0.8em;
                margin: 1px 0 0 0 !important;
                padding: 0 !important;
            }
            .masthead .stats {
                font-size: 0.7em;
                margin: 1px 0 0 0 !important;
                padding: 0 !important;
            }
            .articles {
                column-count: 4 !important;
                column-gap: 15px !important;
                column-rule: 1px solid #ddd !important;
                column-fill: auto !important;
                margin: 0 !important;
                padding: 0 !important;
                width: 100%;
                -webkit-column-count: 4 !important;
                -webkit-column-gap: 15px !important;
                -webkit-column-rule: 1px solid #ddd !important;
                -webkit-column-fill: auto !important;
            }
            .article {
                margin-bottom: 8px;
                padding-bottom: 3px;
                break-inside: avoid !important;
                -webkit-column-break-inside: avoid !important;
                page-break-inside: avoid !important;
                display: block;
            }
            .article:first-child {
                margin-top: 0 !important;
                padding-top: 0 !important;
            }
            .article-title {
                font-size: 1.1em;
                margin: 0 0 2px 0 !important;
            }
            .article-header {
                margin-bottom: 3px;
            }
            .article-meta {
                font-size: 0.65em;
                margin-bottom: 2px;
            }
            .article-content {
                font-size: 0.75em;
                line-height: 1.1;
            }
            .article-content p {
                margin-bottom: 2px;
                text-indent: 8px;
            }
            .article-content h1,
            .article-content h2,
            .article-content h3,
            .article-content h4,
            .article-content h5,
            .article-content h6 {
                margin-top: 6px;
                margin-bottom: 2px;
                font-size: 0.9em !important; /* Consistent heading size for print */
            }
            
            /* Normalize all text sizes in print */
            .article-content * {
                font-size: inherit !important;
            }
            
            .article-content p {
                font-size: 0.75em !important;
            }
            
            .article-content li {
                font-size: 0.75em !important;
            }
            
            .article-summary {
                padding: 1px 4px;
                margin-bottom: 2px;
                font-size: 0.7em !important;
                line-height: 1.2;
            }
            .article-content ul,
            .article-content ol {
                margin: 3px 0;
                padding-left: 15px;
            }
            .article-content li {
                margin-bottom: 1px;
            }
            .article-content blockquote {
                margin: 3px 0;
                padding: 2px 6px;
                font-size: 0.7em !important;
            }
            .article-content pre {
                padding: 3px;
                margin: 3px 0;
                font-size: 0.65em !important;
            }
            .article-content code {
                font-size: 0.7em !important;
            }
            .article-content img {
                margin: 3px 0;
            }
        }
        
        @media (max-width: 768px) {
            .articles {
                column-count: 1;
            }
            
            .masthead h1 {
                font-size: 2em;
            }
            
            .newspaper {
                padding: 15px;
            }
        }
    </style>
</head>
<body>
    <div class="newspaper">
        <header class="masthead">
            <h1>{{.Title}}</h1>
            <div class="date">{{.Date}}</div>
            <div class="stats">{{len .Articles}} articles • {{.TotalWordCount}} words total</div>
        </header>
        
        <main class="articles">
            {{range .Articles}}
            <article class="article">
                <div class="article-header">
                    <h2 class="article-title">{{.Title}}</h2>
                    <div class="article-meta">
                        {{if .Author}}<span class="author">By {{.Author}}</span>{{end}}
                        {{if .PublishedDate}}<span class="date">{{.PublishedDate}}</span>{{end}}
                        {{if .WordCount}}<span class="word-count">({{.WordCount}} words)</span>{{end}}
                    </div>
                </div>
                
                {{if .Summary}}
                <div class="article-summary">
                    {{.Summary}}
                </div>
                {{end}}
                
                <div class="article-content">
                    {{.Content}}
                </div>
            </article>
            {{end}}
        </main>
    </div>
</body>
</html>`

func main() {
	// Load .env file if it exists (ignore error if file doesn't exist)
	if err := godotenv.Load(); err != nil {
		// Only log if it's not a "file not found" error
		if !os.IsNotExist(err) {
			log.Printf("Warning: Error loading .env file: %v", err)
		}
	}

	// Get environment variables
	token := os.Getenv("READWISE_TOKEN")
	if token == "" {
		fmt.Fprintf(os.Stderr, "Error: READWISE_TOKEN environment variable is required\n")
		fmt.Fprintf(os.Stderr, "You can either set it as an environment variable or create a .env file with:\n")
		fmt.Fprintf(os.Stderr, "READWISE_TOKEN=your_token_here\n")
		fmt.Fprintf(os.Stderr, "\nGet your token from: https://readwise.io/access_token\n")
		os.Exit(1)
	}

	cutoffDatetimeStr := os.Getenv("CUTOFF_DATETIME")
	if cutoffDatetimeStr == "" {
		fmt.Fprintf(os.Stderr, "Error: CUTOFF_DATETIME environment variable is required\n")
		fmt.Fprintf(os.Stderr, "You can either set it as an environment variable or create a .env file with:\n")
		fmt.Fprintf(os.Stderr, "CUTOFF_DATETIME=1234567890  # Unix timestamp\n")
		fmt.Fprintf(os.Stderr, "\nTo get a timestamp for 7 days ago, run:\n")
		fmt.Fprintf(os.Stderr, "  date -d '7 days ago' +%%s  # Linux\n")
		fmt.Fprintf(os.Stderr, "  date -v-7d +%%s           # macOS\n")
		os.Exit(1)
	}

	// Parse the cutoff datetime (unix timestamp)
	cutoffTimestamp, err := strconv.ParseInt(cutoffDatetimeStr, 10, 64)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Invalid CUTOFF_DATETIME format. Expected unix timestamp: %v\n", err)
		os.Exit(1)
	}

	cutoffTime := time.Unix(cutoffTimestamp, 0)
	fmt.Printf("Fetching articles since %s\n", cutoffTime.Format(time.RFC3339))

	// Create Readwise client and fetch documents
	client := NewReadwiseClient(token)
	documents, err := client.FetchDocuments(cutoffTime)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching documents: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Fetched %d documents\n", len(documents))

	// Process documents into newspaper format
	newspaperData := ProcessDocuments(documents)

	// Generate HTML
	tmpl, err := template.New("newspaper").Parse(newspaperTemplate)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing template: %v\n", err)
		os.Exit(1)
	}

	// Write to output file
	outputFile := "daily-tech-digest.html"
	file, err := os.Create(outputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	err = tmpl.Execute(file, newspaperData)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error executing template: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Generated newspaper with %d articles\n", len(newspaperData.Articles))
	fmt.Printf("Output written to: %s\n", outputFile)

	// Generate PDF
	pdfPath := "daily-tech-digest.pdf"
	fmt.Printf("Generating PDF from HTML...\n")
	err = generatePDF(outputFile, pdfPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating PDF: %v\n", err)
		fmt.Fprintf(os.Stderr, "HTML file is still available at: %s\n", outputFile)
		os.Exit(1)
	}

	fmt.Printf("PDF generated and saved to: %s\n", pdfPath)
}
