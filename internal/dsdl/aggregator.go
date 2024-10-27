package dsdl

type AggregatorImpl interface {
	FetchFileName() (string, error)
	FetchFileExt() (string, error)
	Download() error
}

type AggregatorConstrFn func() *AggregatorImpl

type Aggregator struct {
	// conventional name
	Name string
	// builder function
	Constructor AggregatorConstrFn
	// regexes tested against url
	AllowedUrlWildcards []string
}
