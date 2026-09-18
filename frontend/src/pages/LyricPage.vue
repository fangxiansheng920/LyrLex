<script setup lang="ts">
import { computed, onActivated, ref } from 'vue'
import { apiPost, serverBase } from '../api/http'
import type { LookupData, LyricLine, Meaning, SongData, Token } from '../api/types'
import { importedLyrics, importedMeta } from '../store'

const inputText = ref('')
const title = ref('')
const artist = ref('')
const language = ref('en')
const lines = ref<LyricLine[]>([])
const songId = ref(0)
const busy = ref(false)
const refreshingLine = ref<number | null>(null)
const toastMsg = ref('')
let toastTimer: ReturnType<typeof setTimeout> | null = null

interface PopoverState {
  x: number
  y: number
  token: Token
  line: LyricLine
  lineIndex: number
  result: LookupData | null
  error: string
  selected: boolean[]
  entry: 'lemma' | 'surface'
  adding: boolean
  refreshing: boolean
}
const popover = ref<PopoverState | null>(null)

const popoverStyle = computed(() => {
  const p = popover.value
  if (!p) return {}
  return { left: `${p.x}px`, top: `${p.y}px` }
})

function toast(msg: string): void {
  toastMsg.value = msg
  if (toastTimer) clearTimeout(toastTimer)
  toastTimer = setTimeout(() => {
    toastMsg.value = ''
  }, 2500)
}

async function parseLyrics(fromImport = false): Promise<void> {
  if (!inputText.value.trim()) {
    toast('请先粘贴歌词')
    return
  }
  busy.value = true
  try {
    const data = await apiPost<{ language: string; lines: string[] }>('/lyrics/parse', {
      text: inputText.value,
    })
    language.value = data.language
    const texts = data.lines

    const [translateData, tokensData] = await Promise.all([
      apiPost<{ translations: string[] }>('/lyrics/translate', {
        source: language.value,
        target: 'zh-CN',
        lines: texts,
      }),
      Promise.all(
        texts.map((l) =>
          apiPost<{ tokens: Token[] }>('/lyrics/tokenize', { language: language.value, line: l }),
        ),
      ),
    ])

    const translations = translateData.translations ?? []
    lines.value = texts.map((text, i) => ({
      text,
      translation: translations[i] ?? '',
      tokens: tokensData[i]?.tokens ?? [],
    }))
    if (!fromImport) songId.value = 0
  } catch (err) {
    toast(err instanceof Error ? err.message : '解析失败')
  } finally {
    busy.value = false
  }
}

onActivated(async () => {
  // 从「歌曲搜索」导入歌词后跳转过来：每次激活都检查一次
  if (importedLyrics.value) {
    inputText.value = importedLyrics.value
    title.value = importedMeta.value?.title ?? ''
    artist.value = importedMeta.value?.artist ?? ''
    songId.value = importedMeta.value?.id ?? 0
    importedLyrics.value = ''
    importedMeta.value = null
    await parseLyrics(true)
  }
})

function onFileChange(event: Event): void {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  const reader = new FileReader()
  reader.onload = () => {
    inputText.value = String(reader.result ?? '')
  }
  reader.readAsText(file, 'utf-8')
}

async function onTokenClick(token: Token, line: LyricLine, lineIndex: number, event: Event): Promise<void> {
  const ev = event as MouseEvent
  const width = 340
  let x = ev.clientX + 12
  let y = ev.clientY + 12
  if (x + width > window.innerWidth) x = Math.max(8, window.innerWidth - width - 8)
  if (y + 300 > window.innerHeight) y = Math.max(8, window.innerHeight - 320)

  popover.value = {
    x,
    y,
    token,
    line,
    lineIndex,
    result: null,
    error: '',
    selected: [],
    entry: 'lemma',
    adding: false,
    refreshing: false,
  }

  try {
    const result = await fetchLookup(token, line)
    if (popover.value) applyLookup(popover.value, result)
  } catch (err) {
    if (popover.value) popover.value.error = err instanceof Error ? err.message : '查词失败'
  }
}

function applyLookup(p: PopoverState, result: LookupData): void {
  p.result = result
  p.selected = result.meanings.map(() => true)
  p.entry = result.lemma && result.lemma !== result.surface ? 'lemma' : 'surface'
  p.error = ''
}

function fetchLookup(token: Token, line: LyricLine, force = false): Promise<LookupData> {
  return apiPost<LookupData>('/words/lookup', {
    language: language.value,
    word: token.surface,
    context: line.text,
    force, // true = 跳过缓存，重新查询
  })
}

async function refreshLookup(): Promise<void> {
  const p = popover.value
  if (!p || p.refreshing) return
  p.refreshing = true
  try {
    applyLookup(p, await fetchLookup(p.token, p.line, true))
  } catch (err) {
    p.error = err instanceof Error ? err.message : '重新查询失败'
  } finally {
    p.refreshing = false
  }
}

async function playAudio(): Promise<void> {
  const p = popover.value
  if (!p) return
  try {
    const base = await serverBase()
    const word = p.result?.surface || p.token.surface
    const url = `${base}/api/words/audio/${language.value}/${encodeURIComponent(word)}`
    const audio = new Audio(url)
    await audio.play()
  } catch (err) {
    toast(err instanceof Error ? err.message : '播放失败')
  }
}

async function addToVocabulary(): Promise<void> {
  const p = popover.value
  if (!p || !p.result) return
  p.adding = true
  try {
    const meanings: Meaning[] = p.result.meanings
      .filter((_, i) => p.selected[i])
      .map((m) => ({ pos: m.pos, meaning: m.meaning }))
    await apiPost('/vocabulary', {
      language: language.value,
      word: p.entry === 'lemma' ? p.result.lemma : p.result.surface,
      lemma: p.result.lemma,
      kana: p.result.kana,
      romaji: p.result.romaji,
      meanings,
      context: {
        text: p.line.text,
        translation_zh: p.line.translation,
        song_id: songId.value,
        line_index: p.lineIndex,
      },
    })
    popover.value = null
    toast('已加入生词本')
  } catch (err) {
    toast(err instanceof Error ? err.message : '收藏失败')
  } finally {
    p.adding = false
  }
}

async function retranslate(lineIndex: number): Promise<void> {
  const line = lines.value[lineIndex]
  if (!line || refreshingLine.value !== null) return
  refreshingLine.value = lineIndex
  try {
    const data = await apiPost<{ translations: string[] }>('/lyrics/translate', {
      source: language.value,
      target: 'zh-CN',
      lines: [line.text],
      force: true, // 跳过缓存，用当前翻译源重翻
    })
    const translated = data.translations?.[0] ?? ''
    if (translated) {
      line.translation = translated
    } else {
      toast('该行暂无可翻译内容')
    }
  } catch (err) {
    toast(err instanceof Error ? err.message : '重新翻译失败')
  } finally {
    refreshingLine.value = null
  }
}

function resetAll(): void {
  inputText.value = ''
  title.value = ''
  artist.value = ''
  language.value = 'en'
  lines.value = []
  songId.value = 0
  popover.value = null
  toast('已重置')
}

async function saveSong(): Promise<void> {
  if (lines.value.length === 0) {
    toast('请先解析歌词')
    return
  }
  try {
    const song = await apiPost<SongData>('/songs', {
      title: title.value.trim() || '未命名歌曲',
      artist: artist.value.trim(),
      language: language.value,
      lines: lines.value.map((l) => ({
        text: l.text,
        translation_zh: l.translation,
        tokens: l.tokens,
      })),
    })
    songId.value = song.id
    toast('歌曲已保存')
  } catch (err) {
    toast(err instanceof Error ? err.message : '保存失败')
  }
}
</script>

<template>
  <section class="page">
    <h2>歌词学习</h2>

    <div class="toolbar">
      <textarea
        v-model="inputText"
        placeholder="在此粘贴歌词（英文 / 日文），或上传 .txt / .lrc 文件"
      ></textarea>
      <div class="actions">
        <label class="file-btn">
          上传歌词
          <input type="file" accept=".txt,.lrc" @change="onFileChange" />
        </label>
        <button :disabled="busy" @click="parseLyrics()">{{ busy ? '处理中…' : '解析并翻译' }}</button>
      </div>
      <div class="meta">
        <input v-model="title" placeholder="歌名（可选）" />
        <input v-model="artist" placeholder="歌手（可选）" />
        <button :disabled="busy || lines.length === 0" @click="saveSong">保存歌曲</button>
        <button class="ghost" :disabled="busy" @click="resetAll">重置</button>
        <span v-if="language" class="lang">语言：{{ language === 'en' ? '英文' : '日文' }}</span>
      </div>
    </div>

    <div v-if="lines.length" class="lyrics">
      <div v-for="(line, li) in lines" :key="li" class="lyric-line">
        <p class="original">
          <template v-if="line.tokens.length">
            <span
              v-for="(t, ti) in line.tokens"
              :key="ti"
              :class="{ word: t.addable }"
              @click="t.addable && onTokenClick(t, line, li, $event)"
            >{{ t.surface }}</span>
          </template>
          <template v-else>{{ line.text }}</template>
        </p>
        <p class="translation-row">
          <span class="translation">{{ line.translation || '（暂无译文，点 ↻ 重试）' }}</span>
          <button
            class="refresh"
            :disabled="refreshingLine !== null"
            title="重新翻译"
            @click="retranslate(li)"
          >
            <span :class="{ spinning: refreshingLine === li }">↻</span>
          </button>
        </p>
      </div>
    </div>

    <div v-if="popover" class="popover" :style="popoverStyle" @click.stop>
      <template v-if="popover.result">
        <div class="head">
          <strong>{{ popover.result.surface }}</strong>
          <span
            v-if="popover.result.lemma && popover.result.lemma !== popover.result.surface"
            class="lemma"
          >
            → {{ popover.result.lemma }}
          </span>
          <div class="head-actions">
            <button class="icon-btn" title="重新查询" @click="refreshLookup">
              <span :class="{ spinning: popover.refreshing }">↻</span>
            </button>
            <button v-if="popover.result.audio" class="icon-btn" title="发音" @click="playAudio">
              🔊
            </button>
            <button class="icon-btn" @click="popover = null">✕</button>
          </div>
        </div>
        <div v-if="popover.result.phonetic" class="phonetic">{{ popover.result.phonetic }}</div>
        <div v-if="popover.result.kana" class="phonetic">
          {{ popover.result.kana }}
          <span v-if="popover.result.romaji" class="romaji">/ {{ popover.result.romaji }}</span>
        </div>
        <div class="meanings">
          <label v-for="(m, mi) in popover.result.meanings" :key="mi" class="meaning">
            <input v-model="popover.selected[mi]" type="checkbox" />
            <span><em v-if="m.pos">{{ m.pos }}.</em> {{ m.meaning }}</span>
          </label>
        </div>
        <div class="entry-choose">
          <label>
            <input v-model="popover.entry" type="radio" value="lemma" />
            以基本形「{{ popover.result.lemma }}」入库
          </label>
          <label>
            <input v-model="popover.entry" type="radio" value="surface" />
            以原形「{{ popover.result.surface }}」入库
          </label>
        </div>
        <button class="add" :disabled="popover.adding" @click="addToVocabulary">
          {{ popover.adding ? '加入中…' : '加入生词本' }}
        </button>
      </template>
      <template v-else-if="popover.error">
        <div class="error">{{ popover.error }}</div>
        <button class="close" @click="popover = null">✕</button>
      </template>
      <template v-else>
        <div class="loading">查询中…</div>
      </template>
    </div>

    <div v-if="toastMsg" class="toast">{{ toastMsg }}</div>
  </section>
</template>

<style scoped>
.toolbar {
  display: flex;
  flex-direction: column;
  gap: 10px;
  margin-bottom: 20px;
  max-width: 720px;
}
textarea {
  height: 140px;
  resize: vertical;
  padding: 12px 14px;
  border: 1px solid rgba(74, 68, 88, 0.14);
  border-radius: 18px;
  font-size: 14px;
  line-height: 1.6;
  background: #fff;
}
.actions {
  display: flex;
  gap: 10px;
  align-items: center;
}
.file-btn {
  position: relative;
  display: inline-block;
  padding: 9px 16px;
  border: 1px solid rgba(74, 68, 88, 0.16);
  border-radius: 999px;
  background: #fff;
  cursor: pointer;
  font-size: 13px;
  color: var(--ink);
}
.file-btn input {
  display: none;
}
button {
  padding: 9px 16px;
  border: none;
  border-radius: 999px;
  background: var(--grad);
  color: #fff;
  cursor: pointer;
  font-size: 13px;
  font-weight: 600;
  box-shadow: 0 4px 12px rgba(90, 165, 220, 0.35);
}
button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
button.ghost {
  background: #fff;
  color: var(--ink);
  border: 1px solid rgba(74, 68, 88, 0.16);
  box-shadow: none;
}
.meta {
  display: flex;
  gap: 8px;
  align-items: center;
  flex-wrap: wrap;
}
.meta input {
  padding: 9px 14px;
  border: 1px solid rgba(74, 68, 88, 0.14);
  border-radius: 999px;
  font-size: 13px;
  width: 170px;
  background: #fff;
}
.lang {
  font-size: 12px;
  color: #555;
}
.lyrics {
  max-width: 720px;
}
.lyric-line {
  margin-bottom: 14px;
}
.original {
  margin: 0;
  font-size: 17px;
  line-height: 1.9;
  color: var(--ink);
}
.original .word {
  cursor: pointer;
  border-radius: 8px;
  padding: 0 3px;
}
.original .word:hover {
  background: var(--accent-soft);
}
.translation-row {
  display: flex;
  align-items: center;
  gap: 6px;
  margin: 2px 0 0;
}
.translation {
  font-size: 13px;
  color: var(--muted);
}
button.refresh {
  padding: 1px 8px;
  font-size: 12px;
  line-height: 1.5;
  background: #fff;
  color: var(--accent-deep);
  border: 1px solid rgba(74, 68, 88, 0.14);
  box-shadow: none;
}
.spinning {
  display: inline-block;
  animation: spin 0.8s linear infinite;
}
@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}
.popover {
  position: fixed;
  width: 340px;
  max-height: 320px;
  overflow: auto;
  background: #fff;
  border: 1px solid rgba(255, 255, 255, 0.9);
  border-radius: 22px;
  box-shadow: 0 12px 40px rgba(180, 140, 170, 0.3);
  padding: 16px;
  z-index: 100;
}
.head {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 16px;
}
.lemma {
  color: var(--muted);
  font-size: 13px;
}
.speaker,
.close {
  background: #fff;
  color: var(--ink);
  border: 1px solid rgba(74, 68, 88, 0.14);
  padding: 5px 9px;
  font-size: 13px;
  margin-left: auto;
  box-shadow: none;
}
.close {
  margin-left: 4px;
}
.head-actions {
  margin-left: auto;
  display: flex;
  gap: 4px;
}
button.icon-btn {
  background: #fff;
  color: var(--ink);
  border: 1px solid rgba(74, 68, 88, 0.14);
  padding: 5px 9px;
  font-size: 13px;
  box-shadow: none;
}
.phonetic {
  color: var(--accent-deep);
  font-size: 13px;
  margin-top: 4px;
}
.romaji {
  color: var(--muted);
  font-size: 12px;
}
.meanings {
  margin-top: 10px;
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.meaning {
  display: flex;
  gap: 8px;
  align-items: flex-start;
  font-size: 13px;
  color: #333;
}
.meaning em {
  color: #888;
  font-style: normal;
  font-size: 12px;
}
.entry-choose {
  margin-top: 10px;
  display: flex;
  flex-direction: column;
  gap: 4px;
  font-size: 12px;
  color: #555;
}
.add {
  width: 100%;
  margin-top: 12px;
  padding: 9px;
}
.loading,
.error {
  font-size: 13px;
  color: #888;
}
.error {
  color: #c0392b;
}
.toast {
  position: fixed;
  left: 50%;
  bottom: 32px;
  transform: translateX(-50%);
  background: rgba(0, 0, 0, 0.78);
  color: #fff;
  padding: 9px 18px;
  border-radius: 999px;
  font-size: 13px;
  z-index: 200;
}
</style>
