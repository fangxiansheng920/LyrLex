# LyrLex（歌词单词学习器）

个人本地学习工具：导入/搜索歌词，逐句自动翻译，点击歌词中的单词即可查看读音、词义并收藏（支持英文与日文）。

方案文档见 `技术方案.md`，**M1–M6 已全部完成**，当前版本 **0.1.3**。

## 目录结构

- `backend/` — Go 后端（Gin + GORM + SQLite + wire）
- `frontend/` — Electron + Vue 3 + Vite
- `data/` — 开发期本地数据（SQLite 与音频缓存，运行时自动创建）
- `技术方案.md` — 冻结的技术方案

## 后端开发

```powershell
cd backend
go mod tidy            # 首次拉取依赖
wire ./internal        # 重新生成 wire_gen.go（修改依赖装配后执行）
go run ./cmd/server    # 默认监听 http://127.0.0.1:18080
```

健康检查：`GET http://127.0.0.1:18080/healthz`

说明：

- 首次运行需要先安装 wire：`go install github.com/google/wire/cmd/wire@v0.7.0`
- 配置：`backend/configs/config.yaml`（非敏感）与 `backend/.env`（敏感，从 `.env.example` 复制）
- 数据目录默认 `../data/`（相对 backend 运行目录），可用环境变量 `LYRICS_DATA_DIR` 覆盖

## 前端开发

```powershell
cd frontend
npm install        # 首次安装依赖
npm run dev        # 仅 Vite 页面开发（浏览器访问）
npm run build      # 类型检查 + 打包页面 + 编译 Electron 主进程
npm start          # 构建后启动 Electron（自动拉起 Go 后端）
```

`npm start` 前需要先编译出 Go 侧车二进制（Electron 会去 `backend/lyrics-server.exe` 找）：

```powershell
cd backend
go build -o lyrics-server.exe ./cmd/server
```

> 提示：若 `npm install` 时 Electron 二进制下载失败（GitHub 连接超时），可先
> `$env:ELECTRON_SKIP_BINARY_DOWNLOAD="1"; npm install` 完成依赖安装，
> 再用镜像补下：`$env:ELECTRON_MIRROR="https://npmmirror.com/mirrors/electron/"; node node_modules/electron/install.js`

## M1 验收标准（已完成）

- Go 后端编译通过，`/healthz` 返回 `{"code":0,...}`
- Electron 启动后能拉起 Go 侧车进程
- 四个页面导航可用；设置中心可切换预设主题与自定义配色

## M2 验收标准（已完成）

- 粘贴/上传英文歌词 → 自动分句、逐句翻译、单词可点击
- 点击单词弹出 Popover：原词/基本形、音标（有则显示）、词义列表、发音、加入生词本
- 加入生词本后可在「生词本」页查看（页面 M4 完成，接口已就绪）
- 歌词可保存为歌曲（歌名/歌手可选），后续可回溯语境

## M3 验收标准（已完成）

- 日文歌词自动识别，kagome + UniDic 分词，不按空格切分
- 功能词合并：行ってしまった → 一个词组，基本形 行く
- 点词显示假名、罗马音、词义（Jisho，英文释义自动转中文）；助词/标点不可收藏
- 日文发音（有道 dictvoice `le=jap`）与日文生词本（JapaneseWord）

## M4 验收标准（已完成）

- 生词本英/日分区，搜索、详情、发音、删除
- 例句回溯：展示来源歌曲、原句、整句翻译
- 支持按歌曲过滤生词（`GET /api/vocabulary?song_id=`）

## M5 验收标准（已完成）

- 在线搜索歌名/歌手（NeteaseCloudMusicApi），选择后导入歌词并进入歌词学习
- 搜索服务不可用时自动降级：提示粘贴/上传歌词
- 本地歌曲库：已导入歌曲可再次打开

## M6 验收标准（已完成）

- 设置中心：翻译 API 配置（有道即时生效）、发音源、打开数据目录
- 设置中心外观：预设/自定义渐变/自定义纯色/按钮颜色/背景图片/背景透明度
- 歌词学习：重置按钮；模块切换用 KeepAlive 保留各页状态
- electron-builder 打包（NSIS 安装包 + portable 便携版），产品名 LyrLex

## 在线歌曲搜索（可选）

在线搜索依赖 [NeteaseCloudMusicApi](https://github.com/Binaryify/NeteaseCloudMusicApi) 服务，用 Docker 一行启动：

```powershell
docker run -d -p 3000:3000 binaryify/netease_cloud_music_api
```

默认地址 `http://localhost:3000`；若部署在其他地址，在 `backend/.env` 配置：

```
NETEASE_API_BASE=http://你的地址:端口
```

不启动该服务时，「歌曲搜索」会自动提示改为粘贴/上传，其余功能不受影响。

## 打包发布

```powershell
cd backend
go build -ldflags "-H=windowsgui" -o lyrics-server.exe ./cmd/server   # 无控制台窗口

cd ../frontend
npm run dist                                  # 构建 + electron-builder --win
```

产物在 `frontend/release/`：
- `LyrLex <版本>.exe` — 便携版，双击即用
- `LyrLex Setup <版本>.exe` — NSIS 安装包

版本号改 `frontend/package.json` 的 `version` 字段即可，打包时自动拼进文件名。

> 注意：若项目位于 OneDrive/Defender 同步目录（如桌面），打包时可能报
> `EPERM: rename win-unpacked.tmp`。把项目挪到非同步目录，或临时关闭同步/实时防护重试即可。

## 翻译源与有道 key（可选）

- 默认使用免费源 MyMemory 翻译，**不需要任何 key**。
- 想要更稳定/更准的翻译，可申请有道智云的「文本翻译（NMT）」，在「设置中心 → 翻译 API 密钥」填入**应用 ID / 应用密钥**，保存后即时生效；调用失败或额度用尽会自动回退免费源。
- 音标发音走有道**免费**接口，无需 key。

## 数据存放位置

打包版与 `npm start` 共用同一目录（Electron `userData`）：

```
%APPDATA%\lyrics-learner-frontend
```

- `lyrics.db` — 生词本、歌曲、歌词、翻译/查词缓存
- `settings.json` — 设置中心保存的 API key 等
- `audio\` — 发音音频缓存

应用内「设置中心 → 数据目录 → 打开目录」可直达；备份/迁移复制整个文件夹即可。

> 仅纯后端调试（`go run ./cmd/server`）时，数据写在项目的 `data\` 目录。
