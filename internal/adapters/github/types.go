package github

import "time"

type Commit struct {
	Sha    string `json:"sha"`
	NodeId string `json:"node_id"`
	Commit struct {
		Committer struct {
			Name  string    `json:"name"`
			Email string    `json:"email"`
			Date  time.Time `json:"date"`
		} `json:"committer"`
		Message string `json:"message"`
		Tree    struct {
			Sha string `json:"sha"`
			Url string `json:"url"`
		} `json:"tree"`
		Url          string `json:"url"`
		CommentCount int    `json:"comment_count"`
		Verification struct {
			Verified  bool        `json:"verified"`
			Reason    string      `json:"reason"`
			Signature interface{} `json:"signature"`
			Payload   interface{} `json:"payload"`
		} `json:"verification"`
	} `json:"commit"`
}

// CommitQueryParams contains query parameters for fetching commits
type CommitQueryParams struct {
	Since   string `url:"since"`
	Until   string `url:"until"`
	Page    int    `url:"page"`
	PerPage int    `url:"per_page"`
}

type Repository struct {
	ID              int64     `db:"id" json:"-"`
	Owner           string    `db:"owner" json:"-"`
	Name            string    `db:"name" json:"name"`
	Description     string    `db:"description" json:"description"`
	URL             string    `db:"url" json:"url"`
	Language        string    `db:"language" json:"language"`
	ForksCount      int       `db:"forks_count" json:"forks_count"`
	StargazersCount int       `db:"stargazers_count" json:"stargazers_count"`
	OpenIssuesCount int       `db:"open_issues_count" json:"open_issues_count"`
	WatchersCount   int       `db:"watchers_count" json:"watchers_count"`
	CreatedAt       time.Time `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time `db:"updated_at" json:"updated_at"`
}
