import type { DictionaryEntry } from "./DictionaryEntry"
import type { SentencePart } from "./Sentence"

export interface ExampleSet {
  entries?: DictionaryEntry[]
  examples: ExampleData[]
}

export interface ExampleData {
  id: string
  text: SentencePart[]
  translations: {[lang:string]: SentencePart[]}
  annotations: []
  source: ExampleSource
  words: {[id:number]: DictionaryEntry[]}
  spans: number[][]
  translatedSpans: {[lang:string]: number[][]}
  translatedAdjacent: {[lang:string]: number[][]}
  wordMap?: {[id:number]: string}
  flags?: string[]
}

export interface ExampleSource {
  id: string
  date: `${number}-${number}-${number}`
  url: string
  title: string
  author: string
}