<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { apiGet, apiPost, apiPut } from '../api/http'
import {
  loadBackgroundImage,
  loadBackgroundOpacity,
  loadTheme,
  presetThemes,
  saveBackgroundImage,
  saveBackgroundOpacity,
  saveTheme,
  type Theme,
  type ThemeColors,
} from '../theme/themes'

interface TranslationTestResult {
  provider: string
  ok: boolean
  result: string
  error_code: string
  error: string
}

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
const testing = ref(false)
const testResult = ref<TranslationTestResult | null>(null)
const bgPreview = ref<string | null>(null)
const bgOpacity = ref(Math.round(loadBackgroundOpacity() * 100))
let toastTimer: ReturnType<typeof setTimeout> | null = null

// 有道常见错误码的中文说明
const youdaoErrMsg: Record<string, string> = {
  '108': '应用 ID 无效（检查 appKey）',
  '110': '应用未绑定「文本翻译」服务',
  '202': '签名校验失败（密钥不对或编码问题）',
  '401': '账户已欠费，请充值',
  '411': '访问频率受限，稍后再试',
  '205': '接入方式选错',
}

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

function compressImage(file: File, maxWidth: number, quality: number): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => {
      const img = new Image()
      img.onload = () => {
        const scale = Math.min(1, maxWidth / img.width)
        const w = Math.max(1, Math.round(img.width * scale))
        const h = Math.max(1, Math.round(img.height * scale))
        const canvas = document.createElement('canvas')
        canvas.width = w
        canvas.height = h
        const ctx = canvas.getContext('2d')
        if (!ctx) {
          resolve(String(reader.result))
          return
        }
        ctx.drawImage(img, 0, 0, w, h)
        resolve(canvas.toDataURL('image/jpeg', quality))
      }
      img.onerror = () => reject(new Error('图片读取失败'))
      img.src = String(reader.result)
    }
    reader.onerror = () => reject(new Error('图片读取失败'))
    reader.readAsDataURL(file)
  })
}

async function onBgFile(event: Event): Promise<void> {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  try {
    const dataUrl = await compressImage(file, 1920, 0.82)
    bgPreview.value = dataUrl
    saveBackgroundImage(dataUrl)
    toast('已应用背景')
  } catch (err) {
    toast(err instanceof Error ? err.message : '背景设置失败')
  } finally {
    input.value = ''
  }
}

function clearBg(): void {
  bgPreview.value = null
  saveBackgroundImage(null)
  toast('已清除背景')
}

function onOpacityChange(): void {
  saveBackgroundOpacity(bgOpacity.value / 100)
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

async function testTranslation(): Promise<void> {
  testing.value = true
  testResult.value = null
  try {
    await saveSettings() // 先保存当前填写的密钥，再测试
    testResult.value = await apiPost<TranslationTestResult>('/settings/test-translation', {})
  } catch (err) {
    testResult.value = {
      provider: 'youdao',
      ok: false,
      result: '',
      error_code: '',
      error: err instanceof Error ? err.message : '测试失败',
    }
  } finally {
    testing.value = false
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
  bgPreview.value = loadBackgroundImage()
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

    <h3>自定义背景图片</h3>
    <div class="custom">
      <label class="file-btn">
        选择图片
        <input type="file" accept="image/*" @change="onBgFile" />
      </label>
      <button class="ghost" @click="clearBg">清除背景</button>
    </div>
    <img v-if="bgPreview" :src="bgPreview" class="bg-preview" alt="背景预览" />
    <div class="opacity-row">
      <span>背景透明度</span>
      <input v-model.number="bgOpacity" type="range" min="0" max="100" @input="onOpacityChange" />
      <span class="opacity-val">{{ bgOpacity }}%</span>
    </div>
    <p class="hint">图片会自动压缩后保存，仅本机可见。</p>

    <h3>按钮颜色</h3>
    <div class="custom">
      <input v-model="buttonColor" type="color" />
      <input v-model="buttonColor" type="text" />
      <button class="ghost" @click="applyAccent">应用</button>
    </div>

    <h3>翻译 API 密钥</h3>
    <div class="form">
      <label>有道应用 ID（appKey）<input v-model="settings.api_keys.youdao_app_key" /></label>
      <label>有道应用密钥（appSecret）<input v-model="settings.api_keys.youdao_app_secret" type="password" /></label>
      <label>百度 App ID<input v-model="settings.api_keys.baidu_app_id" /></label>
      <label>百度密钥<input v-model="settings.api_keys.baidu_secret" type="password" /></label>
      <label>DeepL Auth Key<input v-model="settings.api_keys.deepl_key" type="password" /></label>
    </div>
    <p class="hint">
      填写有道「文本翻译（NMT）」的应用 ID / 密钥后即时生效，翻译失败或留空时自动回退到免费源；百度 / DeepL 暂为预留。
    </p>
    <div class="test-row">
      <button class="ghost" :disabled="testing" @click="testTranslation">
        {{ testing ? '测试中…' : '测试有道翻译' }}
      </button>
      <span v-if="testResult" class="test-result" :class="testResult.ok ? 'ok' : 'bad'">
        <template v-if="testResult.ok">✅ 成功：{{ testResult.result }}</template>
        <template v-else>
          ❌ {{ testResult.error }}
          <span v-if="testResult.error_code">
            （错误码 {{ testResult.error_code }}：{{ youdaoErrMsg[testResult.error_code] || '请查有道文档' }}）
          </span>
        </template>
      </span>
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
.bg-preview {
  display: block;
  margin-top: 10px;
  max-width: 320px;
  max-height: 180px;
  border-radius: 14px;
  box-shadow: var(--shadow);
}
.opacity-row {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-top: 12px;
  font-size: 13px;
  max-width: 360px;
}
.opacity-row input[type='range'] {
  flex: 1;
  accent-color: var(--accent);
}
.opacity-val {
  width: 44px;
  text-align: right;
  color: var(--muted);
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
.test-row {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-top: 10px;
  flex-wrap: wrap;
}
.test-result {
  font-size: 12px;
  color: var(--muted);
}
.test-result.ok {
  color: #2e9e5b;
}
.test-result.bad {
  color: #d95d6a;
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
