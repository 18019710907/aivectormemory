# 设计稿落地实现 - 需求文档

## 1. 背景

设计稿（`docs/design/desktop/desktop-mockup/`）共 10 个 HTML 页面，定义了 AIVectorMemory 的完整 UI。需要将设计稿落地到两个已有项目中：

| 目标 | 技术栈 | 路径 |
|------|--------|------|
| **Desktop 桌面端** | Wails + Vue 3 + TypeScript | `desktop/frontend/src/` |
| **Web 看板** | Python Flask + 原生 JS + CSS | `aivectormemory/web/` |

## 2. 设计稿页面清单

| 页面 | 文件 | 功能 |
|------|------|------|
| 登录/注册 | auth.html | 用户认证、注册（本地存储） |
| 项目首页 | projects.html | 项目卡片列表、新增/删除项目 |
| 统计概览 | memory-stats.html | 4 列统计卡片 + 向量网络可视化 |
| 会话状态 | memory-session.html | 阻塞状态、活跃节点、脉冲指示器 |
| 问题跟踪 | memory-issues.html | 表格列表、搜索/筛选、状态徽章、分页 |
| 任务管理 | memory-tasks.html | 树形层级、勾选完成、进度条 |
| 项目记忆 | project-memories.html | 表格列表、标签、搜索、分页 |
| 全局记忆 | user-memories.html | 表格列表、标签、搜索、分页 |
| 标签管理 | tags.html | 标签分组、颜色变体、计数、增删改 |
| 设置 | settings.html | 配置项 + 维护工具（健康检查、重建向量、备份） |

## 3. 功能需求

### 3.1 认证系统（本地存储）

- **注册**：用户名 + 密码，存入本地 SQLite 数据库（不依赖远程服务）
- **登录**：本地验证，生成 session token
- **会话保持**：token 存 localStorage，过期后需重新登录
- **未登录限制**：未登录用户只能看到 auth 页面，无法访问 Memory 功能
- Desktop 端：通过 Wails Go 后端操作本地 SQLite
- Web 看板：通过 Flask API 操作同一个 SQLite 数据库

### 3.2 导航与布局

- **侧边栏**：固定 240px，Memory（7 项）+ System（设置）
- **顶栏**：52px，页面标题 + 项目选择器 + 操作按钮
- **主题切换**：暗色/亮色，localStorage 持久化
- **项目切换**：顶栏选择器，显示项目名称
- Desktop 端：macOS 风格拖拽栏 + 红绿灯
- Web 看板：标准浏览器布局

### 3.3 各页面功能

按设计稿 1:1 还原，具体交互参考 mockup HTML 文件。

### 3.4 数据源

两个端共用同一个 SQLite 数据库（`~/.aivectormemory/memory.db`），通过已有的 `MemoryManager` API 读写数据。

## 4. 范围界定

### 4.1 本次实现范围

- 10 个页面的 UI 还原（布局、样式、组件）
- 本地认证（注册/登录）
- 暗色/亮色主题
- 侧边栏导航 + 项目切换
- 数据展示（统计、记忆列表、问题、任务、标签）
- 基本 CRUD 操作（增删改查）
- 搜索与筛选
- 分页

### 4.2 不在本次范围

- 远程认证/OAuth
- 多用户协作
- 实时推送/WebSocket
- 移动端适配
- 国际化（沿用已有 i18n 体系）

## 5. 验收标准

1. 两个端（Desktop + Web 看板）均能完整展示 10 个页面
2. 本地注册/登录功能正常
3. 未登录不能访问 Memory 功能
4. 暗色/亮色主题切换正常，跨页面保持
5. 数据读写使用已有 MemoryManager，无需新建数据层
6. UI 样式与设计稿一致（布局、颜色、间距、组件）

## 6. 已有实现现状

### Desktop 端（Wails + Vue 3）
- 已有 Views：Stats、Status、Issues、Tasks、Memories、Tags、Settings、Maintenance、ProjectSelect
- 已有组件体系：stores、composables、router、i18n
- 缺少：auth 页面、projects 首页、设计稿新样式

### Web 看板（Flask + 原生 JS）
- 已有路由：projects、memories、issues、tasks、tags
- 已有静态文件：index.html + style.css + app.js + i18n.js
- 单页应用，原生 JS 渲染
- 缺少：auth 页面、会话状态页、设计稿新样式
