import type { DictionaryEntry } from "./DictionaryEntry"
import type { ExampleSource } from "./ExampleData"

export interface Input {
  id?: string
  text: string
  lookupFilter: {[id:number]: string}
  translations: {[lang:string]: string}
  source: ExampleSource
  annotations: never[]
  flags: string[]
}

export interface InputLookupConstraint {
  id?: string
  pos?: string
  prefixes?: string[]
  infixes?: string[]
  suffixes?: string[]
  lenitions?: string[]
}

export function constraintFrom(e: DictionaryEntry): InputLookupConstraint {
  return {
    id: e.id,
    infixes: e.infixes,
    lenitions: e.lenitions,
    prefixes: e.prefixes,
    suffixes: e.suffixes,
  }
}

export function constraintsEqual(a: InputLookupConstraint, b: InputLookupConstraint) {
  if (a.id !== b.id 
    || a.pos !== b.pos 
    || !!a.prefixes !== !!b.prefixes
    || !!a.infixes !== !!b.infixes
    || !!a.suffixes !== !!b.suffixes
    || !!a.lenitions !== !!b.lenitions
  ) {
    return false
  }

  if (a.prefixes?.find(p => !b.prefixes?.includes(p))) {
    return false
  }
  if (a.infixes?.find(p => !b.infixes?.includes(p))) {
    return false
  }
  if (a.suffixes?.find(p => !b.suffixes?.includes(p))) {
    return false
  }
  if (a.lenitions?.find(p => !b.lenitions?.includes(p))) {
    return false
  }

  return true
}

export function blankInput(id: string, defaultSource?: Partial<ExampleSource>): Input {
  const date = (new Date).toISOString().slice(0, 10) as `${number}-${number}-${number}`;

  return {
    id: id,
    text: "",
    lookupFilter: {},
    annotations: [],
    flags: [],
    translations: { en: "" },
    source: { id: "", author: "", date: date, title: "", url: "", ...(defaultSource||{}) },
  }
}