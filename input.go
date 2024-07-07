package sarfya

import "strings"

type Input struct {
	ID           string            `json:"id,omitempty" yaml:"id"`
	Text         string            `json:"text" yaml:"text"`
	LookupFilter map[int]string    `json:"lookupFilter,omitempty" yaml:"lookup_filter,omitempty"`
	Translations map[string]string `json:"translations" yaml:"translations"`
	Source       Source            `json:"source" yaml:"source,omitempty"`
	Annotations  []Annotation      `json:"annotations" yaml:"annotations,omitempty"`
	Flags        []ExampleFlag     `json:"flags,omitempty" yaml:"flags,omitempty"`
}

type InputLookupConstraints struct {
	ID        *string  `json:"id,omitempty" yaml:"id,omitempty"`
	PoS       *string  `json:"pos,omitempty" yaml:"pos,omitempty"`
	Prefixes  []string `json:"prefixes,omitempty" yaml:"prefixes,omitempty"`
	Infixes   []string `json:"infixes,omitempty" yaml:"infixes,omitempty"`
	Suffixes  []string `json:"suffixes,omitempty" yaml:"suffixes,omitempty"`
	Lenitions []string `json:"lenitions,omitempty" yaml:"lenitions,omitempty"`
}

func (ilc *InputLookupConstraints) ToFilter() WordFilter {
	wf := make(WordFilter, 0, 6)
	if ilc.ID != nil {
		wf = append(wf, *ilc.ID)
	}
	if ilc.PoS != nil {
		wf = append(wf, *ilc.PoS)
	}
	if len(ilc.Prefixes) > 0 {
		wf = append(wf, strings.Join(ilc.Prefixes, "-")+"-")
	}
	if len(ilc.Infixes) > 0 {
		wf = append(wf, "<"+strings.Join(ilc.Infixes, " ")+">")
	}
	if len(ilc.Suffixes) > 0 {
		wf = append(wf, "-"+strings.Join(ilc.Suffixes, "-"))
	}
	if len(ilc.Lenitions) > 0 {
		wf = append(wf, ilc.Lenitions...)
	}

	return wf
}
