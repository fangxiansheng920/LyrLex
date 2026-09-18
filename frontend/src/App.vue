<script setup lang="ts">
import { computed, onMounted } from 'vue'
import LyricPage from './pages/LyricPage.vue'
import VocabularyPage from './pages/VocabularyPage.vue'
import SearchPage from './pages/SearchPage.vue'
import SettingsPage from './pages/SettingsPage.vue'
import { applyTheme, loadTheme } from './theme/themes'
import { activePage, type PageKey } from './store'

const pages: { key: PageKey; label: string; component: unknown }[] = [
  { key: 'lyric', label: '歌词学习', component: LyricPage },
  { key: 'vocabulary', label: '生词本', component: VocabularyPage },
  { key: 'search', label: '歌曲搜索', component: SearchPage },
  { key: 'settings', label: '设置中心', component: SettingsPage },
]

const current = computed(() => pages.find((p) => p.key === activePage.value)?.component)

onMounted(() => {
  applyTheme(loadTheme())
})
</script>

<template>
  <div class="app-shell">
    <aside class="sidebar">
      <div class="logo">LyrLex</div>
      <nav>
        <button
          v-for="p in pages"
          :key="p.key"
          :class="{ active: activePage === p.key }"
          @click="activePage = p.key"
        >
          {{ p.label }}
        </button>
      </nav>
    </aside>
    <main class="content">
      <component :is="current" />
    </main>
  </div>
</template>

<style scoped>
.app-shell {
  display: flex;
  height: 100%;
  background: linear-gradient(135deg, var(--bg-a), var(--bg-b), var(--bg-c));
}
.sidebar {
  width: 220px;
  margin: 14px;
  padding: 16px 12px;
  background: rgba(255, 255, 255, 0.72);
  backdrop-filter: blur(12px);
  border: 1px solid rgba(255, 255, 255, 0.9);
  border-radius: 22px;
  box-shadow: var(--shadow);
  display: flex;
  flex-direction: column;
  gap: 14px;
}
.logo {
  font-weight: 700;
  font-size: 16px;
  padding: 12px 14px;
  background: var(--grad);
  color: #fff;
  border-radius: 16px;
  box-shadow: 0 4px 14px rgba(90, 165, 220, 0.35);
}
nav {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
nav button {
  text-align: left;
  padding: 11px 14px;
  border: 1px solid transparent;
  border-radius: 14px;
  background: transparent;
  cursor: pointer;
  font-size: 14px;
  color: var(--ink);
  transition: all 0.15s ease;
}
nav button:hover {
  background: var(--accent-soft);
}
nav button.active {
  background: var(--grad);
  color: #fff;
  font-weight: 600;
  box-shadow: 0 4px 12px rgba(90, 165, 220, 0.35);
}
.content {
  flex: 1;
  overflow: auto;
  padding: 24px;
}
</style>
