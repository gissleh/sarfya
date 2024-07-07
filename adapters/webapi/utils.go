package webapi

import (
	"github.com/gissleh/sarfya"
	"github.com/labstack/echo/v4"
	"net/http"
	"net/url"
	"strings"
)

func Utils(group *echo.Group, dict sarfya.Dictionary) {
	group.GET("/lookup/:search", func(c echo.Context) error {
		escapedTerms, err := url.QueryUnescape(c.Param("search"))
		if err != nil {
			return err
		}

		terms := strings.Split(escapedTerms, ",")
		res := make([][]sarfya.DictionaryEntry, len(terms))

		for i, term := range terms {
			entries, err := dict.Lookup(c.Request().Context(), term)
			if err != nil {
				return err
			}

			res[i] = entries
		}

		return c.JSON(200, map[string]any{
			"entries": res,
		})
	})

	group.POST("/parse-sentence", func(c echo.Context) error {
		var input struct {
			Text   string `json:"text"`
			Lookup bool   `json:"lookup"`
		}
		if err := c.Bind(&input); err != nil {
			return err
		}
		if input.Text == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "text is required (did you forget content-type?)",
			})
		}

		sentence := sarfya.ParseSentence(input.Text)

		var entries map[int][]entryWithFilter
		wordMap := sentence.WordMap()
		if input.Lookup {
			entries = make(map[int][]entryWithFilter, 16)
			for id, search := range wordMap {
				res, err := dict.Lookup(c.Request().Context(), search)
				if err != nil {
					return err
				}

				resWithFilter := make([]entryWithFilter, len(res))
				for i, entry := range res {
					resWithFilter[i] = entryWithFilter{
						DictionaryEntry: entry,
						Filter:          entry.ToFilter().String(),
					}
				}

				entries[id] = resWithFilter
			}
		}

		return c.JSON(200, struct {
			Sentence  sarfya.Sentence           `json:"parsed"`
			RawText   string                    `json:"rawText"`
			InputText string                    `json:"inputText"`
			WordMap   map[int]string            `json:"wordMap"`
			Entries   map[int][]entryWithFilter `json:"entries"`
		}{
			Sentence:  sentence,
			RawText:   sentence.RawText(),
			InputText: sentence.String(),
			WordMap:   wordMap,
			Entries:   entries,
		})
	})
}

type entryWithFilter struct {
	sarfya.DictionaryEntry
	Filter string `json:"filter"`
}
