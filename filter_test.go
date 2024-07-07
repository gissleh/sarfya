package sarfya

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var wordPefneuniltìranyuti = DictionaryEntry{ID: "2612", Word: "uniltìranyu", PoS: "n.", Definitions: map[string]string{"en": "dreamwalker"}, Source: "https://en.wikibooks.org/wiki/Na%27vi/Na%27vi%E2%80%93English_dictionary/glottal_series#U (2009-12-21)", Prefixes: []string{"pe", "fne"}, Infixes: []string(nil), Suffixes: []string{"ti"}, Lenitions: []string(nil), Comment: []string(nil)}
var wordFìtseng = DictionaryEntry{ID: "364", Word: "fìtseng", PoS: "adv., n.", Definitions: map[string]string{"en": "here, this place"}, Source: "Activist Survival Guide (2009-11-24)", Prefixes: []string(nil), Infixes: []string(nil), Suffixes: []string(nil), Lenitions: []string(nil), Comment: []string(nil)}
var wordPukoti = DictionaryEntry{ID: "4496", Word: "puk", PoS: "n.", Definitions: map[string]string{"en": "book"}, Source: "https://schott.blogs.nytimes.com/2010/03/10/questions-answered-invented-languages/ (2010-03-10)", Prefixes: []string(nil), Infixes: []string(nil), Suffixes: []string{"o", "ti"}, Lenitions: []string(nil), Comment: []string(nil)}
var wordLawa = DictionaryEntry{ID: "968", Word: "law", PoS: "adj.", Definitions: map[string]string{"en": "clear, certain"}, Source: "Paul Frommer, PF | Activist Survival Guide (2009-11-24)", Prefixes: []string(nil), Infixes: []string(nil), Suffixes: []string{"a"}, Lenitions: []string(nil), Comment: []string(nil)}
var wordAlaw = DictionaryEntry{ID: "968", Word: "law", PoS: "adj.", Definitions: map[string]string{"en": "clear, certain"}, Source: "Paul Frommer, PF | Activist Survival Guide (2009-11-24)", Prefixes: []string{"a"}, Infixes: []string(nil), Suffixes: []string(nil), Lenitions: []string(nil), Comment: []string(nil)}
var wordHrr = DictionaryEntry{ID: "880", Word: "krr", PoS: "n.", Definitions: map[string]string{"en": "time"}, Source: "Paul Frommer, PF | Activist Survival Guide (2009-11-24)", Prefixes: []string(nil), Infixes: []string(nil), Suffixes: []string(nil), Lenitions: []string{"k→h"}, Comment: []string(nil)}
var wordMowarit = DictionaryEntry{ID: "10008", Word: "mowar", PoS: "n.", Definitions: map[string]string{"en": "advice, bit or piece of advice"}, Source: "https://naviteri.org/2014/05/mipa-ayliu-mipa-aysafpil-new-words-new-ideas/ (2014-05-31)", Prefixes: []string(nil), Infixes: []string(nil), Suffixes: []string{"it"}, Lenitions: []string(nil), Comment: []string(nil)}

func TestWordFilter_Check(t *testing.T) {
	table := []struct {
		Label          string
		Filter         string
		Entry          DictionaryEntry
		CheckModifiers bool
		Expected       bool
	}{
		{
			"Verb id and infix matches",
			"2648:<ol>", wordUvanSoli, true,
			true,
		},
		{
			"Modifiers doesn't match, but isn't checked",
			"<asy>:pe-:-ti:ts->s:-a-", wordUvanSoli, false,
			true,
		},
		{
			"Modifiers doesn't match, but isn't checked, but the PoS filter at the end does fail",
			"<asy>:pe-:-ti:ts->s:-a-:adj.", wordUvanSoli, false,
			false,
		},
		{
			"Verb id matches, but infix isn't checked",
			"2648:<asy>", wordUvanSoli, false,
			true,
		},
		{
			"Verb id matches, but infix doesn't match",
			"2648:<ìlm>", wordUvanSoli, true,
			false,
		},
		{
			"Prefix doesn't match",
			"pe-", wordUvan, true,
			false,
		},
		{
			"Both prefixes matches",
			"pe-fne-", wordPefneuniltìranyuti, true,
			true,
		},
		{
			"A suffix matches",
			"-ti", wordPefneuniltìranyuti, true,
			true,
		},
		{
			"One suffix matches",
			"-o", wordPukoti, true,
			true,
		},
		{
			"Both suffixes matches",
			"-o-ti", wordPukoti, true,
			true,
		},
		{
			"Only one suffix matches",
			"-o-ti", wordPefneuniltìranyuti, true,
			false,
		},
		{
			"One prefixes matches, one does not",
			"pxe-fne-", wordPefneuniltìranyuti, true,
			false,
		},
		{
			"One PoS matches",
			"n.", wordFìtseng, true,
			true,
		},
		{
			"Both PoS matches",
			"n., adv.", wordFìtseng, true,
			true,
		},
		{
			"Only one PoS matches, the other two does not",
			"n., adv., v.", wordUvanSoli, true,
			false,
		},
		{
			"Suffixed adjective passes affix check",
			"-a-", wordLawa, true,
			true,
		},
		{
			"Prefixed adjective passes affix check",
			"-a-", wordAlaw, true,
			true,
		},
		{
			"Non-prefixed noun does not pass affix check",
			"-pe-", wordUvan, true,
			false,
		},
		{
			"Lenition check matches",
			"k->h", wordHrr, true,
			true,
		},
		{
			"Lenition check does not match",
			"p->f", wordHrr, true,
			false,
		},
		{
			"nolen matches on uvan",
			"nolen", wordUvan, true,
			true,
		},
		{
			"nolen fails on hrr",
			"nolen", wordHrr, true,
			false,
		},
		{
			"noaffix passes on uvan",
			"noaffix", wordUvan, true,
			true,
		},
		{
			"noaffix definitely fails on pefneuniltìranyuti",
			"noaffix", wordPefneuniltìranyuti, true,
			false,
		},
		{
			"Mowarit matches noun, -it and no lenition",
			"10008:n.:-it:nolen", wordMowarit, true,
			true,
		},
	}

	for _, tt := range table {
		t.Run(tt.Label, func(t *testing.T) {
			filter := ParseWordFilter(tt.Filter)
			assert.Equal(t, tt.Expected, filter.Check(&tt.Entry, tt.CheckModifiers))
		})
	}
}
