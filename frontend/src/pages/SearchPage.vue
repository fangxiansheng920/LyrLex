<script setup lang="ts">
import { onActivated, ref } from 'vue'
import { apiGet, apiPost } from '../api/http'
import type { ImportData, SongData, SongMeta } from '../api/types'
import { activePage, importedLyrics, importedMeta } from '../store'

const query = ref('')
const searching = ref(false)
const importing = ref<string>('')
const results = ref<SongMeta[]>([])
const localSongs = ref<SongData[]>([])
const searched = ref(false)
const selectMode = ref(false)
const selectedIds = ref<number[]>([])
const toastMsg = ref('')
let toastTimer: ReturnType<typeof setTimeout> | null = null

function toast(msg: string): void {
  toastMsg.value = msg
  if (toastTimer) clearTimeout(toastTimer)
  toastTimer = setTimeout(() => {
    toastMsg.value = ''
  }, 2500)
}

async function loadLocalSongs(): Promise<void> {
  try {
    localSongs.value = await apiGet<SongData[]>('/songs')
  } catch {
    /* 忽略 */
  }
}

async function search(): Promise<void> {
  if (!query.value.trim()) {
    toast('请输入歌名或歌手')
    return
  }
  searching.value = true
  searched.value = true
  try {
    results.value = await apiPost<SongMeta[]>('/songs/search', { query: query.value.trim() })
  } catch (err) {
    results.value = []
    toast(err instanceof Error ? err.message : '在线搜索不可用，请粘贴或上传歌词')
  } finally {
    searching.value = false
  }
}

function goManual(): void {
  importedLyrics.value = ''
  importedMeta.value = null
  activePage.value = 'lyric'
}

async function importSong(meta: SongMeta): Promise<void> {
  importing.value = meta.id
  try {
    const data = await apiPost<ImportData>('/songs/import', {
      song_id: meta.id,
      title: meta.title,
      artist: meta.artist,
    })
    const text = data.lines
      .map((l) => l.text)
      .filter((t) => t.trim() !== '')
      .join('\n')
    if (!text) {
      toast('该歌曲没有可导入的歌词')
      return
    }
    importedLyrics.value = text
    importedMeta.value = {
      id: data.song.id,
      title: data.song.title,
      artist: data.song.artist,
      language: data.song.language,
    }
    activePage.value = 'lyric'
    toast('已导入，正在解析歌词…')
  } catch (err) {
    toast(err instanceof Error ? err.message : '导入失败，请粘贴或上传歌词')
  } finally {
    importing.value = ''
  }
}

async function openLocalSong(song: SongData): Promise<void> {
  try {
    const data = await apiGet<{ song: SongData; lines: { text: string }[] }>(`/songs/${song.id}/lyrics`)
    const text = data.lines
      .map((l) => l.text)
      .filter((t) => t.trim() !== '')
      .join('\n')
    importedLyrics.value = text
    importedMeta.value = {
      id: song.id,
      title: song.title,
      artist: song.artist,
      language: song.language,
    }
    activePage.value = 'lyric'
  } catch (err) {
    toast(err instanceof Error ? err.message : '打开失败')
  }
}

function toggleSelectMode(): void {
  selectMode.value = !selectMode.value
  selectedIds.value = []
}

function toggleSelected(id: number): void {
  const i = selectedIds.value.indexOf(id)
  if (i >= 0) {
    selectedIds.value.splice(i, 1)
  } else {
    selectedIds.value.push(id)
  }
}

async function batchDelete(): Promise<void> {
  if (selectedIds.value.length === 0) return
  if (!window.confirm(`确定删除选中的 ${selectedIds.value.length} 首歌吗？`)) return
  try {
    await apiPost('/songs/batch-delete', { ids: selectedIds.value })
    toast('已删除')
    selectMode.value = false
    selectedIds.value = []
    await loadLocalSongs()
  } catch (err) {
    toast(err instanceof Error ? err.message : '删除失败')
  }
}

onActivated(loadLocalSongs)
</script>

<template>
  <section class="page">
    <h2>歌曲搜索</h2>

    <div class="toolbar">
      <input v-model="query" placeholder="输入歌名 / 歌手" @keyup.enter="search" />
      <button :disabled="searching" @click="search">{{ searching ? '搜索中…' : '在线搜索' }}</button>
      <button class="ghost" @click="goManual">粘贴 / 上传歌词</button>
    </div>

    <div class="hint">在线搜索依赖 NeteaseCloudMusicApi 服务（见 README）；不可用时请直接粘贴或上传。</div>

    <div v-if="searched" class="section">
      <h3>搜索结果</h3>
      <div v-if="results.length === 0" class="empty">无结果或服务不可用</div>
      <div v-for="m in results" :key="m.id" class="row">
        <div class="info">
          <div class="title">{{ m.title }}</div>
          <div class="artist">{{ m.artist }}</div>
        </div>
        <button :disabled="importing === m.id" @click="importSong(m)">
          {{ importing === m.id ? '导入中…' : '导入' }}
        </button>
      </div>
    </div>

    <div class="section">
      <div class="section-head">
        <h3>本地歌曲库</h3>
        <div class="section-actions">
          <template v-if="!selectMode">
            <button v-if="localSongs.length" class="ghost" @click="toggleSelectMode">选择</button>
          </template>
          <template v-else>
            <button class="danger" :disabled="selectedIds.length === 0" @click="batchDelete">
              删除{{ selectedIds.length ? `(${selectedIds.length})` : '' }}
            </button>
            <button class="ghost" @click="toggleSelectMode">取消</button>
          </template>
        </div>
      </div>
      <div v-if="localSongs.length === 0" class="empty">暂无已导入的歌曲</div>
      <div v-for="s in localSongs" :key="s.id" class="row" :class="{ selectable: selectMode }">
        <input
          v-if="selectMode"
          type="checkbox"
          class="song-check"
          :checked="selectedIds.includes(s.id)"
          @change="toggleSelected(s.id)"
        />
        <div class="info">
          <div class="title">{{ s.title }}</div>
          <div class="artist">{{ s.artist }} · {{ s.language === 'ja' ? '日文' : '英文' }}</div>
        </div>
        <div v-if="!selectMode" class="row-actions">
          <button class="ghost" @click="openLocalSong(s)">打开</button>
        </div>
      </div>
    </div>

    <div v-if="toastMsg" class="toast">{{ toastMsg }}</div>
  </section>
</template>

<style scoped>
.toolbar {
  display: flex;
  gap: 8px;
  align-items: center;
}
.toolbar input {
  padding: 9px 16px;
  border: 1px solid rgba(74, 68, 88, 0.14);
  border-radius: 999px;
  width: 260px;
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
.row-actions {
  display: flex;
  gap: 6px;
}
button.danger {
  background: #ffe3ef;
  color: #d95d6a;
  border: 1px solid transparent;
  box-shadow: none;
}
.hint {
  margin-top: 8px;
  font-size: 12px;
  color: var(--muted);
}
.section {
  margin-top: 20px;
  max-width: 640px;
}
.section-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.section-actions {
  display: flex;
  gap: 6px;
}
.section-head h3 {
  margin-bottom: 0;
}
h3 {
  font-size: 14px;
  color: var(--muted);
  margin: 0 0 8px;
}
.song-check {
  width: 16px;
  height: 16px;
  accent-color: var(--accent);
  cursor: pointer;
  flex: none;
}
.info {
  flex: 1;
  min-width: 0;
}
.row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 10px;
  background: #fff;
  border: 1px solid rgba(255, 255, 255, 0.9);
  border-radius: 18px;
  padding: 12px 16px;
  margin-bottom: 10px;
  box-shadow: var(--shadow);
}
.title {
  font-size: 14px;
  font-weight: 600;
}
.artist {
  font-size: 12px;
  color: var(--muted);
  margin-top: 2px;
}
.empty {
  color: #999;
  font-size: 13px;
  padding: 12px 0;
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
