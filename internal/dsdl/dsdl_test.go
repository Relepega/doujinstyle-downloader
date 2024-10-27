package dsdl

import (
	"context"
	"fmt"
	"log"
	"testing"
)

func TestProperFunctioning(t *testing.T) {
	ctx := context.Background()

	d := NewDSDL(ctx)

	aggregatorUrl := "test"

	// get aggregator
	aggregator, err := d.EvaluateAggregator(aggregatorUrl)
	if err != nil {
		log.Fatalf("Filehost not found matching this url: \"%s\"", aggregatorUrl)
	}

	// get filehost url
	filehostUrl, err := aggregator.EvaluateDownloadUrl()
	if err != nil {
		log.Fatalf("Cannot evaluate a filehost url from this aggregator link", aggregatorUrl)
	}

	// get filehost
	filehost, err := d.EvaluateFilehost(filehostUrl)
	if err != nil {
		log.Fatalf("Filehost not found matching this url: \"%s\"", "")
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
}
