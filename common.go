package sarfya

type Source struct {
	ID     string `json:"id,omitempty" yaml:"id,omitempty"`
	Date   string `json:"date,omitempty" yaml:"date,omitempty"`
	URL    string `json:"url,omitempty" yaml:"url,omitempty"`
	Title  string `json:"title,omitempty" yaml:"title,omitempty"`
	Author string `json:"author,omitempty" yaml:"author,omitempty"`
}

type ExampleFlag string

func (f ExampleFlag) Valid() bool {
	switch f {
	case EFPoetry, EFNonCanon, EFUserTranslation, EFReefDialect, EFProverb, EFSlang, EFFormal, EFSyntax, EFClipped:
		return true
	default:
		return false
	}
}

const (
	// EFPoetry implies that the sentence may take poetic licenses that may not work in daily speech.
	EFPoetry ExampleFlag = "poetry"
	// EFNonCanon are for examples that did not come from KP or anyone else in charge of the language canon.
	// They may still hold some value for word references.
	EFNonCanon ExampleFlag = "non_canon"
	// EFUserTranslation means that the translation in the source's primary language was not provided
	// by the source.
	EFUserTranslation ExampleFlag = "user_translation"
	// EFReefDialect is for the words of the dopest looking Na'vi. This should be used for when there are
	// rules or spellings applied that are only permitted in the reef dialect.
	EFReefDialect ExampleFlag = "reef_dialect"
	// EFProverb is for proverbial expressions of any kind. The meanings should be in the translations.
	EFProverb ExampleFlag = "proverb"
	// EFSlang is for highly informal language.
	EFSlang ExampleFlag = "slang"
	// EFFormal is for when you gotta henga si
	EFFormal ExampleFlag = "formal"
	// EFSyntax are for patterns that has placeholders.
	EFSyntax ExampleFlag = "syntax"
	// EFClipped are for clipped register.
	EFClipped ExampleFlag = "clipped"
	// EFTranscribed are for examples transcribed from audio.
	EFTranscribed ExampleFlag = "transcribed"
)

type Annotation struct {
	Kind  AnnotationKind   `json:"kind" yaml:"kind"`
	Links map[string][]int `json:"links" yaml:"links"`
}

func (a *Annotation) Copy() Annotation {
	linksCopy := make(map[string][]int)
	for k, v := range a.Links {
		linksCopy[k] = v
	}

	return Annotation{
		Kind:  a.Kind,
		Links: linksCopy,
	}
}

func (a *Annotation) Validate() bool {
	switch a.Kind {
	case AKVerbParameters:
		if len(a.Links["verb"]) == 0 {
			return false
		}

		count := 0
		for _, key := range []string{"subject", "predicate", "agent", "patient", "adverb", "adverbial", "dative"} {
			if _, ok := a.Links[key]; ok {
				count += 1
			}
		}

		// One of the above keys must be there, but none outside that list.
		return count > 0 && count+1 == len(a.Links)

	case AKSplitSiVerb:
		return len(a.Links["si"]) == 1 && len(a.Links["noun"]) >= 1

	default:
		return false
	}
}

type AnnotationKind string

const (
	AKVerbParameters AnnotationKind = "verb_parameters"
	AKSplitSiVerb    AnnotationKind = "split_si_verb"
)
