export interface Token {
  surface: string
  lemma: string
  pos: string
  reading: string
  romaji: string
  accent: string
  group_id: number
  addable: boolean
}

export interface Meaning {
  pos: string
  meaning: string
}

export interface LookupData {
  surface: string
  lemma: string
  phonetic: string
  kana: string
  romaji: string
  meanings: Meaning[]
  audio: boolean
  source: string
}

export interface ParseData {
  language: string
  lines: string[]
}

export interface LyricLine {
  text: string
  translation: string
  tokens: Token[]
}

export interface SongData {
  id: number
  title: string
  artist: string
  language: string
  source: string
  created_at: string
}

export interface SongMeta {
  id: string
  title: string
  artist: string
}

export interface SongLineData {
  id: number
  song_id: number
  line_index: number
  text: string
  translation_zh: string
  tokens_json: string
}

export interface ImportData {
  song: SongData
  lines: SongLineData[]
}

export interface VocabularyItem {
  id: number
  word: string
  lemma: string
  phonetic: string
  kana: string
  romaji: string
  created_at: string
}

export interface WordSense {
  id: number
  word_type: string
  word_id: number
  pos: string
  meaning: string
  source: string
}

export interface Example {
  id: number
  word_type: string
  word_id: number
  sense_id: number
  song_id: number
  line_index: number
  text: string
  translation_zh: string
}

export interface VocabularyDetail {
  id: number
  word: string
  lemma: string
  phonetic: string
  kana: string
  romaji: string
  meanings: WordSense[]
  examples: Example[]
  created_at: string
}
