package eventhook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
)

type ESPost struct{}

// Post sends event data to Elasticsearch
// url format: http://host:port/index/_doc or http://host:port/index/_doc/id
// Options can contain:
//   - "index": override index name from URL
//   - "id": document ID (if not in URL)
//   - "username": ES username for basic auth
//   - "password": ES password for basic auth
func (es *ESPost) Post(event *cloudevents.Event, url string, options map[string]string, data string, addtional string) error {
	if url == "" {
		return fmt.Errorf("elasticsearch URL is required")
	}

	// Parse URL to extract index and document ID
	index, docID := parseESURL(url)

	// Override index from options if provided
	if options != nil {
		if optIndex, ok := options["index"]; ok && optIndex != "" {
			index = optIndex
		}
		if optID, ok := options["id"]; ok && optID != "" {
			docID = optID
		}
	}

	// Replace date placeholders in index name
	if index != "" {
		index = replaceDatePlaceholders(index, event.Time())
	}

	// Build Elasticsearch document - only include original event data
	var doc map[string]interface{}
	if err := json.Unmarshal(event.Data(), &doc); err != nil {
		// If data is not JSON, create a document with the raw data
		doc = make(map[string]interface{})
		doc["data"] = string(event.Data())
	}

	// Marshal document to JSON
	docJSON, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("failed to marshal document: %w", err)
	}

	// Build Elasticsearch URL
	esURL := buildESURL(url, index, docID)

	// Create HTTP request
	req, err := http.NewRequest("POST", esURL, bytes.NewBuffer(docJSON))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Add basic auth if provided
	if options != nil {
		if username, ok := options["username"]; ok && username != "" {
			password := options["password"]
			req.SetBasicAuth(username, password)
		}
	}

	// Send request
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request to Elasticsearch: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// Read error response
		var errResp map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil {
			return fmt.Errorf("elasticsearch error (status %d): %v", resp.StatusCode, errResp)
		}
		return fmt.Errorf("elasticsearch error: status %d", resp.StatusCode)
	}

	return nil
}

// parseESURL parses Elasticsearch URL to extract index and document ID
// Examples:
//   - http://localhost:9200/my-index/_doc -> index: "my-index", id: ""
//   - http://localhost:9200/my-index/_doc/my-id -> index: "my-index", id: "my-id"
//   - http://localhost:9200 -> index: "", id: ""
func parseESURL(url string) (index, docID string) {
	// Remove protocol and host
	parts := strings.Split(url, "/")

	for i, part := range parts {
		if part == "_doc" {
			// Found _doc, previous part is index
			if i > 0 {
				index = parts[i-1]
			}
			// Next part after _doc is document ID
			if i+1 < len(parts) {
				docID = parts[i+1]
			}
			break
		}
	}

	return index, docID
}

// buildESURL builds Elasticsearch URL from base URL, index and document ID
func buildESURL(baseURL, index, docID string) string {
	// Remove trailing slash
	baseURL = strings.TrimSuffix(baseURL, "/")

	// If URL already contains /_doc, replace index if needed
	if strings.Contains(baseURL, "/_doc") {
		parts := strings.Split(baseURL, "/")
		for i, part := range parts {
			if part == "_doc" {
				// Replace index if provided
				if index != "" && i > 0 {
					parts[i-1] = index
				}
				// Add document ID if provided
				if docID != "" && (i+1 >= len(parts) || parts[i+1] == "") {
					parts = append(parts[:i+1], docID)
				} else if docID != "" && i+1 < len(parts) {
					parts[i+1] = docID
				}
				return strings.Join(parts, "/")
			}
		}
	}

	// Build URL from scratch
	if index == "" {
		return baseURL
	}

	esURL := fmt.Sprintf("%s/%s/_doc", baseURL, index)
	if docID != "" {
		esURL = fmt.Sprintf("%s/%s", esURL, docID)
	}

	return esURL
}

// replaceDatePlaceholders replaces date placeholders in index name with actual date values
// Supported placeholders:
//   - {date} or {YYYY.MM.DD} -> 2024.01.19
//   - {YYYY-MM-DD} -> 2024-01-19
//   - {YYYYMMDD} -> 20240119
//   - {YYYY.MM} -> 2024.01
//   - {YYYY-MM} -> 2024-01
//   - {YYYY} -> 2024
func replaceDatePlaceholders(index string, eventTime time.Time) string {
	year := eventTime.Format("2006")
	month := eventTime.Format("01")
	day := eventTime.Format("02")

	// Replace various date format placeholders
	index = strings.ReplaceAll(index, "{date}", fmt.Sprintf("%s.%s.%s", year, month, day))
	index = strings.ReplaceAll(index, "{YYYY.MM.DD}", fmt.Sprintf("%s.%s.%s", year, month, day))
	index = strings.ReplaceAll(index, "{YYYY-MM-DD}", fmt.Sprintf("%s-%s-%s", year, month, day))
	index = strings.ReplaceAll(index, "{YYYYMMDD}", fmt.Sprintf("%s%s%s", year, month, day))
	index = strings.ReplaceAll(index, "{YYYY.MM}", fmt.Sprintf("%s.%s", year, month))
	index = strings.ReplaceAll(index, "{YYYY-MM}", fmt.Sprintf("%s-%s", year, month))
	index = strings.ReplaceAll(index, "{YYYY}", year)

	return index
}
