package jsonstorage

import (
	"context"
	"encoding/json"
	"github.com/gissleh/sarfya"
	"os"
	"sync"
)

func New(path string) *Storage {
	return &Storage{
		path:     path,
		readOnly: false,
		examples: make(map[string]sarfya.Example, 1024),
		index:    make(map[string][]string, 1024),
	}
}

func FromData(path string, readOnly bool, data Data) *Storage {
	for _, example := range data.Examples {
		for _, words := range example.Words {
			for i, word := range words {
				if data.DictDefs[word.ID] != nil {
					words[i].Definitions = data.DictDefs[word.ID]
				}
			}
		}
	}

	return &Storage{
		path:     path,
		readOnly: readOnly,
		examples: data.Examples,
		index:    data.Index,
	}
}

func Open(path string, readOnly bool) (*Storage, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var data Data
	if err := json.NewDecoder(file).Decode(&data); err != nil {
		return nil, err
	}

	return FromData(path, readOnly, data), nil
}

type Storage struct {
	mu       sync.Mutex
	path     string
	readOnly bool
	examples map[string]sarfya.Example
	index    map[string][]string
}

type Data struct {
	Examples map[string]sarfya.Example    `json:"examples"`
	Index    map[string][]string          `json:"index"`
	DictDefs map[string]map[string]string `json:"dictDefs"`
}

func (s *Storage) FindExample(ctx context.Context, id string) (*sarfya.Example, error) {
	if !s.readOnly {
		s.mu.Lock()
		defer s.mu.Unlock()
	}

	example, ok := s.examples[id]
	if !ok {
		return nil, sarfya.ErrExampleNotFound
	}

	example = example.Copy()
	return &example, nil
}

func (s *Storage) ListExamples(ctx context.Context) ([]sarfya.Example, error) {
	if !s.readOnly {
		s.mu.Lock()
		defer s.mu.Unlock()
	}

	res := make([]sarfya.Example, 0, len(s.examples))
	for _, example := range s.examples {
		res = append(res, example.Copy())
	}

	return res, nil
}

func (s *Storage) FetchExamples(ctx context.Context, filter *sarfya.Filter, resolved map[int]sarfya.DictionaryEntry) ([]sarfya.Example, error) {
	if !s.readOnly {
		s.mu.Lock()
		defer s.mu.Unlock()
	}

	res := make([]sarfya.Example, 0, len(s.examples))

	if filter != nil && filter.SourceID != nil {
		for _, id := range s.index["src:"+*filter.SourceID] {
			example := s.examples[id]
			res = append(res, example.Copy())
		}
	} else if filter == nil || filter.NeedFullList() {
		for _, example := range s.examples {
			res = append(res, example.Copy())
		}
	} else {
		strategy := filter.WordLookupStrategy(resolved)
		hasAdded := map[string]bool{}
		for _, entries := range strategy {
			if len(entries) == 0 {
				panic("filter.NeedFullList() is supposed to return true if one is empty")
			}

			var shortestList []string
			if len(entries) > 1 {
				for _, entry := range entries {
					if len(s.index[entry.ID]) < len(shortestList) || shortestList == nil {
						shortestList = s.index[entry.ID]
					}
				}
			}

			for _, id := range shortestList {
				if hasAdded[id] {
					continue
				}

				hasAdded[id] = true
				example := s.examples[id]
				res = append(res, example.Copy())
			}
		}
	}

	return res, nil
}

func (s *Storage) ListExamplesForEntry(ctx context.Context, entryID string) ([]sarfya.Example, error) {
	if !s.readOnly {
		s.mu.Lock()
		defer s.mu.Unlock()
	}

	res := make([]sarfya.Example, 0, len(s.index[entryID]))
	for _, exampleID := range s.index[entryID] {
		example := s.examples[exampleID]
		res = append(res, example.Copy())
	}

	return res, nil
}

func (s *Storage) ListExamplesBySource(ctx context.Context, sourceID string) ([]sarfya.Example, error) {
	return s.ListExamplesForEntry(ctx, "src:"+sourceID)
}

func (s *Storage) SaveExample(ctx context.Context, example sarfya.Example) error {
	if s.readOnly {
		return sarfya.ErrReadOnly
	}

	s.mu.Lock()
	s.unIndexExamples(example)
	s.examples[example.ID] = example.Copy()
	s.indexExamples(example)
	s.mu.Unlock()

	return nil
}

func (s *Storage) DeleteExample(ctx context.Context, example sarfya.Example) error {
	if s.readOnly {
		return sarfya.ErrReadOnly
	}

	if _, err := s.FindExample(ctx, example.ID); err != nil {
		return sarfya.ErrExampleNotFound
	}

	s.mu.Lock()
	s.unIndexExamples(example)
	delete(s.examples, example.ID)
	s.mu.Unlock()

	return nil
}

func (s *Storage) WriteToFile() error {
	data := Data{
		Examples: make(map[string]sarfya.Example, 1024),
		Index:    make(map[string][]string, 1024),
		DictDefs: make(map[string]map[string]string, 1024),
	}

	s.mu.Lock()
	for _, example := range s.examples {
		newWords := make(map[int][]sarfya.DictionaryEntry)
		for key, words := range example.Words {
			words := append(words[:0], words...)
			for i, word := range words {
				if word.ID == "" {
					continue
				}

				word := word.Copy()
				data.DictDefs[word.ID] = word.Definitions
				word.Definitions = nil
				words[i] = word
			}

			newWords[key] = words
		}
		example.Words = newWords

		data.Examples[example.ID] = example
	}
	for key, index := range s.index {
		data.Index[key] = append(make([]string, 0, len(index)), index...)
	}
	s.mu.Unlock()

	file, err := os.OpenFile(s.path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	enc := json.NewEncoder(file)

	return enc.Encode(data)
}

func (s *Storage) indexExamples(examples ...sarfya.Example) {
	seen := make(map[string]bool, 128)

	for _, example := range examples {
		for key := range seen {
			delete(seen, key)
		}

		for _, words := range example.Words {
			for _, word := range words {
				if seen[word.ID] {
					continue
				}

				seen[word.ID] = true
				s.index[word.ID] = append(s.index[word.ID], example.ID)
			}
		}

		s.index["src:"+example.Source.ID] = append(s.index["src:"+example.Source.ID], example.ID)
	}
}

func (s *Storage) unIndexExamples(examples ...sarfya.Example) {
	for _, example := range examples {
		if _, ok := s.examples[example.ID]; !ok {
			continue
		}

		for _, words := range example.Words {
			for _, word := range words {
				s.index[word.ID] = sliceWithout(s.index[word.ID], example.ID)
			}
		}

		s.index["src:"+example.Source.ID] = sliceWithout(s.index["src:"+example.Source.ID], example.ID)
	}
}
