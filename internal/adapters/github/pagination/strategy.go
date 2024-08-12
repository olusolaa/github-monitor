package pagination

type Strategy interface {
	InitializeParams() interface{}
	UpdateParams(page int) (interface{}, error)
}
