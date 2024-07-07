export interface DictionaryEntry {
  id?: string
  word: string
  pos: string
  definitions: {[lang:string]: string}
  source?: string
  prefixes?: string[]
  suffixes?: string[]
  infixes?: string[]
  lenitions?: string[]
  comment?: string[]
}
