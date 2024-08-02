package domain

import "time"

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
