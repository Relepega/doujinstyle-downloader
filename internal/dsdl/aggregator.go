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

// alias for
// func(slug string, p playwright.Page) AggregatorImpl
type AggregatorConstrFn func(slug string, p playwright.Page) AggregatorImpl

const AGGR_ERR_UNAVAILABLE_FT = "This aggregator cannot evaluate a filetype extension"

type Aggregator struct {
	// conventional name
	Name string
	// builder function
	Constructor AggregatorConstrFn
	// regexes tested against url
	AllowedUrlWildcards []string
}
