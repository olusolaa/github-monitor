package domain

import "time"

type Commit struct {
	ID           int64     `db:"id" json:"-"`
	RepositoryID int64     `db:"repository_id" json:"-"`
	Hash         string    `db:"hash" json:"hash"`
	Message      string    `db:"message" json:"message"`
	AuthorName   string    `db:"author_name" json:"author_name"`
	AuthorEmail  string    `db:"author_email" json:"author_email"`
	CommitDate   time.Time `db:"commit_date" json:"commit_date"`
	URL          string    `db:"url" json:"url"`
}

type CommitAuthor struct {
	AuthorName  string `json:"author_name" db:"author_name"`
	AuthorEmail string `json:"author_email" db:"author_email"`
	CommitCount int    `json:"commit_count" db:"commit_count"`
}
