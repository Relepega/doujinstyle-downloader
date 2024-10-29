package dsdl

type FilehostImpl interface {
	EvaluateFileName() (string, error)
	EvaluateFileExt() (string, error)
}

type FilehostConstrFn func() *FilehostImpl

type Filehost struct {
	// conventional name
	Name string
	// builder function
	Constructor FilehostConstrFn
	// regexes tested against url
	AllowedUrlWildcards []string
}
