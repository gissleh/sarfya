package sarfya

import (
	"context"
	"errors"
	"strings"
)

type testDictionary map[string]DictionaryEntry

func (t testDictionary) Entry(_ context.Context, id string) (*DictionaryEntry, error) {
	for _, entry := range t {
		if entry.ID == id {
			entryCopy := entry.Copy()
			return &entryCopy, nil
		}
	}

	return nil, errors.New("not found")
}

func (t testDictionary) Lookup(_ context.Context, word string) ([]DictionaryEntry, error) {
	if val, ok := t[strings.ToLower(word)]; ok {
		return []DictionaryEntry{val.Copy()}, nil
	}

	return nil, errors.New("not found")
}
