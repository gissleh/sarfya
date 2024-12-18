package sarfyaservice

import (
	"context"
	"github.com/gissleh/sarfya"
)

type ExampleStorage interface {
	FindExample(ctx context.Context, id string) (*sarfya.Example, error)
	FetchExamples(ctx context.Context, filter *sarfya.Filter, resolved map[int]sarfya.DictionaryEntry) ([]sarfya.Example, error)
	SaveExample(ctx context.Context, example sarfya.Example) error
	DeleteExample(ctx context.Context, example sarfya.Example) error
}
