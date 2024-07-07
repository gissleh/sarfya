export interface SentencePart {
  text: string
  ids?: number[]
  hiddenText?: string
  alt?: boolean
  newline?: boolean
  lp?: boolean
  rp?: boolean
}
