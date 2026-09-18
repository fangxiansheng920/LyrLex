import { ref } from 'vue'

export type PageKey = 'lyric' | 'vocabulary' | 'search' | 'settings'

export const activePage = ref<PageKey>('lyric')

// 跨页传递：从「歌曲搜索」导入歌词后，跳转到「歌词学习」
export const importedLyrics = ref('')
export const importedMeta = ref<{ id: number; title: string; artist: string; language: string } | null>(null)
