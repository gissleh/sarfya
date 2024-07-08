package sourcestorage

import (
	"context"
	"fmt"
	"github.com/gissleh/sarfya"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"path"
	"sort"
	"strings"
	"sync"
)

type Storage struct {
	mu         sync.Mutex
	path       string
	examples   []sarfya.Example
	dictionary sarfya.Dictionary
}

func (s *Storage) FindExample(ctx context.Context, id string) (*sarfya.Example, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, example := range s.examples {
		if example.ID == id {
			exampleCopy := example.Copy()
			return &exampleCopy, nil
		}
	}

	return nil, sarfya.ErrExampleNotFound
}

func (s *Storage) ListExamples(ctx context.Context) ([]sarfya.Example, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	res := make([]sarfya.Example, 0, len(s.examples))
	for _, example := range s.examples {
		res = append(res, example.Copy())
	}

	sort.Slice(res, func(i, j int) bool { return res[i].ListBefore(&res[j]) })

	return res, nil
}

func (s *Storage) ListExamplesForEntry(ctx context.Context, entryID string) ([]sarfya.Example, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	s.mu.Lock()
	res := make([]sarfya.Example, 0, 16)
	for _, example := range s.examples {
	WordCheckLoop:
		for _, words := range example.Words {
			for _, word := range words {
				if word.ID == entryID {
					res = append(res, example.Copy())
					break WordCheckLoop
				}
			}
		}
	}
	s.mu.Unlock()

	sort.Slice(res, func(i, j int) bool { return res[i].ListBefore(&res[j]) })

	return res, nil
}

func (s *Storage) ListExamplesBySource(ctx context.Context, sourceID string) ([]sarfya.Example, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	s.mu.Lock()
	res := make([]sarfya.Example, 0, 32)
	for _, example := range s.examples {
		if example.Source.ID == sourceID {
			res = append(res, example.Copy())
		}
	}
	s.mu.Unlock()

	sort.Slice(res, func(i, j int) bool { return res[i].ListBefore(&res[j]) })

	return res, nil
}

func (s *Storage) SaveExample(ctx context.Context, example sarfya.Example) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for i, existing := range s.examples {
		if existing.ID == example.ID {
			prevSource := existing.Source
			s.examples[i] = example.Copy()

			err := s.save(example.Source)
			if err != nil {
				return err
			}

			if prevSource.ID != example.Source.ID {
				err := s.save(prevSource)
				if err != nil {
					return err
				}
			}

			return nil
		}
	}

	s.examples = append(s.examples, example)

	return s.save(example.Source)
}

func (s *Storage) DeleteExample(ctx context.Context, example sarfya.Example) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for i, existing := range s.examples {
		if existing.ID == example.ID {
			s.examples = append(s.examples[:i], s.examples[i+1:]...)
		}
	}

	return s.save(example.Source)
}

func (s *Storage) WriteAllFiles() error {
	sourceSeen := make(map[string]bool)
	sources := make([]sarfya.Source, 0, len(s.examples))

	s.mu.Lock()
	for _, example := range s.examples {
		if !sourceSeen[example.Source.ID] {
			sourceSeen[example.Source.ID] = true
			sources = append(sources, example.Source)
		}
	}
	s.mu.Unlock()

	for _, source := range sources {
		log.Println("Saving", source)
		err := s.save(source)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Storage) ExampleCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return len(s.examples)
}

func (s *Storage) save(source sarfya.Source) error {
	f, err := os.OpenFile(path.Join(s.path, source.ID+".yaml"), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	savedData := new(sourceFileData)
	savedData.Source = source
	for i, example := range s.examples {
		if example.Source.ID == source.ID {
			s.examples[i].Source = source
			input, err := example.MinimalInput(context.Background(), s.dictionary)
			if err != nil {
				return err
			}
			input.Source = sarfya.Source{}
			savedData.Inputs = append(savedData.Inputs, *input)
		}
	}

	return yaml.NewEncoder(f).Encode(savedData)
}

func Open(ctx context.Context, storagePath string, dictionary sarfya.Dictionary) (*Storage, error) {
	stat, err := os.Stat(storagePath)
	if os.IsNotExist(err) {
		err := os.MkdirAll(storagePath, 0766)
		if err != nil {
			return nil, err
		}

		return &Storage{path: storagePath, examples: []sarfya.Example{}}, nil
	} else if err != nil {
		return nil, err
	}

	if !stat.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", storagePath)
	}

	entries, err := os.ReadDir(storagePath)
	if err != nil {
		return nil, err
	}

	var examples []sarfya.Example
	for _, entry := range entries {
		if entry.IsDir() && !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}

		f, err := os.Open(path.Join(storagePath, entry.Name()))
		if err != nil {
			return nil, err
		}

		loadedData := new(sourceFileData)
		err = yaml.NewDecoder(f).Decode(loadedData)
		_ = f.Close()
		if err != nil {
			return nil, err
		}

		for _, input := range loadedData.Inputs {
			input.Source = loadedData.Source
			example, err := sarfya.NewExample(ctx, input, dictionary)
			if err != nil {
				return nil, fmt.Errorf("could not load example %s/%s: %w", entry.Name(), input.ID, err)
			}

			examples = append(examples, *example)
		}
	}

	return &Storage{path: storagePath, examples: examples, dictionary: dictionary}, nil
}

type sourceFileData struct {
	Source sarfya.Source  `yaml:"source"`
	Inputs []sarfya.Input `yaml:"inputs"`
}
