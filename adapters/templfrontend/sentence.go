package templfrontend

import (
	"bufio"
	"context"
	"fmt"
	"github.com/a-h/templ"
	"github.com/gissleh/sarfya"
	"io"
	"strconv"
	"strings"
)

// sentence is a templ code that generates <a>s and <span>s for the sentence parts.
// This cannot be a normal .templ component since templ inserts spaces for every if-statement.
func sentence(id string, language string, spans [][]int, adjacentSpans [][]int, wordMap map[int][]sarfya.DictionaryEntry, sentence sarfya.Sentence) templ.Component {
	id = templ.EscapeString(id)
	language = templ.EscapeString(language)
	className := "sentence lang-" + language
	inSpan := make(map[int]bool)
	inAdjacent := make(map[int]bool)

	for _, span := range spans {
		for _, index := range span {
			inSpan[index] = true
		}
	}
	for _, span := range adjacentSpans {
		for _, index := range span {
			inAdjacent[index] = true
		}
	}

	idList := func(ids []int) string {
		sb := strings.Builder{}
		sb.Grow(16)
		sb.WriteByte('[')
		for i, id := range ids {
			if i > 0 {
				sb.WriteByte(',')
			}
			sb.WriteString(strconv.Itoa(id))
		}
		sb.WriteByte(']')

		return sb.String()
	}

	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		bw := bufio.NewWriterSize(w, 128)
		_, _ = bw.WriteString("<div class=\"")
		_, _ = bw.WriteString(className)
		_, _ = bw.WriteString("\" data-id=\"")
		_, _ = bw.WriteString(id)
		_, _ = bw.WriteString("\" data-lang=\"")
		_, _ = bw.WriteString(language)
		_, _ = bw.WriteString("\">")

		for i, part := range sentence {
			if part.Newline {
				_, _ = bw.WriteString("<br/>")
				part.Newline = false
			}

			if len(part.IDs) > 0 && wordMap[part.IDs[0]] != nil {
				_, _ = bw.WriteString("<a href=\"")
				_, _ = bw.WriteString(string(templ.URL(fmt.Sprintf("/search/%s:%s", wordMap[part.IDs[0]][0].Word, wordMap[part.IDs[0]][0].ID))))
				if inSpan[i] {
					_, _ = bw.WriteString("\" class=\"selected")
				} else if inAdjacent[i] {
					_, _ = bw.WriteString("\" class=\"adjacent")
				}
				_, _ = bw.WriteString("\" data-ids=\"")
				_, _ = bw.WriteString(idList(part.IDs))
				_, _ = bw.WriteString("\">")
				_, _ = bw.WriteString(templ.EscapeString(part.RawText()))
				_, _ = bw.WriteString("</a>")
			} else {
				wrapInSpan := len(part.IDs) > 0 || inSpan[i] || inAdjacent[i]
				if wrapInSpan {
					_, _ = bw.WriteString("<span")
					if inSpan[i] {
						_, _ = bw.WriteString(" class=\"selected\"")
					} else if inAdjacent[i] {
						_, _ = bw.WriteString(" class=\"adjacent\"")
					}
					if len(part.IDs) > 0 {
						_, _ = bw.WriteString(" data-ids=\"")
						_, _ = bw.WriteString(idList(part.IDs))
						_, _ = bw.WriteString("\"")
					}
					_, _ = bw.WriteString(">")
				}

				_, _ = bw.WriteString(templ.EscapeString(part.RawText()))

				if wrapInSpan {
					_, _ = bw.WriteString("</span>")
				}
			}
		}

		_, _ = bw.WriteString("</div>")
		return bw.Flush()
	})
}
