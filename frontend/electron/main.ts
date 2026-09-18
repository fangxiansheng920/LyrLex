import { app, BrowserWindow, ipcMain, Menu, nativeTheme, shell } from 'electron'
import { spawn, type ChildProcess } from 'node:child_process'
import fs from 'node:fs'
import http from 'node:http'
import net from 'node:net'
import path from 'node:path'

let serverProcess: ChildProcess | null = null
let serverPort = 18080

function findFreePort(): Promise<number> {
  return new Promise((resolve, reject) => {
    const srv = net.createServer()
    srv.listen(0, '127.0.0.1', () => {
      const address = srv.address()
      srv.close()
      if (address && typeof address === 'object') {
        resolve(address.port)
      } else {
        reject(new Error('无法获取空闲端口'))
      }
    })
    srv.on('error', reject)
  })
}

function checkHealth(port: number): Promise<boolean> {
  return new Promise((resolve) => {
    const req = http.get(
      { host: '127.0.0.1', port, path: '/healthz', timeout: 1000 },
      (res) => {
        res.resume()
        resolve(res.statusCode === 200)
      },
    )
    req.on('error', () => resolve(false))
    req.on('timeout', () => {
      req.destroy()
      resolve(false)
    })
  })
}

async function waitForHealth(port: number, timeoutMs = 20000): Promise<void> {
  const deadline = Date.now() + timeoutMs
  while (Date.now() < deadline) {
    if (await checkHealth(port)) return
    await new Promise((resolve) => setTimeout(resolve, 300))
  }
  throw new Error('lyrics-server 健康检查超时')
}

function resolveServerBin(): string {
  const envBin = process.env.LYRICS_SERVER_BIN
  if (envBin && fs.existsSync(envBin)) return envBin

  const candidates = [
    path.join(process.resourcesPath ?? '', 'lyrics-server.exe'),
    path.join(app.getAppPath(), 'backend', 'lyrics-server.exe'),
    path.join(app.getAppPath(), '..', 'backend', 'lyrics-server.exe'),
    path.join(app.getAppPath(), '..', '..', 'backend', 'lyrics-server.exe'),
  ]
  for (const candidate of candidates) {
    if (fs.existsSync(candidate)) return candidate
  }
  return ''
}

async function startServer(): Promise<void> {
  const bin = resolveServerBin()
  if (!bin) {
    console.warn('未找到 lyrics-server 可执行文件，假定后端已手动启动于 18080')
    serverPort = 18080
    return
  }

  serverPort = await findFreePort()
  serverProcess = spawn(bin, ['--port', String(serverPort)], {
    cwd: path.dirname(bin),
    env: { ...process.env, LYRICS_DATA_DIR: app.getPath('userData'), GIN_MODE: 'release' },
    stdio: 'ignore',
    windowsHide: true,
  })
  serverProcess.on('exit', (code) => {
    console.log(`lyrics-server exited: ${code}`)
    serverProcess = null
  })
  await waitForHealth(serverPort)
}

function createWindow(): void {
  const iconPath = path.join(__dirname, '..', 'build', 'icon.png')
  const win = new BrowserWindow({
    width: 1200,
    height: 800,
    backgroundColor: '#ffffff',
    icon: fs.existsSync(iconPath) ? iconPath : undefined,
    webPreferences: {
      preload: path.join(__dirname, 'preload.js'),
      contextIsolation: true,
      nodeIntegration: false,
    },
  })
  win.loadFile(path.join(__dirname, '..', 'dist', 'index.html'))
}

app.whenReady().then(async () => {
  // 标题栏强制浅色（白色），并去掉默认菜单
  nativeTheme.themeSource = 'light'
  Menu.setApplicationMenu(null)

  try {
    await startServer()
  } catch (err) {
    console.error('启动后端失败：', err)
  }

  ipcMain.handle('get-server-port', () => serverPort)
  ipcMain.handle('get-data-dir', () => app.getPath('userData'))
  ipcMain.handle('open-data-dir', async () => {
    const dir = app.getPath('userData')
    const err = await shell.openPath(dir)
    return err || dir
  })

  createWindow()
  app.on('activate', () => {
    if (BrowserWindow.getAllWindows().length === 0) createWindow()
  })
})

app.on('window-all-closed', () => {
  if (process.platform !== 'darwin') app.quit()
})

app.on('will-quit', () => {
  serverProcess?.kill()
})
