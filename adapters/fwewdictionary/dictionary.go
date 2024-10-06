package fwewdictionary

import (
	"context"
	"errors"
	"fmt"
	fwew "github.com/fwew/fwew-lib/v5"
	"github.com/gissleh/sarfya"
	"slices"
	"strings"
	"sync"
)

var globalLock sync.Mutex

type dictionary struct {
	mu    sync.Mutex
	cache map[string]sarfya.DictionaryEntry
}

func (d *dictionary) Entry(ctx context.Context, id string) (*sarfya.DictionaryEntry, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	if d.cache == nil {
		globalLock.Lock()
		allWords, err := fwew.GetFullDict()
		globalLock.Unlock()
		if err != nil {
			return nil, err
		}

		for _, word := range allWords {
			d.cache[word.ID] = d.wordToEntry(word)
		}
	}

	if entry, ok := d.cache[id]; ok {
		entryCopy := entry.Copy()
		return &entryCopy, nil
	}

	return nil, errors.New("entry not found")
}

func (d *dictionary) Lookup(ctx context.Context, word string) (entries []sarfya.DictionaryEntry, err error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	defer func() {
		if err2 := recover(); err2 != nil {
			err = fmt.Errorf("panic: %v", err2)
		}
	}()

	if len(word) > 24 {
		return []sarfya.DictionaryEntry{}, nil
	}

	// Fwew gets a bit ornery with multi-thread access.
	globalLock.Lock()
	res, err := fwew.TranslateFromNaviHash(word, true)
	globalLock.Unlock()
	if err != nil {
		return nil, err
	}
	entries = make([]sarfya.DictionaryEntry, 0, 8)

	hasNumber := false
	for _, matches := range res {
		for _, match := range matches {
			if match.ID == "" {
				continue
			}
			if match.PartOfSpeech == "num." {
				hasNumber = true
			}

			entries = append(entries, d.wordToEntry(match))
		}
	}

	if !hasNumber {
		if n, err := fwew.NaviToNumber(word); err == nil && n > 0 {
			e := d.numberToEntry(n)
			if e.Word == strings.ToLower(word) {
				entries = append(entries, e)
			}
		}
		if strings.HasPrefix(word, "a") {
			if n, err := fwew.NaviToNumber(word[1:]); err == nil && n > 0 {
				e := d.numberToEntry(n)
				if e.Word == strings.ToLower(word[1:]) {
					e.Prefixes = []string{"a"}
					entries = append(entries, e)
				}
			}
		}
		if strings.HasSuffix(word, "a") {
			if n, err := fwew.NaviToNumber(word[:len(word)-1]); err == nil && n > 0 {
				e := d.numberToEntry(n)
				if e.Word == strings.ToLower(word[:len(word)-1]) {
					e.Suffixes = []string{"a"}
					entries = append(entries, e)
				}
			}
		}
	}

	return entries, nil
}

func (d *dictionary) numberToEntry(n int) sarfya.DictionaryEntry {
	numString, _ := fwew.NumberToNavi(n)

	return sarfya.DictionaryEntry{
		ID:          fmt.Sprintf("N%d", n),
		Word:        numString,
		PoS:         "num.",
		OriginalPoS: "num.",
		Definitions: map[string]string{
			"en": fmt.Sprint(n),
		},
		Prefixes:  []string{},
		Infixes:   []string{},
		Suffixes:  []string{},
		Lenitions: []string{},
		Comment:   []string{},
	}
}

func (d *dictionary) wordToEntry(match fwew.Word) sarfya.DictionaryEntry {
	var infixPositions []int
	if match.InfixDots != "NULL" && match.InfixDots != "" {
		if i := strings.Index(match.InfixDots, "."); i != -1 {
			if j := strings.LastIndex(match.InfixDots, "."); j != -1 {
				infixPositions = []int{i, j - 1}
			}
		}
	}

	var suffixes []string
	if match.Affixes.Suffix != nil {
		suffixes = append(match.Affixes.Suffix[:0:0], match.Affixes.Suffix...)
		slices.Reverse(suffixes)
	}

	entry := sarfya.DictionaryEntry{
		ID:           match.ID,
		Word:         match.Navi,
		PoS:          match.PartOfSpeech,
		OriginalPoS:  match.PartOfSpeech,
		Source:       match.Source,
		InfixIndexes: infixPositions,
		Definitions: map[string]string{
			"de": match.DE,
			"en": match.EN,
			"es": match.ES,
			"et": match.ET,
			"fr": match.FR,
			"hu": match.HU,
			"nl": match.NL,
			"pl": match.PL,
			"pt": match.PT,
			"ru": match.RU,
			"sv": match.SV,
			"tr": match.TR,
			"uk": match.UK,
		},
		Prefixes:  match.Affixes.Prefix,
		Infixes:   match.Affixes.Infix,
		Suffixes:  suffixes,
		Lenitions: match.Affixes.Lenition,
		Comment:   match.Affixes.Comment,
	}

	for k, v := range entry.Definitions {
		if v == "" || (k != "en" && v == entry.Definitions["en"]) {
			delete(entry.Definitions, k)
		}
	}

	return entry
}

var once sync.Once
var dict *dictionary

func Global() sarfya.Dictionary {
	once.Do(func() {
		fwew.StartEverything()
		dict = &dictionary{}
	})

	return dict
}
