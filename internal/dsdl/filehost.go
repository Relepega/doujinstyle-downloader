package dsdl

import "github.com/playwright-community/playwright-go"

type FilehostImpl interface {
	PwPageNavigator

	EvaluateFileName() (string, error)
	EvaluateFileExt() (string, error)
	Download(tempDir, finalDir, filename string, progress *int8) error
}

type FilehostConstrFn func(p playwright.Page) FilehostImpl

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
