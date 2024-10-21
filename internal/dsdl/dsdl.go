package dsdl

import (
	"context"
	"fmt"
)

type (
	AggrConstrFn     func() *Aggregator
	FilehostConstrFn func() *Filehost

	Aggregators map[string]AggrConstrFn
	Filehosts   map[string]FilehostConstrFn
)

type DSDL struct {
	// queue & tracker proxy
	Tq *TQProxy

	Aggregators Aggregators
	Filehosts   Filehosts

	// whole application's context
	Ctx context.Context
}

func NewDSDL(ctx context.Context) *DSDL {
	dsdl := &DSDL{
		Aggregators: make(Aggregators),
		Filehosts:   make(Filehosts),
	}

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

func (dsdl *DSDL) RegisterAggregator(name string, constructor AggrConstrFn) error {
	if len(dsdl.Aggregators) == 0 {
		dsdl.Aggregators[name] = constructor
		return nil
	}

	unique := true
	for k := range dsdl.Aggregators {
		if k == name {
			unique = false
		}
	}

	if !unique {
		return fmt.Errorf("Aggregator is already registered")
	}

	dsdl.Aggregators[name] = constructor

	return nil
}

func (dsdl *DSDL) IsValidAggregator(name string) bool {
	for k := range dsdl.Aggregators {
		if k == name {
			return true
		}
	}

	return false
}

func (dsdl *DSDL) RegisterFilehost(name string, constructor FilehostConstrFn) error {
	if len(dsdl.Filehosts) == 0 {
		dsdl.Filehosts[name] = constructor
		return nil
	}

	unique := true
	for k := range dsdl.Filehosts {
		if k == name {
			unique = false
		}
	}

	if !unique {
		return fmt.Errorf("Filehost is already registered")
	}

	dsdl.Filehosts[name] = constructor

	return nil
}

func (dsdl *DSDL) IsValidFilehost(name string) bool {
	for k := range dsdl.Filehosts {
		if k == name {
			return true
		}
	}

	return false
}
