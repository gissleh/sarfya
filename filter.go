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

		isText := strings.HasPrefix(split[0], "\"")
		if isText {
			split[0] = strings.Trim(split[0], "\"")
		}

		if len(split) > 2 && isText {
			return nil, nil, FilterParseError{
				Term:    i,
				Code:    "text_filter_constraints",
				Message: "A text filter term cannot have constraints.",
			}
		}

		term := FilterTerm{
			Operator:    operator,
			Word:        split[0],
			Constraints: split[1:],
			Not:         not,
			IsText:      isText,
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
	spans := make([][]int, 0)
	matches := make([][]int, 0, 16)
	temp := make([]int, 0, 4)
	seen := make(map[int]bool)
	seen2 := make(map[int]bool)
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

		if term.IsText {
			text := example.Text
			if len(term.Constraints) > 0 {
				text = example.Translations[term.Constraints[0]]
			}
			if text != nil {
				search := text.SearchRaw(term.Word)
				if len(search) > 0 {
					if len(term.Constraints) == 0 {
						for _, v := range search {
							matches = append(matches, v)
						}
					} else {
						ids := make([]int, 0, len(search)+4)
						matchSpan := make([]int, 0, 16)
						for _, searchSpan := range search {
							for _, j := range searchSpan {
								for _, id := range text[j].IDs {
									if seen[id] {
										continue
									}

									seen[id] = true
									ids = append(ids, id)
								}
							}

							for j, part := range example.Text {
								if !seen2[j] && part.HasAnyID(ids) {
									seen2[j] = true
									matchSpan = append(matchSpan, j)
								}
							}

							if len(matchSpan) > 0 {
								matches = append(matches, matchSpan)
							}

							for key := range seen {
								delete(seen, key)
							}
						}
					}
				}
			}
		} else {
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
						break
					}
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

									matchedBefore = true

									break
								}
							}
						}
					}

					if term.Operator == FTOASurroundedBy {
						if !matchedAfter || !matchedBefore {
							spans[j] = spans[j][:0]
						} else {
							foundAny = true
						}
					} else {
						if !matchedAfter && !matchedBefore {
							spans[j] = spans[j][:0]
						} else {
							foundAny = true
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

	if len(spans) == 0 && len(f.Terms) > 0 {
		return nil
	}

	translationAdjacent := make(map[string][][]int, len(example.Translations))
	translationSpans := make(map[string][][]int, len(example.Translations))
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
		if term.Word == "*" || term.IsText {
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

func (f *Filter) NeedFullList() bool {
	answer := true
	for _, term := range f.Terms {
		if term.Operator == FTOOr && answer {
			return answer
		}

		if !term.IsText && term.Word != "*" {
			answer = false
		}
	}

	return answer
}

func (f *Filter) WordLookupStrategy(resolved map[int]DictionaryEntry) [][]DictionaryEntry {
	var curr []DictionaryEntry
	var res [][]DictionaryEntry

	for i, term := range f.Terms {
		if term.Operator == FTOOr {
			res = append(res, curr)
			curr = []DictionaryEntry{}
		}

		curr = append(curr, resolved[i])
	}

	res = append(res, curr)
	return res
}

type FilterTerm struct {
	Operator    string
	Word        string
	Constraints WordFilter
	Not         bool
	IsText      bool
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
		exact := false
		if strings.HasPrefix(constraint, "=") {
			exact = true
			constraint = constraint[1:]
		}

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
			} else if alternative == "nosuffix" {
				if len(e.Suffixes) == 0 {
					ok = true
					break
				}
			} else if alternative == "noinfix" {
				if len(e.Infixes) == 0 {
					ok = true
					break
				}
			} else if alternative == "noprefix" {
				if len(e.Prefixes) == 0 {
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

				if exact {
					for _, prefix := range e.Prefixes {
						if !inStringList(affixes, prefix, nil) {
							continue AlternativeCheckLoop
						}
					}
					for _, suffix := range e.Suffixes {
						if !inStringList(affixes, suffix, nil) {
							continue AlternativeCheckLoop
						}
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

				if exact {
					for _, suffix := range e.Suffixes {
						if !inStringList(suffixes, suffix, suffixAliases) {
							continue AlternativeCheckLoop
						}
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

				if exact {
					for _, prefix := range e.Prefixes {
						if !inStringList(prefixes, prefix, suffixAliases) {
							continue AlternativeCheckLoop
						}
					}
				}

				ok = true
				break
			} else if strings.HasPrefix(alternative, "<") && strings.HasSuffix(alternative, ">") {
				if !checkModifiers {
					modifiers += 1
					break
				}

				if !e.IsVerb() {
					continue AlternativeCheckLoop
				}

				infixes := strings.Split(alternative[1:len(alternative)-1], " ")
				for _, infix := range infixes {
					if infix == "" {
						continue
					}

					if !e.HasInfix(infix) {
						continue AlternativeCheckLoop
					}
				}

				if exact {
					for _, infix := range e.Infixes {
						if !inStringList(infixes, infix, infixAliases) {
							continue AlternativeCheckLoop
						}
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

				if exact {
					for _, lenition := range e.Lenitions {
						if !inStringList(lenitions, lenition, nil) {
							continue AlternativeCheckLoop
						}
					}
				}

				ok = true
				break
			} else if strings.ContainsRune(alternative, '.') {
				searchSplit := strings.Split(alternative, ",")
				posSplit := strings.Split(e.PoS, ",")

				for i := range searchSplit {
					searchSplit[i] = strings.TrimSpace(searchSplit[i])
				}
				for i := range posSplit {
					posSplit[i] = strings.TrimSpace(posSplit[i])
				}

				for _, search := range searchSplit {
					if !inStringList(posSplit, search, nil) {
						continue AlternativeCheckLoop
					}
				}

				if exact {
					for _, pos := range posSplit {
						if !inStringList(searchSplit, pos, nil) {
							continue AlternativeCheckLoop
						}
					}
				}

				ok = true
				break
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

	Spans               [][]int            `json:"spans"`
	TranslationAdjacent map[string][][]int `json:"translatedAdjacent"`
	TranslationSpans    map[string][][]int `json:"translatedSpans"`
	WordMap             map[int]string     `json:"wordMap"`
}

func inStringList(list []string, value string, aliases map[string]string) bool {
	if alias, ok := aliases[value]; ok {
		value = alias
	}
	for _, item := range list {
		if item == value || value == aliases[item] {
			return true
		}
	}

	return false
}
