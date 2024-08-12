package pagination

import (
	"context"
	"fmt"
	"github.com/olusolaa/github-monitor/pkg/errors"
	"github.com/olusolaa/github-monitor/pkg/httpclient"
	"net/http"
	"strings"
)

// Manager manages paginated API requests.
type Manager struct {
	requestBuilder  *httpclient.RequestBuilder
	requestExecutor *httpclient.Client
	responseHandler *httpclient.ResponseHandler
}

// NewManager creates a new instance of PaginationManager.
func NewManager(rb *httpclient.RequestBuilder, re *httpclient.Client, rh *httpclient.ResponseHandler) *Manager {
	return &Manager{
		requestBuilder:  rb,
		requestExecutor: re,
		responseHandler: rh,
	}
}

func (pm *Manager) FetchAllPages(
	ctx context.Context,
	path string,
	strategy Strategy,
	processPage func(interface{}) error,
	out interface{}) error {
	var fetchErr error

	for page := 1; ; page++ {
		params, err := strategy.UpdateParams(page)
		if err != nil {
			fetchErr = err
			break
		}

		req, err := pm.requestBuilder.BuildRequest(ctx, http.MethodGet, path, params, nil)
		if err != nil {
			fetchErr = errors.New("BUILD_REQUEST_ERROR", fmt.Sprintf("failed to build request for page %d", page), err, errors.Critical)
			break
		}

		resp, err := pm.requestExecutor.Do(req)
		if err != nil {
			fetchErr = errors.New("REQUEST_EXECUTION_ERROR", fmt.Sprintf("failed to get data for page %d", page), err, errors.Critical)
			break
		}
		defer resp.Body.Close()

		if err = pm.responseHandler.HandleResponse(resp, out); err != nil {
			fetchErr = errors.New("RESPONSE_HANDLING_ERROR", fmt.Sprintf("failed to process response for page %d", page), err, errors.Critical)
			break
		}

		if err := processPage(out); err != nil {
			fetchErr = errors.New("PROCESS_PAGE_ERROR", fmt.Sprintf("error processing data for page %d", page), err, errors.Critical)
			break
		}

		// Check if there are more pages
		if !pm.HasNextPage(resp) {
			break
		}

		// Check for context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Continue fetching next page
		}
	}

	return fetchErr
}

// HasNextPage checks if there is a next page based on the Link header.
func (pm *Manager) HasNextPage(resp *http.Response) bool {
	linkHeader := resp.Header.Get("Link")
	if linkHeader == "" {
		// No Link header found, assuming no more pages
		return false
	}
	// Check if "next" relation exists in the Link header
	return pm.parseLinkHeader(linkHeader, "next") != ""
}

// parseLinkHeader parses the Link header and returns the URL for the given relation (rel).
func (pm *Manager) parseLinkHeader(header, rel string) string {
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
