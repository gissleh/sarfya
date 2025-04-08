package sarfya

import (
	"context"
	"fmt"
	"slices"
	"strings"
)

func NewExample(ctx context.Context, input Input, dictionary Dictionary) (*Example, error) {
	res := &Example{
		ID:           input.ID,
		Translations: make(map[string]Sentence, len(input.Translations)),
		Annotations:  nil,
		Source:       input.Source,
		Words:        make(map[int][]DictionaryEntry),
		Flags:        append(input.Flags[:0:0], input.Flags...),
	}

	allowReef := slices.Contains(input.Flags, EFReefDialect)

	for i, flag := range input.Flags {
		if !flag.Valid() {
			return nil, ExampleError{
				Part:    "flags",
				Key:     fmt.Sprint(i),
				Message: fmt.Sprintf("Flag %#+v is not supported.", flag),
			}
		}
	}

	res.Text = ParseSentence(strings.TrimSpace(input.Text))
	for key, translation := range input.Translations {
		translation = strings.TrimSpace(translation)
		if translation == "" {
			continue
		}

		sentence := ParseSentence(translation)

		for _, part := range sentence {
			if len(part.IDs) == 0 {
				continue
			}

			for _, partID := range part.IDs {
				if !res.Text.HasPartID(partID) {
					return nil, ExampleError{
						Part:    "translations",
						Key:     key,
						Message: fmt.Sprintf("ID %d not found in Na'vi text", partID),
						Link:    partID,
					}
				}
			}
		}

		res.Translations[key] = sentence
	}

	for id, word := range res.Text.WordMap() {
		matches, err := dictionary.Lookup(ctx, word, allowReef)
		if err != nil {
			return nil, ExampleError{
				Part:    "text.wordMap",
				Key:     fmt.Sprint(id),
				Message: fmt.Sprintf("Word lookup \"%s\" failed: %s", word, err),
				Words:   matches,
			}
		}

		filter := ParseMultiWordFilter(input.LookupFilter[id])

		ri := 0
		for _, match := range matches {
			if !filter.Check(&match, true) {
				continue
			}

			matches[ri] = match
			ri += 1
		}
		matches = matches[:ri]

		if len(matches) == 0 {
			return nil, ExampleError{
				Part:    "text.wordMap",
				Key:     fmt.Sprint(id),
				Message: fmt.Sprintf("Word \"%s\" has no matches", word),
				Words:   matches,
			}
		}

		res.Words[id] = matches
	}

	for i, annotation := range input.Annotations {
		if !annotation.Validate() {
			return nil, ExampleError{
				Part:    "annotations",
				Key:     fmt.Sprint(i),
				Message: fmt.Sprintf("Annotation of type %s could not be validated.", annotation.Kind),
			}
		}

		for key := range annotation.Links {
			for _, link := range annotation.Links[key] {
				if !res.Text.HasPartID(link) {
					return nil, ExampleError{
						Part:    "annotations",
						Key:     fmt.Sprint(i),
						Message: "Linked ID not found in Na'vi text.",
						Link:    link,
					}
				}
			}
		}

		res.Annotations = append(res.Annotations, annotation.Copy())
	}

	return res, nil
}

type ExampleError struct {
	Part    string            `json:"part"`
	Key     string            `json:"key"`
	Message string            `json:"message"`
	Link    int               `json:"link,omitempty"`
	Words   []DictionaryEntry `json:"words,omitempty"`
}

func (e ExampleError) Error() string {
	suffix := ""
	if e.Words != nil && len(e.Words) > 0 {
		sb := strings.Builder{}
		sb.Grow(len(e.Words) * 32)

		for _, word := range e.Words {
			if sb.Len() > 0 {
				sb.WriteString(", ")
			} else {
				sb.WriteString(": ")
			}

			sb.WriteString(word.Word)
			sb.WriteString(" (")
			sb.WriteString(word.ID)
			sb.WriteString(":")
			sb.WriteString(word.PoS)
			sb.WriteString(")")
		}

		suffix = sb.String()
	}

	return fmt.Sprintf("%s.%s: %s%s", e.Part, e.Key, e.Message, suffix)
}

type Example struct {
	ID           string                    `json:"id" yaml:"id"`
	Text         Sentence                  `json:"text" yaml:"text"`
	Translations map[string]Sentence       `json:"translations" yaml:"translations"`
	Annotations  []Annotation              `json:"annotations" yaml:"annotations"`
	Source       Source                    `json:"source" yaml:"source"`
	Words        map[int][]DictionaryEntry `json:"words" yaml:"words"`
	Flags        []ExampleFlag             `json:"flags,omitempty" json:"flags,omitempty"`
}

// ListBefore can be used to sort a list of examples in this order:
//
// 1. newer ones first
//
// 2. blog before forum
//
// 3. Na'vi text in alphabetical order.
//
// That way, more recent examples will float to the top, and it causes a predictable
// pattern.
//
// Usage: `sort.Slice(list, func(i, j int) bool { return list[i].ListBefore(&list[j]) })`
func (e *Example) ListBefore(another *Example) bool {
	if e.Source.Date == another.Source.Date {
		if e.Source.ID == another.Source.ID {
			if e.Text[0].Text == another.Text[0].Text {
				return e.Text.RawText() < another.Text.RawText()
			}

			return e.Text[0].Text < another.Text[0].Text
		}

		return e.ID < another.ID
	}

	return e.Source.Date > another.Source.Date
}

func (e *Example) Input() Input {
	// It should never use the context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	input, err := e.MinimalInput(ctx, nil)
	if err != nil {
		panic(err)
	}

	return *input
}

func (e *Example) MinimalInput(ctx context.Context, dictionary Dictionary) (*Input, error) {
	input := Input{
		ID:           e.ID,
		Text:         e.Text.String(),
		LookupFilter: make(map[int]string, len(e.Words)),
		Translations: make(map[string]string, len(e.Translations)),
		Source:       e.Source,
		Annotations:  e.Annotations,
		Flags:        append(e.Flags[:0:0], e.Flags...),
	}

	allowReef := slices.Contains(e.Flags, EFReefDialect)

	for i, translation := range e.Translations {
		input.Translations[i] = translation.String()
	}
	wordMap := e.Text.WordMap()
	for i, words := range e.Words {
		var dictWords []DictionaryEntry
		if dictionary != nil {
			res, err := dictionary.Lookup(ctx, wordMap[i], allowReef)
			if err != nil {
				return nil, err
			}

			dictWords = res
		}

		if len(dictWords) != len(words) {
			filterSet := make(MultiWordFilter, 0, len(words))

			for _, word := range words {
				filterSet = append(filterSet, word.ToFilter())
			}

			input.LookupFilter[i] = filterSet.String()
		}
	}

	return &input, nil
}

func (e *Example) Copy() Example {
	e2 := *e
	e2.Text = append(e.Text[:0:0], e.Text...)
	e2.Translations = make(map[string]Sentence)
	for key, translation := range e.Translations {
		e2.Translations[key] = append(translation[:0:0], translation...)
	}
	e2.Annotations = make([]Annotation, 0, len(e.Annotations))
	for _, annotation := range e.Annotations {
		links := make(map[string][]int)
		for key, value := range annotation.Links {
			links[key] = append(value[:0:0], value...)
		}

		e2.Annotations = append(e2.Annotations, Annotation{
			Kind:  annotation.Kind,
			Links: links,
		})
	}
	e2.Words = make(map[int][]DictionaryEntry, len(e2.Words))
	for key, entries := range e.Words {
		e2.Words[key] = make([]DictionaryEntry, len(entries))
		for i := range entries {
			e2.Words[key][i] = entries[i].Copy()
		}
	}

	return e2
}

func (e *Example) HasFlag(flag ExampleFlag) bool {
	for _, existingFlag := range e.Flags {
		if existingFlag == flag {
			return true
		}
	}

	return false
}

func (e *Example) HasWord(id string) bool {
	for _, words := range e.Words {
		for _, word := range words {
			if word.ID == id {
				return true
			}
		}
	}

	return false
}
