package pagination

type CommitStrategy struct {
	params CommitQueryParams
}

func NewCommitStrategy(since, until string) *CommitStrategy {
	return &CommitStrategy{
		params: CommitQueryParams{
			Since:   since,
			Until:   until,
			PerPage: 100,
		},
	}
}

func (s *CommitStrategy) InitializeParams() interface{} {
	return s.params
}

func (s *CommitStrategy) UpdateParams(page int) (interface{}, error) {
	s.params.Page = page
	return s.params, nil
}
