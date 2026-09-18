/// <reference types="vite/client" />

declare module '*.vue' {
  import type { DefineComponent } from 'vue'
  const component: DefineComponent<{}, {}, any>
  export default component
}

interface Window {
  api?: {
    getServerPort(): Promise<number>
    getDataDir(): Promise<string>
    openDataDir(): Promise<string>
  }
}
