export type ThemeColors = [string, string, string]

export interface Theme {
  name: string
  colors: ThemeColors
  accent: string // 按钮/高亮主色
}

export const DEFAULT_ACCENT = '#6db9e8'

export const presetThemes: Theme[] = [
  { name: '默认白', colors: ['#ffffff', '#ffffff', '#ffffff'], accent: DEFAULT_ACCENT },
  { name: '清新蓝', colors: ['#8ac4e2', '#b5dbe5', '#d3eed9'], accent: DEFAULT_ACCENT },
  { name: '珊瑚蓝', colors: ['#27a6cc', '#80bfd4', '#fcc5c5'], accent: DEFAULT_ACCENT },
  { name: '青柠', colors: ['#a8d5ba', '#d7e8bd', '#fff3c7'], accent: DEFAULT_ACCENT },
  { name: '暮紫', colors: ['#303d67', '#79799a', '#fddfdc'], accent: DEFAULT_ACCENT },
]

const STORAGE_KEY = 'lyrics-theme'

function hexToRgb(hex: string): { r: number; g: number; b: number } | null {
  const m = /^#?([0-9a-f]{6})$/i.exec(hex.trim())
  if (!m) return null
  const n = parseInt(m[1], 16)
  return { r: (n >> 16) & 255, g: (n >> 8) & 255, b: n & 255 }
}

function rgbToHex(r: number, g: number, b: number): string {
  const clamp = (v: number) => Math.max(0, Math.min(255, Math.round(v)))
  return (
    '#' +
    [clamp(r), clamp(g), clamp(b)]
      .map((v) => v.toString(16).padStart(2, '0'))
      .join('')
  )
}

function mix(a: string, b: string, t: number): string {
  const ca = hexToRgb(a)
  const cb = hexToRgb(b)
  if (!ca || !cb) return a
  return rgbToHex(
    ca.r + (cb.r - ca.r) * t,
    ca.g + (cb.g - ca.g) * t,
    ca.b + (cb.b - ca.b) * t,
  )
}

export function applyTheme(theme: Theme): void {
  const root = document.documentElement
  root.style.setProperty('--bg-a', theme.colors[0])
  root.style.setProperty('--bg-b', theme.colors[1])
  root.style.setProperty('--bg-c', theme.colors[2])

  const accent = theme.accent || DEFAULT_ACCENT
  const rgb = hexToRgb(accent)
  if (rgb) {
    const lighter = mix(accent, '#ffffff', 0.4)
    const deeper = mix(accent, '#000000', 0.18)
    root.style.setProperty('--accent', accent)
    root.style.setProperty('--accent-deep', deeper)
    root.style.setProperty('--accent-soft', `rgba(${rgb.r}, ${rgb.g}, ${rgb.b}, 0.18)`)
    root.style.setProperty('--grad', `linear-gradient(135deg, ${lighter} 0%, ${accent} 100%)`)
    root.style.setProperty(
      '--grad-violet',
      `linear-gradient(135deg, ${mix(accent, '#ffffff', 0.55)} 0%, ${lighter} 100%)`,
    )
  }
}

export function saveTheme(theme: Theme): void {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(theme))
  applyTheme(theme)
}

export function loadTheme(): Theme {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (raw) {
      const parsed = JSON.parse(raw) as Theme
      if (Array.isArray(parsed.colors) && parsed.colors.length === 3) {
        return {
          name: parsed.name || '自定义',
          colors: [parsed.colors[0], parsed.colors[1], parsed.colors[2]],
          accent: parsed.accent || DEFAULT_ACCENT,
        }
      }
    }
  } catch {
    /* ignore */
  }
  return presetThemes[0]
}
