package placeholderdictionary

import (
	"context"
	"github.com/gissleh/sarfya"
	"strings"
)

type placeholderDictionary struct{}

func (d *placeholderDictionary) Entry(_ context.Context, id string) (*sarfya.DictionaryEntry, error) {
	if len(id) == 2 && id[0] == 'P' && id[1] >= 'A' && id[1] <= 'Z' && id[1] != 'V' {
		return d.generateEntry(id[1:]), nil
	}

	return nil, sarfya.ErrDictionaryEntryNotFound
}

func (d *placeholderDictionary) Lookup(_ context.Context, search string, _ bool) ([]sarfya.DictionaryEntry, error) {
	chunks := strings.Split(search, "-")
	letter := ""
	prefixes := make([]string, 0)
	suffixes := make([]string, 0)

	for _, chunk := range chunks {
		if len(chunk) == 1 && chunk[0] >= 'A' && chunk[0] <= 'Z' && chunk[0] != 'V' {
			letter = chunk
			continue
		}

		if letter == "" {
			prefixes = append(prefixes, chunk)
		} else {
			suffixes = append(suffixes, chunk)
		}
	}

	if letter == "" {
		return []sarfya.DictionaryEntry{}, nil
	}

	entry := *d.generateEntry(letter)
	entry.Prefixes = prefixes
	entry.Suffixes = suffixes

	return []sarfya.DictionaryEntry{entry}, nil
}

func (d *placeholderDictionary) generateEntry(letter string) *sarfya.DictionaryEntry {
	pos := "n."
	if letter == "N" || letter == "M" {
		pos = "prop.n."
	}

	return &sarfya.DictionaryEntry{
		ID:   "P" + letter,
		Word: letter,
		PoS:  pos,
		Definitions: map[string]string{
			"en": "Placeholder",
			"de": "Platzhalter",
		},
		Source: "Placeholder (github.com/gissleh/sarfya/adapters/placeholderdictionary)",
	}
}

func New() sarfya.Dictionary {
	return &placeholderDictionary{}
}
