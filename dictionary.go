package sarfya

import (
	"context"
	"errors"
	"strings"
)

type Dictionary interface {
	Entry(ctx context.Context, id string) (*DictionaryEntry, error)
	Lookup(ctx context.Context, search string) ([]DictionaryEntry, error)
}

type DictionaryEntry struct {
	ID          string            `json:"id,omitempty" yaml:"id,omitempty"`
	Word        string            `json:"word" yaml:"word"`
	PoS         string            `json:"pos" yaml:"poS"`
	Definitions map[string]string `json:"definitions" yaml:"definitions"`
	Source      string            `json:"source,omitempty" yaml:"source,omitempty"`
	Prefixes    []string          `json:"prefixes,omitempty" yaml:"prefixes,omitempty"`
	Infixes     []string          `json:"infixes,omitempty" yaml:"infixes,omitempty"`
	Suffixes    []string          `json:"suffixes,omitempty" yaml:"suffixes,omitempty"`
	Lenitions   []string          `json:"lenitions,omitempty" yaml:"lenitions,omitempty"`
	Comment     []string          `json:"comment,omitempty" yaml:"comment,omitempty"`
}

func (e *DictionaryEntry) HasPrefix(prefix string) bool {
	for _, p := range e.Prefixes {
		if p == prefix {
			return true
		}
	}

	return false
}

func (e *DictionaryEntry) HasSuffix(suffix string) bool {
	if alias, ok := suffixAliases[suffix]; ok {
		suffix = alias
	}

	for _, s := range e.Suffixes {
		if alias, ok := suffixAliases[s]; ok {
			s = alias
		}

		if s == suffix {
			return true
		}
	}

	return false
}

func (e *DictionaryEntry) HasInfix(infix string) bool {
	if aliasedInfix, ok := infixAliases[infix]; ok {
		infix = aliasedInfix
	}

	for _, i := range e.Infixes {
		if aliasedInfix, ok := infixAliases[i]; ok {
			i = aliasedInfix
		}

		if i == infix {
			return true
		}
	}

	return false
}

func (e *DictionaryEntry) HasLenition(lenition string) bool {
	lenition = strings.Replace(lenition, "->", "→", 1)

	for _, l := range e.Lenitions {
		if l == lenition {
			return true
		}
	}

	return false
}

func (e *DictionaryEntry) ToFilter() WordFilter {
	wf := append(make(WordFilter, 0, 2), e.ID, e.PoS)

	if len(e.Prefixes) > 0 {
		wf = append(wf, strings.Join(e.Prefixes, "-")+"-")
	}
	if len(e.Infixes) > 0 {
		wf = append(wf, "<"+strings.Join(e.Infixes, " ")+">")
	}
	if len(e.Suffixes) > 0 {
		wf = append(wf, "-"+strings.Join(e.Suffixes, "-"))
	}
	if len(e.Prefixes) == 0 && len(e.Infixes) == 0 && len(e.Suffixes) == 0 {
		wf = append(wf, "noaffix")
	}
	if len(e.Lenitions) > 0 {
		wf = append(wf, e.Lenitions...)
	} else {
		wf = append(wf, "nolen")
	}

	return wf
}

var suffixAliases = map[string]string{
	"yä":  "y",
	"ä":   "y",
	"ru":  "r",
	"ur":  "r",
	"ti":  "t",
	"it":  "t",
	"ìri": "ri",
	"ìl":  "l",
}

var infixAliases = map[string]string{
	"iyev": "ìyev",
	"eiy":  "ei",
}

func (e *DictionaryEntry) Copy() DictionaryEntry {
	e2 := *e

	e2.Definitions = make(map[string]string, len(e.Definitions))
	for k, v := range e.Definitions {
		e2.Definitions[k] = v
	}
	e2.Prefixes = append(e.Prefixes[:0:0], e.Prefixes...)
	e2.Infixes = append(e.Infixes[:0:0], e.Infixes...)
	e2.Suffixes = append(e.Suffixes[:0:0], e.Suffixes...)
	e2.Lenitions = append(e.Lenitions[:0:0], e.Lenitions...)
	e2.Comment = append(e.Comment[:0:0], e.Comment...)

	return e2
}

// CombinedDictionary will call the interface methods on all referenced dictionaries. Entry will
// take the earliest answer, while Lookup will combine them.
type CombinedDictionary []Dictionary

func (c CombinedDictionary) Entry(ctx context.Context, id string) (*DictionaryEntry, error) {
	for _, dict := range c {
		res, err := dict.Entry(ctx, id)
		if errors.Is(err, ErrDictionaryEntryNotFound) {
			continue
		} else if err != nil {
			return nil, err
		}

		return res, nil
	}

	return nil, ErrDictionaryEntryNotFound
}

func (c CombinedDictionary) Lookup(ctx context.Context, search string) ([]DictionaryEntry, error) {
	allRes := make([]DictionaryEntry, 0, 16)
	for _, dict := range c {
		res, err := dict.Lookup(ctx, search)
		if err != nil && !errors.Is(err, ErrDictionaryEntryNotFound) {
			return nil, err
		}

		allRes = append(allRes, res...)
	}

	return allRes, nil
}
