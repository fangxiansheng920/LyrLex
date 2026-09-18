<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { apiGet, apiPut } from '../api/http'
import {
  loadTheme,
  presetThemes,
  saveTheme,
  type Theme,
  type ThemeColors,
} from '../theme/themes'

interface SettingsData {
  api_keys: {
    youdao_app_key: string
    youdao_app_secret: string
    baidu_app_id: string
    baidu_secret: string
    deepl_key: string
  }
  pronunciation_source: string
}

const currentTheme = ref<Theme>(loadTheme())
const customColor = ref(currentTheme.value.colors[0])
const gradA = ref(currentTheme.value.colors[0])
const gradB = ref(currentTheme.value.colors[1])
const gradC = ref(currentTheme.value.colors[2])
const buttonColor = ref(currentTheme.value.accent)
const settings = ref<SettingsData>({
  api_keys: {
    youdao_app_key: '',
    youdao_app_secret: '',
    baidu_app_id: '',
    baidu_secret: '',
    deepl_key: '',
  },
  pronunciation_source: 'youdao',
})
const dataDir = ref('')
const saving = ref(false)
const toastMsg = ref('')
let toastTimer: ReturnType<typeof setTimeout> | null = null

function toast(msg: string): void {
  toastMsg.value = msg
  if (toastTimer) clearTimeout(toastTimer)
  toastTimer = setTimeout(() => {
    toastMsg.value = ''
  }, 2500)
}

function pick(theme: Theme): void {
  // 预设只换背景，保留当前按钮颜色
  const next: Theme = { ...theme, accent: buttonColor.value }
  currentTheme.value = next
  saveTheme(next)
}

function applyCustom(): void {
  const colors: ThemeColors = [customColor.value, customColor.value, customColor.value]
  const next: Theme = { name: '自定义', colors, accent: buttonColor.value }
  currentTheme.value = next
  saveTheme(next)
}

function applyGradient(): void {
  const colors: ThemeColors = [gradA.value, gradB.value, gradC.value]
  const next: Theme = { name: '自定义渐变', colors, accent: buttonColor.value }
  currentTheme.value = next
  saveTheme(next)
}

function applyAccent(): void {
  const next: Theme = { ...currentTheme.value, accent: buttonColor.value }
  currentTheme.value = next
  saveTheme(next)
}

async function loadSettings(): Promise<void> {
  try {
    settings.value = await apiGet<SettingsData>('/settings')
  } catch (err) {
    toast(err instanceof Error ? err.message : '设置加载失败')
  }
}

async function loadDataDir(): Promise<void> {
  try {
    const data = await apiGet<{ path: string }>('/settings/datadir')
    dataDir.value = data.path
  } catch {
    /* 忽略 */
  }
}

async function saveSettings(): Promise<void> {
  saving.value = true
  try {
    settings.value = await apiPut<SettingsData>('/settings', {
      api_keys: settings.value.api_keys,
      pronunciation_source: settings.value.pronunciation_source,
    })
    toast('已保存')
  } catch (err) {
    toast(err instanceof Error ? err.message : '保存失败')
  } finally {
    saving.value = false
  }
}

async function openDataDir(): Promise<void> {
  if (window.api?.openDataDir) {
    try {
      await window.api.openDataDir()
    } catch (err) {
      toast(err instanceof Error ? err.message : '打开失败')
    }
  }
}

onMounted(() => {
  loadSettings()
  loadDataDir()
})
</script>

<template>
  <section class="page">
    <h2>设置中心</h2>

    <h3>主题配色</h3>
    <div class="themes">
      <button v-for="t in presetThemes" :key="t.name" class="theme-card" @click="pick(t)">
        <span
          class="swatch"
          :style="{ background: `linear-gradient(135deg, ${t.colors.join(', ')})` }"
        ></span>
        {{ t.name }}
      </button>
    </div>

    <h3>自定义纯色</h3>
    <div class="custom">
      <input v-model="customColor" type="color" />
      <input v-model="customColor" type="text" />
      <button class="ghost" @click="applyCustom">应用</button>
    </div>

    <h3>自定义渐变色</h3>
    <div class="custom">
      <input v-model="gradA" type="color" />
      <input v-model="gradA" type="text" />
      <input v-model="gradB" type="color" />
      <input v-model="gradB" type="text" />
      <input v-model="gradC" type="color" />
      <input v-model="gradC" type="text" />
      <button class="ghost" @click="applyGradient">应用</button>
    </div>

    <h3>按钮颜色</h3>
    <div class="custom">
      <input v-model="buttonColor" type="color" />
      <input v-model="buttonColor" type="text" />
      <button class="ghost" @click="applyAccent">应用</button>
    </div>

    <h3>翻译 API 密钥（留空则使用默认免费源）</h3>
    <div class="form">
      <label>有道 App Key<input v-model="settings.api_keys.youdao_app_key" /></label>
      <label>有道 App Secret<input v-model="settings.api_keys.youdao_app_secret" type="password" /></label>
      <label>百度 App ID<input v-model="settings.api_keys.baidu_app_id" /></label>
      <label>百度密钥<input v-model="settings.api_keys.baidu_secret" type="password" /></label>
      <label>DeepL Auth Key<input v-model="settings.api_keys.deepl_key" type="password" /></label>
    </div>

    <h3>发音源</h3>
    <div class="form">
      <label>
        <select v-model="settings.pronunciation_source">
          <option value="youdao">有道词典发音（默认）</option>
        </select>
      </label>
      <p class="hint">当前仅内置有道发音，后续可扩展 Edge TTS 等。</p>
    </div>

    <button class="primary" :disabled="saving" @click="saveSettings">
      {{ saving ? '保存中…' : '保存设置' }}
    </button>

    <h3>数据目录</h3>
    <div class="datadir">
      <code>{{ dataDir || '获取中…' }}</code>
      <button class="ghost" @click="openDataDir">打开目录</button>
    </div>

    <div v-if="toastMsg" class="toast">{{ toastMsg }}</div>
  </section>
</template>

<style scoped>
h3 {
  margin: 20px 0 8px;
  font-size: 14px;
  color: #555;
}
.themes {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}
.theme-card {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 9px 14px;
  border: 1px solid rgba(74, 68, 88, 0.12);
  border-radius: 999px;
  background: #ffffff;
  cursor: pointer;
  box-shadow: var(--shadow);
}
.swatch {
  width: 48px;
  height: 24px;
  border-radius: 999px;
  display: inline-block;
}
.custom {
  display: flex;
  gap: 8px;
  align-items: center;
  flex-wrap: wrap;
}
.form {
  display: flex;
  flex-direction: column;
  gap: 10px;
  max-width: 420px;
}
.form label {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  font-size: 13px;
}
.form input,
.form select {
  flex: 1;
  max-width: 260px;
  padding: 8px 12px;
  border: 1px solid rgba(74, 68, 88, 0.14);
  border-radius: 999px;
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
button.primary {
  margin-top: 16px;
}
.hint {
  color: var(--muted);
  font-size: 12px;
}
.datadir {
  display: flex;
  gap: 10px;
  align-items: center;
}
.datadir code {
  background: var(--accent-soft);
  padding: 8px 12px;
  border-radius: 999px;
  font-size: 12px;
  word-break: break-all;
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
