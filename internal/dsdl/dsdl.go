package dsdl

import (
	"context"
	"fmt"
	"log"
	"regexp"

	"github.com/playwright-community/playwright-go"

	"github.com/relepega/doujinstyle-downloader/internal/dsdl/db"
)

type (
	Aggregators []*Aggregator
	Filehosts   []*Filehost
)

type PwPageNavigator interface {
	Page() playwright.Page
}

const (
	ERR_REGISTERED_AGGREGATOR = "Aggregator is already registered"
	ERR_REGISTERED_FILEHOST   = "Filehost is already registered"
)

type DSDL struct {
	// queue & tracker proxy
	Tq *TQProxy

	aggregators Aggregators
	filehosts   Filehosts

	browser playwright.Browser

	// whole application's context
	Ctx context.Context
}

func NewDSDL(ctx context.Context) *DSDL {
	dsdl := &DSDL{}

	dsdl.Ctx = context.WithValue(ctx, "dsdl", dsdl)

	pw, err := playwright.Run()
	if err != nil {
		log.Fatalln("could not start playwright: ", err)
	}

	t := true
	tout := 0.0

	browser, err := pw.Chromium.Launch(
		playwright.BrowserTypeLaunchOptions{
			Headless:      &t,
			Timeout:       &tout,
			HandleSIGHUP:  &t,
			HandleSIGINT:  &t,
			HandleSIGTERM: &t,
		},
	)
	if err != nil {
		log.Fatalln("Couldn't start a new browser: ", err)
	}

	dsdl.browser = browser

	return dsdl
}

func NewDSDLWithBrowser(ctx context.Context, browser playwright.Browser) *DSDL {
	dsdl := &DSDL{
		browser: browser,
	}

	dsdl.Ctx = context.WithValue(ctx, "dsdl", dsdl)

	return dsdl
}

func (dsdl *DSDL) NewTQProxy(dbType db.DBType, fn QueueRunner) {
	if dsdl.Tq != nil {
		return
	}

	dsdl.Tq = newTQWrapperFromEngine(dbType, fn, dsdl.Ctx, dsdl)
}

func (dsdl *DSDL) GetTQProxy() *TQProxy {
	return dsdl.Tq
}

func (dsdl *DSDL) GetBrowserInstance() playwright.Browser {
	return dsdl.browser
}

func (dsdl *DSDL) RegisterAggregator(f *Aggregator) error {
	unique := true

	if len(dsdl.filehosts) == 0 {
		goto addAggr
	}

	for _, v := range dsdl.filehosts {
		if v.Name == f.Name {
			unique = false
		}
	}

	if !unique {
		return fmt.Errorf(ERR_REGISTERED_AGGREGATOR)
	}

addAggr:
	dsdl.aggregators = append(dsdl.aggregators, f)

	return nil
}

func (dsdl *DSDL) IsValidAggregator(name string) bool {
	if len(dsdl.aggregators) == 0 {
		return false
	}

	for _, v := range dsdl.aggregators {
		if v.Name == name {
			return true
		}
	}

	return false
}

func (dsdl *DSDL) EvaluateAggregator(aggrID string) (AggregatorConstrFn, error) {
	if len(dsdl.aggregators) == 0 {
		return nil, fmt.Errorf("Cannot evaluate aggregator from empty registration list")
	}

	for _, v := range dsdl.aggregators {
		if v.Name == aggrID {
			return v.Constructor, nil
		}
	}

	return nil, fmt.Errorf("Aggregator not found:\"%s\"", aggrID)
}

func (dsdl *DSDL) EvaluateAggregatorFromUrl(url string) (AggregatorConstrFn, error) {
	if len(dsdl.aggregators) == 0 {
		return nil, fmt.Errorf("Cannot evaluate aggregator from empty registration list")
	}

	for _, v := range dsdl.aggregators {
		for _, wildcard := range v.AllowedUrlWildcards {
			r, _ := regexp.Compile(wildcard)

			if r.MatchString(url) {
				return v.Constructor, nil
			}
		}
	}

	return nil, fmt.Errorf("Aggregator not found for this url: \"%s\"", url)
}

func (dsdl *DSDL) RegisterFilehost(f *Filehost) error {
	unique := true

	if len(dsdl.filehosts) == 0 {
		goto addFh
	}

	for _, v := range dsdl.filehosts {
		if v.Name == f.Name {
			unique = false
		}
	}

	if !unique {
		return fmt.Errorf(ERR_REGISTERED_FILEHOST)
	}

addFh:
	dsdl.filehosts = append(dsdl.filehosts, f)

	return nil
}

func (dsdl *DSDL) IsValidFilehost(name string) bool {
	if len(dsdl.filehosts) == 0 {
		return false
	}

	for _, v := range dsdl.filehosts {
		if v.Name == name {
			return true
		}
	}

	return false
}

func (dsdl *DSDL) EvaluateFilehost(url string) (FilehostConstrFn, error) {
	if len(dsdl.filehosts) == 0 {
		return nil, fmt.Errorf("Cannot evaluate filehost from empty registration list")
	}

	for _, v := range dsdl.filehosts {
		for _, wildcard := range v.AllowedUrlWildcards {
			r, _ := regexp.Compile(wildcard)

			if r.MatchString(url) {
				return v.Constructor, nil
			}
		}
	}

	return nil, fmt.Errorf("Filehost not found for this url: \"%s\"", url)
}
