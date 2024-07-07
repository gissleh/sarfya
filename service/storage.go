package service

import (
	"context"
	"github.com/gissleh/sarfya"
)

type ExampleStorage interface {
	FindExample(ctx context.Context, id string) (*sarfya.Example, error)
	ListExamples(ctx context.Context) ([]sarfya.Example, error)
	ListExamplesForEntry(ctx context.Context, entryID string) ([]sarfya.Example, error)
	ListExamplesBySource(ctx context.Context, sourceID string) ([]sarfya.Example, error)
	SaveExample(ctx context.Context, example sarfya.Example) error
	DeleteExample(ctx context.Context, example sarfya.Example) error
}
