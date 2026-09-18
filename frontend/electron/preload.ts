import { contextBridge, ipcRenderer } from 'electron'

contextBridge.exposeInMainWorld('api', {
  getServerPort: (): Promise<number> => ipcRenderer.invoke('get-server-port'),
  getDataDir: (): Promise<string> => ipcRenderer.invoke('get-data-dir'),
  openDataDir: (): Promise<string> => ipcRenderer.invoke('open-data-dir'),
})
