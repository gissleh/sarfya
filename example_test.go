package sarfya

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

var validTestInput = Input{
	ID:   "test-0001",
	Text: "1Uvan 2a 3oe 4(uvan soli|soli) 5lu 6'o'.",
	LookupFilter: map[int]string{
		6: "adj.",
		1: "2644",
	},
	Translations: map[string]string{
		"en": "1(The game) 2that 3I 4played 5was 6fun.",
	},
	Source: Source{
		URL:    "",
		Title:  "Test Case",
		Author: "gissleh",
	},
	Annotations: []Annotation{
		{Kind: AKSplitSiVerb, Links: map[string][]int{
			"noun": {1},
			"si":   {4},
		}},
	},
	Flags: []ExampleFlag{EFNonCanon},
}

var wordUvan = DictionaryEntry{ID: "2644", Word: "uvan", PoS: "n.", Definitions: map[string]string{"en": "game"}, Source: "https://wiki.learnnavi.org/index.php?title=Canon#Midsummer_Night.27s_Dream_Vocabulary (2010-01-31)", Prefixes: []string(nil), Infixes: []string(nil), Suffixes: []string(nil), Lenitions: []string(nil), Comment: []string(nil)}
var wordUvanSoli = DictionaryEntry{ID: "2648", Word: "uvan si", PoS: "vin.", Definitions: map[string]string{"en": "play (a game)"}, Source: "https://wiki.learnnavi.org/index.php?title=Canon#Midsummer_Night.27s_Dream_Vocabulary (2010-01-31) | https://forum.learnnavi.org/index.php?msg=204535 (2010-05-06)", Prefixes: []string(nil), Infixes: []string{"ol"}, Suffixes: []string(nil), Lenitions: []string(nil), Comment: []string(nil)}

var dummyDict = testDictionary{
	"'o'":       DictionaryEntry{ID: "6896", Word: "'o'", PoS: "adj.", Definitions: map[string]string{"en": "bringing fun, exciting"}, Source: "https://naviteri.org/2010/09/getting-to-know-you-part-3/ (2010-09-29)", Prefixes: []string(nil), Infixes: []string(nil), Suffixes: []string(nil), Lenitions: []string(nil), Comment: []string(nil)},
	"a":         DictionaryEntry{ID: "120", Word: "a", PoS: "part.", Definitions: map[string]string{"en": "clause-level attributive marker"}, Source: "Activist Survival Guide (2009-11-24)", Prefixes: []string(nil), Infixes: []string(nil), Suffixes: []string(nil), Lenitions: []string(nil), Comment: []string(nil)},
	"oe":        DictionaryEntry{ID: "1380", Word: "oe", PoS: "pn.", Definitions: map[string]string{"en": "I, me"}, Source: "Paul Frommer, PF | Activist Survival Guide (2009-11-24)", Prefixes: []string(nil), Infixes: []string(nil), Suffixes: []string(nil), Lenitions: []string(nil), Comment: []string(nil)},
	"lu":        DictionaryEntry{ID: "1044", Word: "lu", PoS: "vin.", Definitions: map[string]string{"en": "be, am, is, are"}, Source: "Activist Survival Guide (2009-11-24)", Prefixes: []string(nil), Infixes: []string(nil), Suffixes: []string(nil), Lenitions: []string(nil), Comment: []string(nil)},
	"uvan":      wordUvan,
	"uvan soli": wordUvanSoli,
}

func TestProcess(t *testing.T) {
	t.Run("happy_path", func(t *testing.T) {
		res, err := NewExample(context.Background(), validTestInput, dummyDict)
		assert.NoError(t, err)

		assert.Equal(t, &Example{
			ID: validTestInput.ID,
			Words: map[int][]DictionaryEntry{
				1: {dummyDict["uvan"]},
				2: {dummyDict["a"]},
				3: {dummyDict["oe"]},
				4: {dummyDict["uvan soli"]},
				5: {dummyDict["lu"]},
				6: {dummyDict["'o'"]},
			},
			Text: ParseSentence(validTestInput.Text),
			Translations: map[string]Sentence{
				"en": ParseSentence(validTestInput.Translations["en"]),
			},
			Source: Source{
				URL:    "",
				Title:  "Test Case",
				Author: "gissleh",
			},
			Annotations: []Annotation{
				{Kind: AKSplitSiVerb, Links: map[string][]int{
					"noun": {1},
					"si":   {4},
				}},
			},
			Flags: []ExampleFlag{EFNonCanon},
		}, res)
	})
}
