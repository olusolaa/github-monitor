package pagination

// CommitQueryParams contains query parameters for fetching commits
type CommitQueryParams struct {
	Since   string `url:"since"`
	Until   string `url:"until"`
	Page    int    `url:"page"`
	PerPage int    `url:"per_page"`
}
