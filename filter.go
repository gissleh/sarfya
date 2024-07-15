package sarfya

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

// ParseFilter parses the filter and gives a list of combinations of dictionary entries that this filter could be used with.
func ParseFilter(ctx context.Context, str string, dictionary Dictionary) (*Filter, []map[int]DictionaryEntry, error) {
	filter := &Filter{}

	longestAt := make(map[int]int)
	nextOperator := FTOAnd
	i := 0
	for len(str) > 0 {
		for key := range longestAt {
			delete(longestAt, key)
		}

		operator := nextOperator
		termString := str
		nearestIndex := len(str)
		nextStart := len(str)
		selectedOp := FTOAnd
		for _, alias := range operatorAliases {
			op := alias[1]
			match := alias[0]

			opIndex := strings.Index(str, match)
			if opIndex != -1 && opIndex <= nearestIndex && len(match) > longestAt[opIndex] {
				selectedOp = op
				nearestIndex = opIndex
				longestAt[opIndex] = len(match)
				nextStart = opIndex + len(match)
			}
		}

		termString = str[:nearestIndex]
		nextOperator = selectedOp
		str = str[nextStart:]

		termString = strings.TrimSpace(termString)

		not := strings.HasPrefix(termString, "!")
		if not {
			termString = termString[1:]
		}

		if len(termString) == 0 {
			return nil, nil, FilterParseError{
				Term:    i,
				Code:    "empty_query_term",
				Message: "A filter term cannot be empty.",
			}
		}

		if operator == FTOAnd {
			if strings.HasPrefix(termString, "src:") {
				sourceID := termString[4:]
				filter.SourceID = &sourceID
				continue
			}

			if strings.HasPrefix(termString, "flag:") {
				flag := ExampleFlag(termString[5:])
				if strings.HasPrefix(string(flag), "-") {
					if !flag[1:].Valid() {
						return nil, nil, FilterParseError{
							Term: i, Code: "flag_not_understood",
							Message: "The flag you specified is not found.",
						}
					}
				} else {
					if !flag.Valid() {
						return nil, nil, FilterParseError{
							Term: i, Code: "flag_not_understood",
							Message: "The flag you specified is not found.",
						}
					}
				}

				filter.Flags = append(filter.Flags, flag)
				continue
			}

			if termString == "opt:noadjacent" || termString == "opt:no_adjacent" || termString == "option:no_adjacent" {
				filter.NoAdjacent = true
				continue
			}
		}

		split := strings.SplitN(termString, ":", 10)
		if len(split) == 10 {
			return nil, nil, FilterParseError{
				Term:    i,
				Code:    "too_many_constraints",
				Message: "A filter term cannot have more than 8 constraints.",
			}
		}

		term := FilterTerm{
			Operator:    operator,
			Word:        split[0],
			Constraints: split[1:],
			Not:         not,
		}
		filter.Terms = append(filter.Terms, term)

		i += 1
		if i == 8 && len(str) > 0 {
			return nil, nil, FilterParseError{
				Term:    i,
				Code:    "too_many_terms",
				Message: "A filter cannot have more than 8 terms.",
			}
		}
	}

	maps, err := filter.lookupWords(ctx, dictionary)
	if err != nil {
		return nil, nil, err
	}

	return filter, maps, nil
}

type Filter struct {
	Terms      []FilterTerm  `json:"terms" yaml:"terms"`
	SourceID   *string       `json:"sourceID" yaml:"source_id"`
	Flags      []ExampleFlag `json:"flags" yaml:"flags"`
	NoAdjacent bool          `json:"noAdjacent,omitempty" yaml:"noAdjacent,omitempty"`
}

func (f *Filter) CheckExample(example Example, resolved map[int]DictionaryEntry) *FilterMatch {
	selections := make([]int, 0)
	spans := make([][]int, 0)
	matches := make([][]int, 0, 16)
	temp := make([]int, 0, 4)
	expandableStart := 0
	skipTo := 0

	if f.SourceID != nil && example.Source.ID != *f.SourceID {
		return nil
	}

	for _, flag := range f.Flags {
		if strings.HasPrefix(string(flag), "-") {
			if example.HasFlag(flag[1:]) {
				return nil
			}
		} else {
			if !example.HasFlag(flag) {
				return nil
			}
		}
	}

	for i, term := range f.Terms {
		if skipTo > i {
			continue
		}

		matches = matches[:0]
		failed := false
		entry := resolved[i]

		for id, words := range example.Words {
			for _, word := range words {
				matchesWord := word.ID == entry.ID || term.Word == "*"

				passed := matchesWord && term.Constraints.Check(&word, true)
				if passed == !term.Not {
					temp = temp[:0]

					for j, part := range example.Text {
						if part.HasID(id) {
							temp = append(temp, j)
						}
					}

					matches = append(matches, append(temp[:0:0], temp...))
					selections = append(selections, id)
					break
				}
			}
		}

		sort.Slice(matches, func(i, j int) bool {
			if (len(matches[i]) > 0) != (len(matches[j]) > 0) {
				return len(matches[i]) > 0
			}

			return matches[i][0] < matches[j][0]
		})

		switch term.Operator {
		case FTOOr:
			{
				if skipTo != i {
					skipTo = len(f.Terms)
					break
				}

				expandableStart = len(spans)
				spans = append(spans, matches...)
			}
		case FTOAnd:
			{
				if len(matches) == 0 {
					failed = true
				}

				expandableStart = len(spans)
				spans = append(spans, matches...)
			}
		case FTOFollowedBy, FTONextTo, FTOASurroundedBy:
			{
				foundAny := false

				for j, span := range spans {
					if len(span) == 0 || j < expandableStart {
						continue
					}

					matchedAfter := false
					matchedBefore := false

					nextLinked := example.Text.NextLinked(span[len(span)-1])
					if nextLinked != -1 {
						for _, match := range matches {
							if nextLinked == match[0] {
								spans[j] = append(spans[j], match...)
								foundAny = true
								matchedAfter = true
								break
							}
						}
					}

					if term.Operator != FTOFollowedBy {
						prevLinked := example.Text.PrevLinked(span[0])
						if prevLinked != -1 {
							for _, match := range matches {
								if prevLinked == match[len(match)-1] {
									spans[j] = append(match, spans[j]...)

									foundAny = true
									matchedBefore = true

									break
								}
							}
						}
					}

					if term.Operator == FTOASurroundedBy {
						if !matchedAfter || !matchedBefore {
							spans[j] = spans[j][:0]
						}
					} else {
						if !matchedAfter && !matchedBefore {
							spans[j] = spans[j][:0]
						}
					}
				}

				if !foundAny {
					failed = true
				}
			}
		case FTOBefore:
			{
				foundAny := false

				for j, span := range spans {
					if j < expandableStart {
						continue
					}

					var selected []int
					earliest := len(example.Text)

					for _, match := range matches {
						if span[len(span)-1] < match[0] && match[0] < earliest {
							selected = match
						}
					}

					if selected != nil {
						spans[j] = append(span, selected...)
						foundAny = true
					} else {
						spans[j] = span[:0]
					}
				}

				if !foundAny {
					failed = true
				}
			}
		case FTOSurrounding:
			{
				foundAny := false

				for j, span := range spans {
					if len(span) >= 2 && j >= expandableStart {
						found := false

						temp = temp[:0]
						for k := range span[:len(span)-1] {
							for _, match := range matches {
								if span[k] < match[0] && span[k+1] > match[len(match)-1] {
									temp = append(temp, match...)
									foundAny = true
									found = true
								}
							}

							if len(temp) > 0 {
								spans[j] = append(span[:k+1], append(temp, span[k+1:]...)...)
							}
						}

						if !found {
							spans[j] = span[:0]
						}
					} else {
						spans[j] = span[:0]
					}
				}

				if !foundAny {
					failed = true
				}
			}
		default:
			failed = true
		}

		if failed {
			orOffset := -1
			for j, term2 := range f.Terms[i+1:] {
				if term2.Operator == FTOOr {
					orOffset = j + 1
					break
				}
			}

			if orOffset != -1 {
				skipTo = i + orOffset
				spans = spans[:0]
				selections = selections[:0]
				expandableStart = 0
			} else {
				return nil
			}
		}
	}

	example = example.Copy()
	example.Text = example.Text.WithoutAlts(spans)
	for lang, translation := range example.Translations {
		example.Translations[lang] = translation.WithoutAlts(spans)
	}

	ri := 0
	for _, span := range spans {
		if len(span) > 0 {
			spans[ri] = span
			ri += 1
		}
	}
	spans = spans[:ri]

	ri = 0
SelectionRetainLoop:
	for _, selection := range selections {
		for _, span := range spans {
			for _, index := range span {
				if example.Text[index].HasID(selection) {
					selections[ri] = selection
					ri += 1
					continue SelectionRetainLoop
				}
			}
		}
	}
	selections = selections[:ri]

	if len(selections) == 0 && len(f.Terms) > 0 {
		return nil
	}

	translationAdjacent := make(map[string][][]int, len(example.Translations))
	translationSpans := make(map[string][][]int, len(example.Translations))
	seen := make(map[int]bool, 8)
	ids := make([]int, 0, 8)
	revIDs := make([]int, 0, 8)

	nonAdjacentMap := map[int]bool{}
	if f.NoAdjacent {
		for _, span := range spans {
			for _, i := range span {
				nonAdjacentMap[i] = true
			}
		}
	}

	for lang, translated := range example.Translations {
		translationSpans[lang] = make([][]int, len(spans))
		translationAdjacent[lang] = make([][]int, len(spans))

		isEN := lang == "en" // TODO: Use context to limit translations

		if len(translated) == 0 {
			continue
		}

		for i, span := range spans {
			translationSpans[lang][i] = []int{}
			translationAdjacent[lang][i] = []int{}

			// Find the IDs to select with.
			ids = ids[:0]
			for key := range seen {
				delete(seen, key)
			}
			for _, index := range span {
				part := example.Text[index]
				for _, id := range part.IDs {
					if seen[id] {
						continue
					}
					ids = append(ids, id)
					seen[id] = true
				}
			}

			// Select with the translation and add any unseen IDs for the adjacent list.
			revIDs = revIDs[:0]
			for key := range seen {
				delete(seen, key)
			}
			for j, part := range example.Translations[lang] {
				if part.HasAnyID(ids) {
					for _, id := range part.IDs {
						if seen[id] {
							continue
						}
						revIDs = append(revIDs, id)
						seen[id] = true
					}

					translationSpans[lang][i] = append(translationSpans[lang][i], j)
				}
			}

			for j, part := range example.Text {
				isInSpan := false
				for _, index := range span {
					if j == index {
						isInSpan = true
						break
					}
				}

				if !isInSpan && part.HasAnyID(revIDs) {
					if isEN && f.NoAdjacent && !nonAdjacentMap[j] {
						return nil
					}

					translationAdjacent[lang][i] = append(translationAdjacent[lang][i], j)
				}
			}
		}
	}

	return &FilterMatch{
		Example:             example,
		Selections:          selections,
		Spans:               spans,
		TranslationAdjacent: translationAdjacent,
		TranslationSpans:    translationSpans,
		WordMap:             example.Text.WordMap(),
	}
}

// lookupWords get all combinations of DictionaryEntries that are matched by the general criteria.
// It will not check prefixes, infixes, suffixes and lenitions here. It will return an error if the
// dictionary failed.
func (f *Filter) lookupWords(ctx context.Context, dictionary Dictionary) ([]map[int]DictionaryEntry, error) {
	maps := make([]map[int]DictionaryEntry, 0, len(f.Terms)*2)
	maps = append(maps, map[int]DictionaryEntry{})

	for i, term := range f.Terms {
		if term.Word == "*" {
			continue
		}

		entries, err := dictionary.Lookup(ctx, term.Word)
		if err != nil {
			return nil, err
		}
		filteredEntries := entries[:0]
		for _, entry := range entries {
			if term.Constraints.Check(&entry, false) {
				filteredEntries = append(filteredEntries, entry)
			}
		}

		if len(filteredEntries) == 0 {
			return nil, FilterParseError{
				Term:    i,
				Code:    "no_matched_entries",
				Message: fmt.Sprintf("No dictionary entry matched word or constraints of %+v", term.Word),
			}
		}

		// Add first found entry to all maps
		for _, m := range maps {
			m[i] = filteredEntries[0]
		}

		// For every subsequent entry, create a copy of the maps with them.
		// So that if there's 2 words for entry 0 and 3 for 1, it'll be:
		// 6 maps with all combinations.
		existingMaps := maps[:]
		for _, entry := range filteredEntries[1:] {
			for _, m := range existingMaps {
				m2 := make(map[int]DictionaryEntry)
				for key, value := range m {
					m2[key] = value.Copy()
				}
				m2[i] = entry

				maps = append(maps, m2)
			}
		}
	}

	return maps, nil
}

type FilterTerm struct {
	Operator    string
	Word        string
	Constraints WordFilter
	Not         bool
}

var operatorAliases = [][2]string{
	{FTOSurrounding, FTOSurrounding},
	{FTOASurroundedBy, FTOASurroundedBy},
	{FTOBefore, FTOBefore},
	{FTOFollowedBy, FTOFollowedBy},
	{FTONextTo, FTONextTo},
	{FTOAnd, FTOAnd},
	{FTOOr, FTOOr},
	{"AND", FTOAnd},
	{"OR", FTOOr},
	{"NEXT TO", FTONextTo},
	{"FOLLOWED BY", FTOFollowedBy},
	{"BEFORE", FTOBefore},
	{"SURROUNDED BY", FTOASurroundedBy},
	{"SURROUNDING", FTOSurrounding},
}

const FTOSurrounding = ">+<"
const FTOASurroundedBy = "++"
const FTOBefore = "+>>"
const FTOFollowedBy = "+>"
const FTONextTo = "+"
const FTOAnd = "&&"
const FTOOr = "||"

func ParseWordFilter(str string) WordFilter {
	if str == "" {
		return nil
	}

	return strings.Split(str, ":")
}

type WordFilter []string

func (wf WordFilter) String() string {
	return strings.Join(wf, ":")
}

func (wf WordFilter) Check(e *DictionaryEntry, checkModifiers bool) bool {
	if wf == nil {
		return true
	}

	for _, constraint := range wf {
		ok := false
		modifiers := 0
		alternatives := strings.Split(constraint, "|")
	AlternativeCheckLoop:
		for _, alternative := range alternatives {
			if alternative == "nolen" {
				if len(e.Lenitions) == 0 {
					ok = true
					break
				}
			} else if alternative == "noaffix" {
				if len(e.Prefixes) == 0 && len(e.Infixes) == 0 && len(e.Suffixes) == 0 {
					ok = true
					break
				}
			} else if strings.HasPrefix(alternative, "-") && strings.HasSuffix(alternative, "-") {
				if !checkModifiers {
					modifiers += 1
					break
				}

				affixes := strings.Split(alternative[1:len(alternative)-1], "-")
				for _, affix := range affixes {
					if !e.HasPrefix(affix) && !e.HasSuffix(affix) {
						continue AlternativeCheckLoop
					}
				}

				ok = true
				break

			} else if strings.HasPrefix(alternative, "-") {
				if !checkModifiers {
					modifiers += 1
					break
				}

				suffixes := strings.Split(alternative[1:], "-")
				for _, suffix := range suffixes {
					if !e.HasSuffix(suffix) {
						continue AlternativeCheckLoop
					}
				}

				ok = true
				break
			} else if strings.HasSuffix(alternative, "-") {
				if !checkModifiers {
					modifiers += 1
					break
				}

				prefixes := strings.Split(alternative[:len(alternative)-1], "-")
				for _, prefix := range prefixes {
					if !e.HasPrefix(prefix) {
						continue AlternativeCheckLoop
					}
				}

				ok = true
				break
			} else if strings.HasPrefix(alternative, "<") && strings.HasSuffix(alternative, ">") {
				if !checkModifiers {
					modifiers += 1
					break
				}

				infixes := strings.Split(alternative[1:len(alternative)-1], " ")
				for _, infix := range infixes {
					if !e.HasInfix(infix) {
						continue AlternativeCheckLoop
					}
				}

				ok = true
				break
			} else if strings.Contains(alternative, "->") || strings.ContainsRune(alternative, '→') {
				if !checkModifiers {
					modifiers += 1
					break
				}

				lenitions := strings.Split(alternative, " ")
				for _, lenition := range lenitions {
					if !e.HasLenition(lenition) {
						break AlternativeCheckLoop
					}
				}

				ok = true
				break
			} else if strings.ContainsRune(alternative, '.') {
				for _, apos := range strings.Split(alternative, ",") {
					apos = strings.TrimSpace(apos)

					for _, pos := range strings.Split(e.PoS, ", ") {
						if strings.TrimSpace(pos) == apos {
							ok = true
							break AlternativeCheckLoop
						}
					}
				}
			} else if e.ID == alternative {
				ok = true
				break
			}
		}

		if !ok && (checkModifiers || modifiers != len(alternatives)) {
			return false
		}
	}

	return true
}

type MultiWordFilter []WordFilter

func ParseMultiWordFilter(str string) MultiWordFilter {
	if str == "" {
		return nil
	}

	parts := strings.Split(str, ";")
	res := make(MultiWordFilter, 0, len(parts))
	for _, part := range parts {
		res = append(res, ParseWordFilter(strings.TrimSpace(part)))
	}

	return res
}

func (mwf MultiWordFilter) String() string {
	sb := strings.Builder{}
	for i, filter := range mwf {
		if i > 0 {
			sb.WriteString(";")
		}

		sb.WriteString(filter.String())
	}

	return sb.String()
}

func (mwf MultiWordFilter) Check(e *DictionaryEntry, checkModifiers bool) bool {
	if mwf == nil {
		return true
	}

	for _, filter := range mwf {
		if filter.Check(e, checkModifiers) {
			return true
		}
	}

	return false
}

type FilterParseError struct {
	Term    int    `json:"term"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e FilterParseError) Error() string {
	return fmt.Sprintf("error on term %d in filter: %s", e.Term, e.Message)
}

type FilterMatch struct {
	Example

	Selections          []int              `json:"selections"`
	Spans               [][]int            `json:"spans"`
	TranslationAdjacent map[string][][]int `json:"translatedAdjacent"`
	TranslationSpans    map[string][][]int `json:"translatedSpans"`
	WordMap             map[int]string     `json:"wordMap"`
}
