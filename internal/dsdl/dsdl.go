package dsdl

import (
	"context"
	"fmt"
)

type (
	Aggregators []*Aggregator
	Filehosts   []*Filehost
)

const (
	ERR_REGISTERED_AGGREGATOR = "Aggregator is already registered"
	ERR_REGISTERED_FILEHOST   = "Filehost is already registered"
)

type DSDL struct {
	// queue & tracker proxy
	Tq *TQProxy

	aggregators Aggregators
	filehosts   Filehosts

	// whole application's context
	Ctx context.Context
}

func NewDSDL(ctx context.Context) *DSDL {
	dsdl := &DSDL{}

	dsdl.Ctx = context.WithValue(ctx, "dsdl", dsdl)

	return dsdl
}

func (dsdl *DSDL) NewTQProxy(fn QueueRunner) {
	if dsdl.Tq != nil {
		return
	}

	dsdl.Tq = NewTQWrapper(fn, dsdl.Ctx)
}

func (dsdl *DSDL) GetTQProxy() *TQProxy {
	return dsdl.Tq
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
	for _, v := range dsdl.aggregators {
		if v.Name == name {
			return true
		}
	}

	return false
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
	for _, v := range dsdl.filehosts {
		if v.Name == name {
			return true
		}
	}

	return false
}
