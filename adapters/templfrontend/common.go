package templfrontend

import (
	"context"
	"github.com/gissleh/sarfya"
	"strings"
)

var demoCtxKey = &struct{ key string }{key: "demoCtxKey"}
var langCtxKey = &struct{ key string }{key: "langCtxKey"}

func findLanguage(match sarfya.FilterMatch, langs string) string {
	for _, l := range strings.Split(langs, ",") {
		l := strings.TrimSpace(l)

		if match.Translations[l] != nil {
			return l
		}
	}

	return ""
}

func createDemo(dictionary sarfya.Dictionary) *sarfya.FilterMatch {
	demoData, err := sarfya.NewExample(context.Background(), sarfya.Input{
		ID:   "demo",
		Text: "1Kaltxì, 2ulte 3zola'u 4nìprrte' 5fìweptseng6ne! 7(Pamrel si) 8lì'uor 9nefä 10fu 11takuk 12pumit 17a 13mì 14fìpamrel 15fte 16fwivew.",
		LookupFilter: map[int]string{
			11: "vtr.",
		},
		Translations: map[string]string{
			"en": "1Hello, 2and 3+4(welcome) 6to 5(this website)! 7Write 8(a word) 9above 10or 11click 12one 17+13in 14(this text) 15to 16search.",
			"no": "1Hey, 3+4(velkommen) 6til 5(denne nettsiden)! 7Skrive 8(et ord) 9ovenfor 10eller 11(klikk på) 12en 17+13i 14(denne teksten) 15(for å) 16søke.",
		},
		Source: sarfya.Source{},
	}, dictionary)
	if err != nil {
		return nil
	}

	return &sarfya.FilterMatch{
		Example:             *demoData,
		Selections:          []int{},
		Spans:               [][]int{},
		TranslationAdjacent: map[string][][]int{"en": {}},
		TranslationSpans: map[string][][]int{
			"en": {},
		},
	}
}
