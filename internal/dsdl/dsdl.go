package dsdl

import (
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
	db *db.SQLiteDB

	aggregators Aggregators
	filehosts   Filehosts

	browser playwright.Browser
}

func NewDSDL(browser playwright.Browser) *DSDL {
	dsdl := &DSDL{}

	// start browser
	pw, err := playwright.Run()
	if err != nil {
		log.Fatalln("could not start playwright: ", err)
	}

	t := true
	tout := 0.0

	if browser != nil {
		dsdl.browser = browser
	} else {
		b, err := pw.Chromium.Launch(
			playwright.BrowserTypeLaunchOptions{
				Headless:      &t,
				Timeout:       &tout,
				HandleSIGHUP:  &t,
				HandleSIGINT:  &t,
				HandleSIGTERM: &t,
			},
		)
		if err != nil {
			defer log.Fatalln("Couldn't start a new browser: ", err)
		}

		dsdl.browser = b
	}

	dsdl.browser = browser

	// start database
	sqlite := db.NewSQLite(false)
	dsdl.db = restoreDB(sqlite)

	return dsdl
}

func (dsdl *DSDL) Shutdown() error {
	log.Println("DSDL: Started shutdown procedure")

	_ = dsdl.Browser().Close()
	// if err != nil && err.Error() != "Connection closed" {
	// 	return fmt.Errorf("DSDL: An error occurred while shutting down playwright: %v", err)
	// }

	err := dsdl.db.Close()
	if err != nil {
		return fmt.Errorf("DSDL: An error occurred while shutting down database: %v", err)
	}

	log.Println("DSDL: Shutdown successful")

	return nil
}

func (dsdl *DSDL) Browser() playwright.Browser { return dsdl.browser }

func (dsdl *DSDL) DB() *db.SQLiteDB { return dsdl.db }

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
