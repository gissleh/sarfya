import type { DictionaryEntry } from "./DictionaryEntry";
import type { SentencePart } from "./Sentence";

export interface ParsedSentence {
  parsed: SentencePart[]
  rawText: string
  inputText: string
  wordMap: {[id: number]: string}
  entries: {[id: number]: DictionaryEntryWithFilter[]}
}

export interface DictionaryEntryWithFilter extends DictionaryEntry {
  filter: string
}