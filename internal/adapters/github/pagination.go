package github

import (
	"context"
	"fmt"
	"github.com/olusolaa/github-monitor/pkg/httpclient"
	"net/http"
	"strings"
)

// PaginationManager manages paginated API requests.
type PaginationManager struct {
	requestBuilder  *httpclient.RequestBuilder
	requestExecutor *httpclient.Client
	responseHandler *httpclient.ResponseHandler
}

// NewPaginationManager creates a new instance of PaginationManager.
func NewPaginationManager(rb *httpclient.RequestBuilder, re *httpclient.Client, rh *httpclient.ResponseHandler) *PaginationManager {
	return &PaginationManager{
		requestBuilder:  rb,
		requestExecutor: re,
		responseHandler: rh,
	}
}

// FetchAllPages fetches all pages of a paginated API response and processes each page's data.
func (pm *PaginationManager) FetchAllPages(ctx context.Context,
	path string,
	params interface{},
	processPage func(interface{}) error,
	out interface{}) error {
	var fetchErr error

	for page := 1; ; page++ {
		switch p := params.(type) {
		case map[string]string:
			p["page"] = fmt.Sprintf("%d", page)
		default:
		}

		req, err := pm.requestBuilder.BuildRequest(ctx, http.MethodGet, path, params, nil)
		if err != nil {
			return fmt.Errorf("failed to build request for page %d: %w", page, err)
		}

		resp, err := pm.requestExecutor.Do(req)
		if err != nil {
			return fmt.Errorf("failed to get data for page %d: %w", page, err)
		}
		defer resp.Body.Close()

		if err = pm.responseHandler.HandleResponse(resp, out); err != nil {
			return fmt.Errorf("failed to process response for page %d: %w", page, err)
		}

		if err := processPage(out); err != nil {
			return fmt.Errorf("error processing data for page %d: %w", page, err)
		}

		// Check if there are more pages
		if !pm.HasNextPage(resp) {
			fmt.Printf("Last page reached: %d\n", page)
			break
		}

		// Check for context cancellation
		select {
		case <-ctx.Done():
			fmt.Println("Context done, stopping fetch.")
			return ctx.Err()
		default:
			// Continue fetching next page
		}
	}

	fmt.Println("All pages fetched successfully.")
	return fetchErr
}

// HasNextPage checks if there is a next page based on the Link header.
func (pm *PaginationManager) HasNextPage(resp *http.Response) bool {
	linkHeader := resp.Header.Get("Link")
	if linkHeader == "" {
		// No Link header found, assuming no more pages
		return false
	}
	// Check if "next" relation exists in the Link header
	return pm.parseLinkHeader(linkHeader, "next") != ""
}

// parseLinkHeader parses the Link header and returns the URL for the given relation (rel).
func (pm *PaginationManager) parseLinkHeader(header, rel string) string {
	links := strings.Split(header, ",")
	for _, link := range links {
		parts := strings.Split(link, ";")
		if len(parts) < 2 {
			continue
		}
		urlPart := strings.Trim(parts[0], " <>")
		relPart := strings.Trim(parts[1], " ")
		if strings.Contains(relPart, fmt.Sprintf(`rel="%s"`, rel)) {
			return urlPart
		}
	}
	return ""
}
