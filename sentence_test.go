package sarfya

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

var sentenceTestTable = []struct {
	Raw     string
	Res     Sentence
	WordMap map[int]string
	RawText string
}{
	{"1oel 2ngati 3kameie.", Sentence{
		{IDs: []int{1}, Text: "oel"},
		{Text: " "},
		{IDs: []int{2}, Text: "ngati"},
		{Text: " "},
		{IDs: []int{3}, Text: "kameie"},
		{Text: ".", SentenceBoundary: true},
	}, map[int]string{
		1: "oel",
		2: "ngati",
		3: "kameie",
	}, "oel ngati kameie."},
	{"143fìkem 9118ìlä 648281fya'o.", Sentence{
		{IDs: []int{143}, Text: "fìkem"},
		{Text: " "},
		{IDs: []int{9118}, Text: "ìlä"},
		{Text: " "},
		{IDs: []int{648281}, Text: "fya'o"},
		{Text: ".", SentenceBoundary: true},
	}, map[int]string{
		143:    "fìkem",
		9118:   "ìlä",
		648281: "fya'o",
	}, "fìkem ìlä fya'o."},
	{"1(Nari si), 2ma 3'eylan", Sentence{
		{IDs: []int{1}, Text: "Nari si"},
		{Text: ", "},
		{IDs: []int{2}, Text: "ma"},
		{Text: " "},
		{IDs: []int{3}, Text: "'eylan"},
	}, map[int]string{
		1: "Nari si",
		2: "ma",
		3: "'eylan",
	}, "Nari si, ma 'eylan"},
	{"1+2Meholpxay", Sentence{
		{IDs: []int{1, 2}, Text: "Meholpxay"},
	}, map[int]string{
		1: "Meholpxay",
		2: "Meholpxay",
	}, "Meholpxay"},
	{"1Tsakem 2rä'ä 1si!", Sentence{
		{IDs: []int{1}, Text: "Tsakem"},
		{Text: " "},
		{IDs: []int{2}, Text: "rä'ä"},
		{Text: " "},
		{IDs: []int{1}, Text: "si"},
		{Text: "!", SentenceBoundary: true},
	}, map[int]string{
		1: "Tsakem si",
		2: "rä'ä",
	}, "Tsakem rä'ä si!"},
	{"1oel 2ngati 3(kam)3+4(ei)3(e).", Sentence{
		{IDs: []int{1}, Text: "oel"},
		{Text: " "},
		{IDs: []int{2}, Text: "ngati"},
		{Text: " "},
		{IDs: []int{3}, Text: "kam"},
		{IDs: []int{3, 4}, Text: "ei"},
		{IDs: []int{3}, Text: "e"},
		{Text: ".", SentenceBoundary: true},
	}, map[int]string{
		1: "oel",
		2: "ngati",
		3: "kameie",
		4: "ei",
	}, "oel ngati kameie."},
	{"1oel 2tsole'a 3(3a) 4'uot.", Sentence{
		{IDs: []int{1}, Text: "oel"},
		{Text: " "},
		{IDs: []int{2}, Text: "tsole'a"},
		{Text: " "},
		{IDs: []int{3}, Text: "3a"},
		{Text: " "},
		{IDs: []int{4}, Text: "'uot"},
		{Text: ".", SentenceBoundary: true},
	}, map[int]string{
		1: "oel",
		2: "tsole'a",
		3: "3a",
		4: "'uot",
	}, "oel tsole'a 3a 'uot."},
	{"{(Pxelì'u mì mekemyo)}", Sentence{
		{Text: "(Pxelì'u mì mekemyo)"},
	}, map[int]string{}, "(Pxelì'u mì mekemyo)"},
	{"{(}1Nìn: 2Mekemyo 3a 4le'awtu{)}", Sentence{
		{Text: "("},
		{IDs: []int{1}, Text: "Nìn"},
		{Text: ": "},
		{IDs: []int{2}, Text: "Mekemyo"},
		{Text: " "},
		{IDs: []int{3}, Text: "a"},
		{Text: " "},
		{IDs: []int{4}, Text: "le'awtu"},
		{Text: ")"},
	}, map[int]string{
		1: "Nìn",
		2: "Mekemyo",
		3: "a",
		4: "le'awtu",
	}, "(Nìn: Mekemyo a le'awtu)"},
	{"1uvan 2a 3oe 4(uvan soli|soli) 5lu 6'o'.", Sentence{
		{IDs: []int{1}, Text: "uvan"},
		{Text: " "},
		{IDs: []int{2}, Text: "a"},
		{Text: " "},
		{IDs: []int{3}, Text: "oe"},
		{Text: " "},
		{IDs: []int{4}, Text: "soli", HiddenText: "uvan soli"},
		{Text: " "},
		{IDs: []int{5}, Text: "lu"},
		{Text: " "},
		{IDs: []int{6}, Text: "'o'"},
		{Text: ".", SentenceBoundary: true},
	}, map[int]string{
		1: "uvan",
		2: "a",
		3: "oe",
		4: "uvan soli",
		5: "lu",
		6: "'o'",
	}, "uvan a oe soli lu 'o'."},
	{"1Fo 2pähem 3pesrrpxì/4trrpxìpe.", Sentence{
		{IDs: []int{1}, Text: "Fo"},
		{Text: " "},
		{IDs: []int{2}, Text: "pähem"},
		{Text: " "},
		{IDs: []int{3}, Text: "pesrrpxì"},
		{IDs: []int{4}, Text: "trrpxìpe", Alt: true},
		{Text: ".", SentenceBoundary: true},
	}, map[int]string{
		1: "Fo",
		2: "pähem",
		3: "pesrrpxì",
		4: "trrpxìpe",
	}, "Fo pähem pesrrpxì."},
	{"1Kìng 2a'awve\n3Kìng 4amuve", Sentence{
		{IDs: []int{1}, Text: "Kìng"},
		{Text: " "},
		{IDs: []int{2}, Text: "a'awve"},
		{IDs: []int{3}, Text: "Kìng", Newline: true},
		{Text: " "},
		{IDs: []int{4}, Text: "amuve"},
	}, map[int]string{
		1: "Kìng",
		2: "a'awve",
		3: "Kìng",
		4: "amuve",
	}, "Kìng a'awve\nKìng amuve"},
	{"1+4(Tìomum)1+2+4(mì) 3+4oeyä.", Sentence{
		{IDs: []int{1, 4}, Text: "Tìomum"},
		{IDs: []int{1, 2, 4}, Text: "mì"},
		{Text: " "},
		{IDs: []int{3, 4}, Text: "oeyä"},
		{Text: ".", SentenceBoundary: true},
	}, map[int]string{
		1: "Tìomummì",
		2: "mì",
		3: "oeyä",
		4: "Tìomummì oeyä",
	}, "Tìomummì oeyä."},
	{"1(Ean)-2(na)-3(ta'leng)-1a 4tute.", Sentence{
		{Text: "Ean", IDs: []int{1}},
		{Text: "-"},
		{Text: "na", IDs: []int{2}},
		{Text: "-"},
		{Text: "ta'leng", IDs: []int{3}},
		{Text: "-"},
		{Text: "a", IDs: []int{1}},
		{Text: " "},
		{Text: "tute", IDs: []int{4}},
		{Text: ".", SentenceBoundary: true},
	}, map[int]string{
		1: "Eana",
		2: "na",
		3: "ta'leng",
		4: "tute",
	}, "Ean-na-ta'leng-a tute."},
	{"1Fìtìmungwrr 2horenä 3seiyi 4oe 3irayo-!", Sentence{
		{Text: "Fìtìmungwrr", IDs: []int{1}},
		{Text: " "},
		{Text: "horenä", IDs: []int{2}},
		{Text: " "},
		{Text: "seiyi", IDs: []int{3}},
		{Text: " "},
		{Text: "oe", IDs: []int{4}},
		{Text: " "},
		{Text: "irayo", IDs: []int{3}, Prepend: true},
		{Text: "!", SentenceBoundary: true},
	}, map[int]string{
		1: "Fìtìmungwrr",
		2: "horenä",
		3: "irayo seiyi",
		4: "oe",
	}, "Fìtìmungwrr horenä seiyi oe irayo!"},
	{"{(}Rä'ä syar!{)}", Sentence{
		{Text: "("},
		{Text: "Rä'ä syar!", SentenceBoundary: true},
		{Text: ")"},
	}, map[int]string{}, "(Rä'ä syar!)"},
	{"1yeyfya 4akawnärìp (/ )2mì 3mekemyo 5akoum", Sentence{
		{Text: "yeyfya", IDs: []int{1}},
		{Text: " "},
		{Text: "akawnärìp", IDs: []int{4}},
		{Text: " "},
		{Text: "/ "},
		{Text: "mì", IDs: []int{2}},
		{Text: " "},
		{Text: "mekemyo", IDs: []int{3}},
		{Text: " "},
		{Text: "akoum", IDs: []int{5}},
	}, map[int]string{
		1: "yeyfya",
		2: "mì",
		3: "mekemyo",
		4: "akawnärìp",
		5: "akoum",
	}, "yeyfya akawnärìp / mì mekemyo akoum"},
	{"(1. Holpxay mì mekemyo)", Sentence{
		{Text: "1. Holpxay mì mekemyo", SentenceBoundary: true},
	}, map[int]string{}, "1. Holpxay mì mekemyo"},
	{"", Sentence{}, map[int]string{}, ""},
	{"1", Sentence{
		{IDs: []int{1}, Text: ""},
	}, map[int]string{1: ""}, ""},
	{"", Sentence{}, map[int]string{}, ""},
	// Remember to change `TestSentence_HasPartID` if more are added.
	{"(tìng nari) ma sa'nu, kea holpxay.", Sentence{
		{Text: "tìng nari ma sa'nu, kea holpxay.", SentenceBoundary: true},
	}, map[int]string{}, "tìng nari ma sa'nu, kea holpxay."},
	{"1(Yak soli", Sentence{
		{IDs: []int{1}, Text: "Yak soli"},
	}, map[int]string{
		1: "Yak soli",
	}, "Yak soli"},
	{"{(}Pxelì'u rofa kxemyo", Sentence{
		{Text: "("},
		{Text: "Pxelì'u rofa kxemyo"},
	}, map[int]string{}, "(Pxelì'u rofa kxemyo"},
}

func TestParseSentence(t *testing.T) {
	for _, tt := range sentenceTestTable {
		t.Run(tt.Raw, func(t *testing.T) {
			assert.Equal(t, tt.Res, ParseSentence(tt.Raw))
		})
	}
}

func TestSentence_WordMap(t *testing.T) {
	for _, tt := range sentenceTestTable {
		t.Run(tt.Raw, func(t *testing.T) {
			assert.Equal(t, tt.WordMap, tt.Res.WordMap())
		})
	}
}

func TestSentence_String(t *testing.T) {
	for _, tt := range sentenceTestTable[:len(sentenceTestTable)-3] {
		t.Run(tt.Raw, func(t *testing.T) {
			assert.Equal(t, tt.Raw, tt.Res.String())
		})
	}
}

func TestSentence_RawText(t *testing.T) {
	for _, tt := range sentenceTestTable {
		t.Run(tt.Raw, func(t *testing.T) {
			assert.Equal(t, tt.RawText, tt.Res.RawText())
		})
	}
}

func TestSentence_HasPartID(t *testing.T) {
	sentence := Sentence{
		SentencePart{IDs: []int{1}, Text: "Kam"},
		SentencePart{IDs: []int{1, 2}, Text: "ei"},
		SentencePart{IDs: []int{1}, Text: "e"},
	}

	assert.True(t, sentence.HasPartID(1))
	assert.True(t, sentence.HasPartID(2))
	assert.False(t, sentence.HasPartID(3))
}

func TestSentence_WithoutAlts(t *testing.T) {
	sentence := Sentence{
		{IDs: []int{1}, Text: "Fo"}, {Text: " "},
		{IDs: []int{2}, Text: "pähem"}, {Text: " "},
		{IDs: []int{3}, Text: "pesrrpxì"},
		{IDs: []int{4}, Text: "trrpxìpe", Alt: true}, {Text: "."},
	}
	sentenceA := Sentence{
		{IDs: []int{1}, Text: "Fo"}, {Text: " "},
		{IDs: []int{2}, Text: "pähem"}, {Text: " "},
		{IDs: []int{3}, Text: "pesrrpxì"}, {Text: "."},
	}
	sentenceB := Sentence{
		{IDs: []int{1}, Text: "Fo"}, {Text: " "},
		{IDs: []int{2}, Text: "pähem"}, {Text: " "},
		{IDs: []int{4}, Text: "trrpxìpe"}, {Text: "."},
	}

	assert.Equal(t, sentenceA, sentence.WithoutAlts([][]int{{4}}))
	assert.Equal(t, sentenceB, sentence.WithoutAlts([][]int{{5, 6}}))
}

func TestSentencePart_HasID(t *testing.T) {
	a := SentencePart{IDs: []int{1, 2, 3}}
	b := SentencePart{IDs: []int{3, 4, 5}}
	c := SentencePart{IDs: []int{6, 7, 8}}

	assert.True(t, a.HasID(3))
	assert.True(t, b.HasID(3))
	assert.False(t, c.HasID(3))

	assert.True(t, b.HasAnyID(a.IDs))
	assert.False(t, c.HasAnyID(a.IDs))
	assert.False(t, c.HasAnyID(b.IDs))
}

func TestSentence_SearchRaw(t *testing.T) {
	table := []struct {
		S string
		Q string
		R [][]int
	}{
		{"", "stuff", [][]int{}},
		{"1oel 2ngati 3kameie.", "oe", [][]int{{0}}},
		{"1oel 2ngati 3kameie.", "ngati", [][]int{{2}}},
		{"1oel 2ngati 3kameie.", "kame", [][]int{{4}}},
		{"1oel 2ngati 3kameie.", "ati", [][]int{{2}}},
		{"1oel 2ngati 3kameie.", "ati ka", [][]int{{2, 3, 4}}},
		{"1oel 2ngati 3kameie.", "oel ngati", [][]int{{0, 1, 2}}},
		{"1oel 2ngati 3kameie.", " oel", [][]int{{0}}},
		{"1oel 2ngati 3kameie.", " oel ", [][]int{{0, 1}}},
		{"1oel 2ngati 3kameie.", " ngati ", [][]int{{1, 2, 3}}},
		{"1oel 2ngati 3kameie.", " kameie ", [][]int{{3, 4, 5}}},
		{"1oel 2ngati 3kameie.", "kameie ", [][]int{{4, 5}}},
		{"1oel 2ngati 3kameie.", " kame ", [][]int{}},
		{"1hello, 2world!", "hello world", [][]int{{0, 1, 2}}},
		{"1Fìpamrelìri 2oeyä 3lu 4munea 5'ut 6alu 7tsukfwew", "lu", [][]int{{4}, {10}}},
		{"1This 2is 3a 4long 5text {(}6with 7parentheses{)}", "th", [][]int{{0}, {11}, {13}}},
		{"Run {(fìlì'u)}, rutxe!", "fìlì'u", [][]int{{1}}},
	}

	for _, row := range table {
		t.Run(fmt.Sprintf("`%s`,`%s`", row.S, row.Q), func(t *testing.T) {
			s := ParseSentence(row.S)
			assert.Equal(t, row.R, s.SearchRaw(row.Q))
		})
	}
}
