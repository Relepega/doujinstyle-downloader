package aggregator

import "github.com/relepega/doujinstyle-downloader/internal/dsdl"

type Doujinstyle struct {
	dsdl.Aggregator

	Url string
}

func (d *Doujinstyle) EvaluateFileName() (string, error) {}

func (d *Doujinstyle) EvaluateFileExt() (string, error) {}

func (d *Doujinstyle) EvaluateDownloadUrl() (string, error) {}
