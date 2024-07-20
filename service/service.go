package service

import (
	"context"
	"encoding/base64"
	"errors"
	"github.com/gissleh/sarfya"
	"github.com/google/uuid"
	"sort"
	"sync"
	"sync/atomic"
)

type Service struct {
	Dictionary sarfya.Dictionary
	Storage    ExampleStorage
	ReadOnly   bool
}

func (s *Service) FindExample(ctx context.Context, id string) (*sarfya.Example, error) {
	return s.Storage.FindExample(ctx, id)
}

func (s *Service) QueryExample(ctx context.Context, filterString string) ([]ExampleGroup, error) {
	filter, resolvedMaps, err := sarfya.ParseFilter(ctx, filterString, s.Dictionary)
	if err != nil {
		return nil, err
	}

	res := make([]ExampleGroup, 0, len(resolvedMaps))
	for _, resolvedMap := range resolvedMaps {
		group := ExampleGroup{}

		seen := make(map[string]bool)

		examples := make([]sarfya.Example, 0, 16)

		if len(resolvedMap) > 0 {
			for _, entry := range resolvedMap {
				entryExamples, err := s.Storage.ListExamplesForEntry(ctx, entry.ID)
				if err != nil {
					return nil, err
				}

				for _, example := range entryExamples {
					if seen[example.ID] {
						continue
					}
					seen[example.ID] = true

					examples = append(examples, example)
				}
			}
		} else {
			if filter.SourceID != nil {
				examples, err = s.Storage.ListExamplesBySource(ctx, *filter.SourceID)
				if err != nil {
					return nil, err
				}
			} else {
				examples, err = s.Storage.ListExamples(ctx)
				if err != nil {
					return nil, err
				}
			}
		}

		wg := &sync.WaitGroup{}
		matches := make([]*sarfya.FilterMatch, len(examples))
		nextIndex := int32(-1)
		for i := 0; i < 3; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()

				i := int(atomic.AddInt32(&nextIndex, 1))
				for i < len(examples) {
					matches[i] = filter.CheckExample(examples[i], resolvedMap)
					i = int(atomic.AddInt32(&nextIndex, 1))
				}
			}()
		}
		wg.Wait()

		for _, match := range matches {
			if match != nil {
				group.Examples = append(group.Examples, *match)
			}
		}
		if len(group.Examples) == 0 {
			continue
		}

		sort.Slice(group.Examples, func(i, j int) bool {
			return group.Examples[i].Example.ListBefore(&group.Examples[j].Example)
		})

		for i := range filter.Terms {
			if entry, ok := resolvedMap[i]; ok {
				group.Entries = append(group.Entries, entry.Copy())
			}
		}
		res = append(res, group)
	}

	return res, nil
}

func (s *Service) SaveExample(ctx context.Context, input sarfya.Input, dry bool) (*sarfya.Example, error) {
	if s.ReadOnly {
		return nil, sarfya.ErrReadOnly
	}

	if input.Source.ID == "" || input.Source.Date == "" || input.Source.URL == "" {
		return nil, errors.New("missing fields in source")
	}

	example, err := sarfya.NewExample(ctx, input, s.Dictionary)
	if err != nil {
		return nil, err
	}

	if !dry {
		if example.ID == "" {
			id := uuid.New()
			example.ID = base64.RawURLEncoding.EncodeToString(id[:])
		}

		err = s.Storage.SaveExample(ctx, *example)
		if err != nil {
			return nil, err
		}
	}

	return example, nil
}

func (s *Service) DeleteExample(ctx context.Context, id string) (*sarfya.Example, error) {
	if s.ReadOnly {
		return nil, sarfya.ErrReadOnly
	}

	example, err := s.Storage.FindExample(ctx, id)
	if err != nil {
		return nil, err
	}

	err = s.Storage.DeleteExample(ctx, *example)
	if err != nil {
		return nil, err
	}

	return example, nil
}

type ExampleGroup struct {
	Entries  []sarfya.DictionaryEntry `json:"entries,omitempty"`
	Examples []sarfya.FilterMatch     `json:"examples"`
}
