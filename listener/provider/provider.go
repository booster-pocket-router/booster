package provider

import (
	"context"

	"github.com/booster-proj/core"
)

type Confidence int

const (
	Low Confidence = iota
	High
)

type Merged struct {
	ErrHook func(ref, network, address string, err error)
}

func (m *Merged) Provide(ctx context.Context, level Confidence) ([]core.Source, error) {
	interfaces, err := new(Local).provide(ctx, level)
	if err != nil {
		return []core.Source{}, err
	}

	sources := make([]core.Source, 0, len(interfaces))
	for _, v := range interfaces {
		v.ErrHook = m.ErrHook
		sources = append(sources, v)
	}
	return sources, nil
}
