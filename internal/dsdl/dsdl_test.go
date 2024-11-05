package dsdl

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"testing"
)

func TestProperFunctioning(t *testing.T) {
	ctx := context.Background()

	d := NewDSDL(ctx)

	aggregatorName := "test"

	// get aggregator
	aggregatorConstructor, err := d.EvaluateAggregator(aggregatorName)
	if err != nil {
		log.Fatalf("Filehost not found matching this url: \"%s\"", aggregatorName)
	}

	aggregator := aggregatorConstructor("112334", nil)

	// get filehost url
	filehostPage, err := aggregator.EvaluateDownloadPage()
	if err != nil {
		log.Fatalf(
			"Cannot evaluate a filehost url from this aggregator link: \"%s\"",
			aggregatorName,
		)
	}

	// get filehost
	filehost, err := d.EvaluateFilehost(filehostPage.URL())
	if err != nil {
		log.Fatalf("Filehost not found matching this url: \"%s\"", filehostPage.URL())
	}

	var fn string
	var fext string

	fn, err = filehost.EvaluateFileName()
	if err != nil {
		fn, err = aggregator.EvaluateFileName()
		if err != nil {
			log.Fatalln("Cannot evaluate a proper filename")
		}
	}

	fext, err = filehost.EvaluateFileExt()
	if err != nil {
		fext, err = aggregator.EvaluateFileExt()
		if err != nil {
			log.Fatalln("Cannot evaluate a proper file extension")
		}
	}

	filename := fmt.Sprintf("%s.%s", fn, fext)

	dlpath := filepath.Join(".", "test-downloads", filename)

	// this has to come from task
	var progress int8
	filehost.Download(dlpath, &progress)
}
