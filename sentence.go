package sarfya

import (
	"fmt"
	"io"
	"strings"
	"unicode"
)

const syntaxSet = ".,;:–—!? ()[]{}/0123456789\n"

var paraReplacer = strings.NewReplacer("(", "", ")", "", "{", "", "}", "", "[", "", "]", "")
var quoteReplacer = strings.NewReplacer("‘", "'", "’", "'", "ʼ", "'")
var closures = [][2]string{
	{"(", ")"},
	{"{", "}"},
}

func ParseSentence(raw string) Sentence {
	parts := make([]SentencePart, 0, len(raw)/4)
	curr := raw

parseLoop:
	for len(curr) > 0 {
		currIDs := make([]int, 0)
		currID := 0

		isAlt := false
		if curr[0] == '/' {
			isAlt = true
			curr = curr[1:]
		}

		isNewline := false
		if len(curr) > 0 && curr[0] == '\n' {
			isNewline = true
			curr = curr[1:]
		}

		for len(curr) > 0 {
			if curr[0] >= '0' && curr[0] <= '9' {
				currID = (currID * 10) + int(curr[0]-'0')
				curr = curr[1:]
			} else if curr[0] == '+' {
				currIDs = append(currIDs, currID)
				currID = 0
				curr = curr[1:]
			} else {
				break
			}
		}

		if len(curr) > 0 {
			for _, closure := range closures {
				opener := closure[0]
				closer := closure[1]

				if strings.HasPrefix(curr, opener) {
					endIndex := strings.Index(curr, closer)
					if endIndex == -1 {
						endIndex = len(curr)
					}

					if currID != 0 {
						parts = append(parts, SentencePart{
							IDs:     append(currIDs, currID),
							Text:    curr[len(opener):endIndex],
							Alt:     isAlt,
							Newline: isNewline,
						})

						currID = 0
						currIDs = currIDs[:0]
					} else {
						parts = append(parts, SentencePart{
							Text:    curr[len(opener):endIndex],
							Alt:     isAlt,
							Newline: isNewline,
						})
					}

					if len(curr) != endIndex {
						curr = curr[endIndex+len(closer):]
					} else {
						curr = curr[endIndex:]
					}

					continue parseLoop
				}
			}
		}

		canAppendNonLinked := !isNewline &&
			len(parts) > 0 &&
			parts[len(parts)-1].HiddenText == "" &&
			len(parts[len(parts)-1].IDs) == 0 &&
			!strings.ContainsAny(parts[len(parts)-1].Text, "()[]{}")

		punctuationIndex := strings.IndexAny(curr, syntaxSet)
		if punctuationIndex == 0 {
			if canAppendNonLinked {
				parts[len(parts)-1].Text += curr[:1]
			} else {
				parts = append(parts, SentencePart{
					Text:    curr[:1],
					Alt:     isAlt,
					Newline: isNewline,
				})
			}

			curr = curr[1:]
			continue
		}

		if punctuationIndex == -1 {
			punctuationIndex = len(curr)
		}

		if currID != 0 {
			parts = append(parts, SentencePart{
				IDs:     append(currIDs, currID),
				Text:    curr[0:punctuationIndex],
				Alt:     isAlt,
				Newline: isNewline,
			})
		} else if canAppendNonLinked {
			parts[len(parts)-1].Text += curr[0:punctuationIndex]
		} else {
			parts = append(parts, SentencePart{
				Text:    curr[0:punctuationIndex],
				Newline: isNewline,
			})
		}

		curr = curr[punctuationIndex:]
	}

	for i, part := range parts {
		parts[i].Text = quoteReplacer.Replace(part.Text)

		if strings.Contains(part.Text, "|") {
			split := strings.SplitN(part.Text, "|", 2)
			parts[i].HiddenText = split[0]
			parts[i].Text = split[1]
		}

		if text, found := strings.CutSuffix(part.Text, "-"); found && len(part.IDs) > 0 && len(text) != 0 {
			parts[i].Text = text
			parts[i].Prepend = true
		}
	}

	return parts
}

type Sentence []SentencePart

func (s Sentence) String() string {
	sb := strings.Builder{}
	for i, part := range s {
		if part.Newline {
			sb.WriteByte('\n')
		}
		if part.Alt {
			sb.WriteByte('/')
		}

		for i, id := range part.IDs {
			if i > 0 {
				sb.WriteByte('+')
			}
			sb.WriteString(fmt.Sprint(id))
		}
		if strings.ContainsAny(part.Text, "()[]") {
			sb.WriteByte('{')
			if part.HiddenText != "" {
				sb.WriteString(part.HiddenText)
				sb.WriteByte('|')
			}
			sb.WriteString(part.Text)
			if part.Prepend {
				sb.WriteByte('-')
			}
			sb.WriteByte('}')
		} else if part.HiddenText != "" || strings.ContainsAny(part.Text, "()[]{}0123456789/") || ((s.collidesWith(i-1) || s.collidesWith(i+1) || s.isDash(i+1) || strings.ContainsAny(part.Text, syntaxSet)) && s.collidesWith(i)) {
			sb.WriteByte('(')
			if part.HiddenText != "" {
				sb.WriteString(part.HiddenText)
				sb.WriteByte('|')
			}
			sb.WriteString(part.Text)
			if part.Prepend {
				sb.WriteByte('-')
			}
			sb.WriteByte(')')
		} else {
			sb.WriteString(part.Text)
			if part.Prepend {
				sb.WriteByte('-')
			}
		}
	}

	return sb.String()
}

func (s Sentence) RawText() string {
	res := strings.Builder{}
	res.Grow(64)
	for _, part := range s {
		_ = part.WriteRawTo(&res)
	}

	return res.String()
}

func (s Sentence) WordMap() map[int]string {
	noSpaceMap := make(map[int]bool)
	res := make(map[int]string, len(s))

	for i, part := range s {
		if len(part.IDs) == 0 {
			continue
		}

		for _, id := range part.IDs {
			text := part.Text
			if part.HiddenText != "" {
				text = part.HiddenText
			}

			text = paraReplacer.Replace(text)

			if res[id] == "" {
				res[id] = text
			} else if noSpaceMap[id] {
				if part.Prepend {
					res[id] = text + res[id]
				} else {
					res[id] = res[id] + text
				}
				noSpaceMap[id] = false
			} else {
				if part.Prepend {
					res[id] = text + " " + res[id]
				} else {
					res[id] = res[id] + " " + text
				}
			}

			noSpace := i < len(s)-1 && (len(s[i+1].IDs) > 0 || s[i+1].Text == "-")
			if noSpace {
				noSpaceMap[id] = true
			}
		}
	}

	return res
}

func (s Sentence) HasPartID(id int) bool {
	for _, part := range s {
		for _, partID := range part.IDs {
			if id == partID {
				return true
			}
		}
	}

	return false
}

func (s Sentence) SearchRaw(query string) [][]int {
	var indicesStack [64]int
	indices := indicesStack[:0]

	querySb := strings.Builder{}
	for _, ch := range query {
		if unicode.IsLetter(ch) || ch == ' ' {
			querySb.WriteRune(unicode.ToLower(ch))
		}
	}
	query = querySb.String()

	sb := strings.Builder{}
	sb.WriteRune(' ')
	for _, part := range s {
		indices = append(indices, sb.Len())
		for _, ch := range part.Text {
			if unicode.IsLetter(ch) || ch == ' ' {
				sb.WriteRune(unicode.ToLower(ch))
			}
		}
	}
	sb.WriteRune(' ')

	text := sb.String()
	pos := 0
	res := make([][]int, 0)

	for pos < len(text) {
		relIndex := strings.Index(text[pos:], query)
		if relIndex == -1 {
			break
		}

		index := relIndex + pos
		pos = index + 1

		startIndex := -1
		endIndex := -1
		for i, partIndex := range indices {
			nextIndex := len(text)
			if i < len(indices)-1 {
				nextIndex = indices[i+1]
			}

			if partIndex >= index || nextIndex > index {
				if partIndex >= index+len(query) {
					break
				}

				if startIndex == -1 {
					startIndex = i
				}
				endIndex = i
			}
		}

		if startIndex != -1 && endIndex != -1 {
			match := make([]int, 0, endIndex-startIndex+1)
			for i := startIndex; i <= endIndex; i++ {
				match = append(match, i)
			}

			res = append(res, match)
		}
	}

	return res
}

func (s Sentence) isDash(index int) bool {
	return index >= 0 && index < len(s) && s[index].Text == "-"
}

func (s Sentence) collidesWith(index int) bool {
	return index >= 0 && index < len(s) && len(s[index].IDs) > 0 && !s[index].Newline && !s[index].Alt
}

func (s Sentence) WithoutAlts(spans [][]int) Sentence {
	res := make(Sentence, 0, len(s))
	for i, part := range s {
		part := part
		if !part.Alt {
			res = append(res, part)
			continue
		}

		found := false
	CheckIDLoop:
		for _, span := range spans {
			for _, index := range span {
				if index == i {
					found = true
					break CheckIDLoop
				}
			}
		}

		if found {
			part.Alt = false
			res[len(res)-1] = part
		}

		for _, span := range spans {
			for k := range span {
				if span[k] >= i {
					span[k] -= 1
				}
			}
		}
	}

	return res
}

func (s Sentence) NextLinked(index int) int {
	for i, part := range s[index+1:] {
		if len(part.IDs) > 0 {
			return index + i + 1
		}
	}

	return -1
}

func (s Sentence) PrevLinked(index int) int {
	for i := index - 1; i >= 0; i-- {
		part := s[i]
		if len(part.IDs) > 0 {
			return i
		}
	}

	return -1
}

type SentencePart struct {
	IDs        []int  `json:"ids,omitempty" yaml:"ids,omitempty"`
	Text       string `json:"text" yaml:"text"`
	HiddenText string `json:"hiddenText,omitempty" yaml:"hidden_text,omitempty"`
	Alt        bool   `json:"alt,omitempty" yaml:"alt,omitempty"`
	Newline    bool   `json:"newline,omitempty" yaml:"newline,omitempty"`
	Prepend    bool   `json:"prepend,omitempty" yaml:"prepend,omitempty"`
}

func (p *SentencePart) HasAnyID(ids []int) bool {
	for _, id := range ids {
		if p.HasID(id) {
			return true
		}
	}

	return false
}

func (p *SentencePart) HasID(id int) bool {
	for _, pID := range p.IDs {
		if pID == id {
			return true
		}
	}

	return false
}

func (p *SentencePart) RawText() string {
	sb := strings.Builder{}
	sb.Grow(len(p.Text) + 8)
	_ = p.WriteRawTo(&sb)

	return sb.String()
}

func (p *SentencePart) WriteRawTo(w io.StringWriter) error {
	if p.Alt {
		return nil
	}
	if p.Newline {
		_, err := w.WriteString("\n")
		if err != nil {
			return err
		}
	}
	_, err := w.WriteString(p.Text)
	if err != nil {
		return err
	}

	return nil
}
