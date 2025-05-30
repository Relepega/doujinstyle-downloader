package dsdl

import "github.com/playwright-community/playwright-go"

type FilehostImpl interface {
	PwPageNavigator

	SetPage(p playwright.Page)
	Page() playwright.Page
	EvaluateFileName() (string, error)
	EvaluateFileExt() (string, error)
	Download(tempDir, finalDir, filename string, setProgress func(p int8)) error
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
