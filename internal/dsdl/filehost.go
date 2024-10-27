package dsdl

type FilehostImpl interface {
	FetchFileName() (string, error)
	FetchFileExt() (string, error)
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
