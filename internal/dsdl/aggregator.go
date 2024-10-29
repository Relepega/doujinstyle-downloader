package dsdl

import "github.com/playwright-community/playwright-go"

type AggregatorImpl interface {
	PwPageNavigator

	Url() string
	Is404() (bool, error)
	EvaluateFileName() (string, error)
	EvaluateFileExt() (string, error)
	EvaluateDownloadPage() (playwright.Page, error)
}

type AggregatorConstrFn func() AggregatorImpl

const AGGR_ERR_UNAVAILABLE_FT = "This aggregator cannot evaluate a filetype extension"

type Aggregator struct {
	// conventional name
	Name string
	// builder function
	Constructor AggregatorConstrFn
	// regexes tested against url
	AllowedUrlWildcards []string
}
