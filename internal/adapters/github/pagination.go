package github

import (
	"context"
	"fmt"
	"github.com/olusolaa/github-monitor/pkg/httpclient"
	"net/http"
	"strings"
	"sync"
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
	var (
		wg       sync.WaitGroup
		mu       sync.Mutex
		fetchErr error
		lastPage bool
	)
	maxConcurrentRequests := 5
	sem := make(chan struct{}, maxConcurrentRequests)

	fetchPage := func(page int) {
		defer wg.Done()
		sem <- struct{}{} // Acquire semaphore
		defer func() { <-sem }()

		// Update page number in parameters
		switch p := params.(type) {
		case map[string]string:
			p["page"] = fmt.Sprintf("%d", page)
		default:
			// Handle other parameter types if necessary
		}

		// Build request
		req, err := pm.requestBuilder.BuildRequest(ctx, http.MethodGet, path, params, nil)
		if err != nil {
			mu.Lock()
			if fetchErr == nil {
				fetchErr = fmt.Errorf("failed to build request for page %d: %w", page, err)
			}
			mu.Unlock()
			return
		}

		// Check if req is nil to prevent further issues
		if req == nil {
			mu.Lock()
			if fetchErr == nil {
				fetchErr = fmt.Errorf("nil request returned for page %d", page)
			}
			mu.Unlock()
			return
		}

		// Execute request
		resp, err := pm.requestExecutor.Do(req)
		if err != nil {
			mu.Lock()
			if fetchErr == nil {
				fetchErr = fmt.Errorf("failed to get data for page %d: %w", page, err)
			}
			mu.Unlock()
			return
		}
		defer resp.Body.Close()

		// Handle response
		if err = pm.responseHandler.HandleResponse(resp, out); err != nil {
			mu.Lock()
			if fetchErr == nil {
				fetchErr = fmt.Errorf("failed to process response for page %d: %w", page, err)
			}
			mu.Unlock()
			return
		}

		// Process the page data
		if err := processPage(out); err != nil {
			mu.Lock()
			if fetchErr == nil {
				fetchErr = fmt.Errorf("error processing data for page %d: %w", page, err)
			}
			mu.Unlock()
			return
		}

		// Check if there are more pages
		if !pm.HasNextPage(resp) {
			mu.Lock()
			lastPage = true
			mu.Unlock()
		}
	}

	// Start fetching pages concurrently
	for page := 1; ; page++ {
		// Check if we should stop fetching more pages
		mu.Lock()
		if lastPage || fetchErr != nil {
			mu.Unlock()
			break
		}
		mu.Unlock()

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			wg.Add(1)
			go fetchPage(page)
		}
	}

	// Wait for all fetches to complete
	wg.Wait()
	return fetchErr
}

// HasNextPage checks if there is a next page based on the Link header.
func (pm *PaginationManager) HasNextPage(resp *http.Response) bool {
	linkHeader := resp.Header.Get("Link")
	return pm.parseLinkHeader(linkHeader, "next") != ""
}

// parseLinkHeader parses the Link header and returns the URL for the given relation (rel).
func (pm *PaginationManager) parseLinkHeader(header, rel string) string {
	// Example format: <https://api.github.com/repositories/1300192/commits?page=2>; rel="next"
	links := strings.Split(header, ",")
	for _, link := range links {
		parts := strings.Split(link, ";")
		if len(parts) < 2 {
			continue
		}
		urlPart := strings.Trim(parts[0], " <>")
		relPart := strings.Trim(parts[1], " ")
		if relPart == fmt.Sprintf("rel=%q", rel) {
			return urlPart
		}
	}
	return ""
}

type Manager struct {
	client Client
}

func NewManager(client Client) *Manager {
	return &Manager{
		client: client,
	}
}

func (m *Manager) FetchAllPages(
	ctx context.Context,
	path string,
	params map[string]string,
	processPage func(interface{}) error,
	out interface{},
) error {
	var (
		wg       sync.WaitGroup
		mu       sync.Mutex
		fetchErr error
		lastPage bool
	)
	maxConcurrentRequests := 5
	sem := make(chan struct{}, maxConcurrentRequests)

	fetchPage := func(page int) {
		defer wg.Done()
		sem <- struct{}{} // Acquire semaphore
		defer func() { <-sem }()

		// Update page number in parameters
		params["page"] = fmt.Sprintf("%d", page)

		// Build request
		req, err := m.client.requestBuilder.BuildRequest(ctx, http.MethodGet, path, params, nil)
		if err != nil {
			mu.Lock()
			fetchErr = err
			mu.Unlock()
			return
		}

		// Execute request
		resp, err := m.client.httpClient.Do(req)
		if err != nil {
			mu.Lock()
			fetchErr = err
			mu.Unlock()
			return
		}
		defer resp.Body.Close()

		// Handle response
		if err := m.client.responseHandler.HandleResponse(resp, out); err != nil {
			mu.Lock()
			fetchErr = err
			mu.Unlock()
			return
		}

		// Process the page data
		if err := processPage(out); err != nil {
			mu.Lock()
			fetchErr = err
			mu.Unlock()
			return
		}

		// Check if there are more pages
		if !m.HasNextPage(resp) {
			mu.Lock()
			lastPage = true
			mu.Unlock()
		}
	}

	// Start fetching pages concurrently
	for page := 1; ; page++ {
		mu.Lock()
		if lastPage || fetchErr != nil {
			mu.Unlock()
			break
		}
		mu.Unlock()

		wg.Add(1)
		go fetchPage(page)
	}

	// Wait for all fetches to complete
	wg.Wait()
	return fetchErr
}

func (m *Manager) HasNextPage(resp *http.Response) bool {
	linkHeader := resp.Header.Get("Link")
	return parseLinkHeader(linkHeader, "next") != ""
}

func parseLinkHeader(header, rel string) string {
	links := strings.Split(header, ",")
	for _, link := range links {
		parts := strings.Split(link, ";")
		if len(parts) < 2 {
			continue
		}
		urlPart := strings.Trim(parts[0], " <>")
		relPart := strings.Trim(parts[1], " ")
		if relPart == fmt.Sprintf("rel=%q", rel) {
			return urlPart
		}
	}
	return ""
}
