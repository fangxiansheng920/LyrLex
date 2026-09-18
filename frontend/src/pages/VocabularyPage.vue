<script setup lang="ts">
import { onActivated, ref } from 'vue'
import { apiDelete, apiGet, serverBase } from '../api/http'
import type { SongData, VocabularyDetail, VocabularyItem } from '../api/types'

type Lang = 'en' | 'ja'

const lang = ref<Lang>('en')
const query = ref('')
const items = ref<VocabularyItem[]>([])
const loading = ref(false)
const selected = ref<VocabularyDetail | null>(null)
const songTitles = ref<Map<number, string>>(new Map())
const toastMsg = ref('')
let toastTimer: ReturnType<typeof setTimeout> | null = null
let audio: HTMLAudioElement | null = null

function toast(msg: string): void {
  toastMsg.value = msg
  if (toastTimer) clearTimeout(toastTimer)
  toastTimer = setTimeout(() => {
    toastMsg.value = ''
  }, 2500)
}

async function load(): Promise<void> {
  loading.value = true
  try {
    items.value = await apiGet<VocabularyItem[]>(
      `/vocabulary?language=${lang.value}&q=${encodeURIComponent(query.value.trim())}`,
    )
  } catch (err) {
    toast(err instanceof Error ? err.message : '加载失败')
  } finally {
    loading.value = false
  }
}

async function loadSongs(): Promise<void> {
  try {
    const songs = await apiGet<SongData[]>('/songs')
    songTitles.value = new Map(songs.map((s) => [s.id, `${s.title} - ${s.artist}`.replace(/ - $/, '')]))
  } catch {
    /* 忽略：没有歌曲时只显示编号 */
  }
}

function switchLang(l: Lang): void {
  if (lang.value === l) return
  lang.value = l
  selected.value = null
  load()
}

async function openDetail(item: VocabularyItem): Promise<void> {
  try {
    selected.value = await apiGet<VocabularyDetail>(`/vocabulary/${lang.value}/${item.id}`)
  } catch (err) {
    toast(err instanceof Error ? err.message : '详情加载失败')
  }
}

function songTitle(id: number): string {
  return songTitles.value.get(id) || `歌曲 #${id}`
}

async function playAudio(word: string): Promise<void> {
  try {
    const base = await serverBase()
    audio?.pause()
    audio = new Audio(`${base}/api/words/audio/${lang.value}/${encodeURIComponent(word)}`)
    await audio.play()
  } catch (err) {
    toast(err instanceof Error ? err.message : '播放失败')
  }
}

async function remove(item: VocabularyItem): Promise<void> {
  if (!window.confirm(`确定删除「${item.word}」吗？`)) return
  try {
    await apiDelete(`/vocabulary/${lang.value}/${item.id}`)
    if (selected.value?.id === item.id) selected.value = null
    toast('已删除')
    await load()
  } catch (err) {
    toast(err instanceof Error ? err.message : '删除失败')
  }
}

function fmtDate(iso: string): string {
  return iso.slice(0, 10)
}

onActivated(() => {
  // 每次切回生词本时刷新列表（保留搜索词/分区/选中项），确保新收藏能出现
  load()
  if (songTitles.value.size === 0) loadSongs()
})
</script>

<template>
  <section class="page">
    <h2>生词本</h2>

    <div class="toolbar">
      <div class="tabs">
        <button :class="{ active: lang === 'en' }" @click="switchLang('en')">英文区</button>
        <button :class="{ active: lang === 'ja' }" @click="switchLang('ja')">日文区</button>
      </div>
      <div class="search">
        <input v-model="query" :placeholder="lang === 'en' ? '搜索单词' : '搜索单词 / 假名'" @keyup.enter="load" />
        <button @click="load">搜索</button>
      </div>
    </div>

    <div class="body">
      <div class="list">
        <div v-if="loading" class="empty">加载中…</div>
        <div v-else-if="items.length === 0" class="empty">暂无生词，去「歌词学习」点词收藏吧</div>
        <div
          v-for="item in items"
          :key="item.id"
          class="item"
          :class="{ active: selected?.id === item.id }"
          @click="openDetail(item)"
        >
          <div class="main">
            <span class="word">{{ item.word }}</span>
            <span v-if="item.lemma && item.lemma !== item.word" class="lemma">{{ item.lemma }}</span>
          </div>
          <div class="sub">
            <span v-if="item.phonetic">{{ item.phonetic }}</span>
            <span v-if="item.kana">{{ item.kana }}</span>
            <span v-if="item.romaji" class="romaji">{{ item.romaji }}</span>
            <span class="date">{{ fmtDate(item.created_at) }}</span>
          </div>
          <div class="actions" @click.stop>
            <button class="mini" @click="playAudio(item.word)">🔊</button>
            <button class="mini danger" @click="remove(item)">删除</button>
          </div>
        </div>
      </div>

      <div v-if="selected" class="detail">
        <div class="detail-head">
          <div>
            <strong>{{ selected.word }}</strong>
            <span v-if="selected.lemma && selected.lemma !== selected.word" class="lemma">
              （{{ selected.lemma }}）
            </span>
          </div>
          <button class="close" @click="selected = null">✕</button>
        </div>
        <div class="pron">
          <span v-if="selected.phonetic">/{{ selected.phonetic }}/</span>
          <span v-if="selected.kana">{{ selected.kana }}</span>
          <span v-if="selected.romaji" class="romaji">/ {{ selected.romaji }}</span>
          <button class="mini" @click="playAudio(selected.word)">🔊</button>
        </div>

        <h4>词义</h4>
        <ul class="meanings">
          <li v-for="m in selected.meanings" :key="m.id">
            <em v-if="m.pos">{{ m.pos }}.</em> {{ m.meaning }}
          </li>
        </ul>

        <h4>例句</h4>
        <ul class="examples">
          <li v-for="e in selected.examples" :key="e.id">
            <div class="src">{{ songTitle(e.song_id) }} · 第 {{ e.line_index + 1 }} 行</div>
            <div class="text">{{ e.text }}</div>
            <div v-if="e.translation_zh" class="trans">{{ e.translation_zh }}</div>
          </li>
        </ul>
        <div v-if="selected.examples.length === 0" class="empty small">暂无例句</div>
      </div>
    </div>

    <div v-if="toastMsg" class="toast">{{ toastMsg }}</div>
  </section>
</template>

<style scoped>
.toolbar {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 16px;
}
.tabs {
  display: flex;
  gap: 8px;
  padding: 6px;
  background: rgba(255, 255, 255, 0.75);
  border: 1px solid rgba(255, 255, 255, 0.9);
  border-radius: 999px;
  box-shadow: var(--shadow);
}
.tabs button {
  padding: 8px 20px;
  border: 1px solid rgba(74, 68, 88, 0.12);
  border-radius: 999px;
  background: #fff;
  cursor: pointer;
  font-size: 14px;
  color: var(--ink);
}
.tabs button:hover {
  background: var(--accent-soft);
}
.tabs button.active {
  background: var(--grad);
  border-color: transparent;
  color: #fff;
  font-weight: 600;
  box-shadow: 0 4px 12px rgba(90, 165, 220, 0.35);
}
.search {
  display: flex;
  gap: 8px;
}
.search input {
  padding: 9px 14px;
  border: 1px solid rgba(74, 68, 88, 0.14);
  border-radius: 999px;
  width: 220px;
  font-size: 13px;
  background: #fff;
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
.body {
  display: flex;
  gap: 16px;
  align-items: flex-start;
}
.list {
  flex: 1;
  min-width: 320px;
  max-width: 460px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  background: #fff;
  border: 1px solid rgba(255, 255, 255, 0.9);
  border-radius: 18px;
  padding: 12px 16px;
  cursor: pointer;
  box-shadow: var(--shadow);
  transition: transform 0.12s ease;
}
.item:hover {
  transform: translateY(-1px);
}
.item.active {
  border-color: var(--accent);
  box-shadow: 0 0 0 2px var(--accent-soft), var(--shadow);
}
.main {
  display: flex;
  align-items: baseline;
  gap: 8px;
}
.word {
  font-size: 16px;
  font-weight: 600;
}
.lemma {
  color: #888;
  font-size: 12px;
}
.sub {
  display: flex;
  gap: 8px;
  font-size: 12px;
  color: #666;
  margin-top: 2px;
}
.romaji {
  color: #999;
}
.date {
  color: #bbb;
}
.actions {
  display: flex;
  gap: 6px;
}
.mini {
  padding: 5px 10px;
  font-size: 12px;
  background: #fff;
  color: var(--ink);
  border: 1px solid rgba(74, 68, 88, 0.14);
  box-shadow: none;
}
.mini.danger {
  background: #ffe3ef;
  color: #d95d6a;
  border-color: transparent;
}
.detail {
  flex: 1;
  min-width: 320px;
  background: #fff;
  border: 1px solid rgba(255, 255, 255, 0.9);
  border-radius: 22px;
  padding: 18px;
  max-height: 70vh;
  overflow: auto;
  box-shadow: var(--shadow);
}
.detail-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 18px;
}
.close {
  background: #f0f0f0;
  color: #333;
  padding: 4px 8px;
}
.pron {
  margin: 6px 0 10px;
  color: #2f81f7;
  font-size: 13px;
  display: flex;
  align-items: center;
  gap: 8px;
}
.pron .romaji {
  color: #888;
}
h4 {
  margin: 12px 0 6px;
  font-size: 13px;
  color: #555;
}
.meanings,
.examples {
  margin: 0;
  padding-left: 18px;
  font-size: 13px;
  color: #333;
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.meanings em {
  color: #888;
  font-style: normal;
  font-size: 12px;
}
.examples li {
  margin-bottom: 6px;
}
.src {
  font-size: 11px;
  color: #999;
}
.text {
  color: #222;
}
.trans {
  color: #666;
  font-size: 12px;
}
.empty {
  color: #999;
  font-size: 13px;
  padding: 20px 0;
  text-align: center;
}
.empty.small {
  padding: 8px 0;
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
