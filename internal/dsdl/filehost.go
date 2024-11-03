package dsdl

import "github.com/playwright-community/playwright-go"

type FilehostImpl interface {
	PwPageNavigator

	EvaluateFileName() (string, error)
	EvaluateFileExt() (string, error)
	Download() error
}

type FilehostConstrFn func() *FilehostImpl

type Filehost struct {
	// open webpage
	p playwright.Page
	// conventional name
	Name string
	// builder function
	Constructor FilehostConstrFn
	// regexes tested against url
	AllowedUrlWildcards []string
}
