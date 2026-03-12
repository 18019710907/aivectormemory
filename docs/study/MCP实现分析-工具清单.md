# AIVectorMemory MCP 实现分析（方法数量 & 功能清单）

本文面向本仓库的 **Python MCP Server**（`aivectormemory/` 包），梳理：

- **MCP 协议层**到底暴露了多少个 *JSON-RPC method*
- **tools（工具）层**一共有多少个 *tool*，分别做什么
- 每个工具的关键参数、主要数据流（落库 / 向量检索 / 状态聚合）

> 代码基准：`aivectormemory/server.py`、`aivectormemory/protocol.py`、`aivectormemory/tools/__init__.py` 及各工具实现。

---

## 1. MCP 协议层：4 个 JSON-RPC methods

在 `aivectormemory/server.py` 的 `MCPServer.run()` 中注册了 4 个 method handler：

- `**initialize`**  
  - **作用**：初始化 DB、递增并写入 `session_state.last_session_id`，并返回 MCP capabilities + serverInfo。  
  - **关键点**：每次初始化都会生成新的 `session_id`，用于写入 memories / issues / tasks / state 等记录的会话归因。
- `**notifications/initialized`**  
  - **作用**：当前实现为空（no-op），用于兼容 MCP 的初始化通知流程。
- `**tools/list`**  
  - **作用**：返回 `TOOL_DEFINITIONS`（工具元数据：name/description/inputSchema）。  
  - **数据来源**：`aivectormemory/tools/__init__.py`。
- `**tools/call`**  
  - **作用**：按 `params.name` 分发到 `TOOL_HANDLERS[name]` 执行，并把结果包装成 MCP `content: [{type:"text", text:"..."}]` 返回。  
  - **错误处理**：  
    - 找不到工具：`METHOD_NOT_FOUND`  
    - 参数问题：`INVALID_PARAMS`  
    - 其他异常：`SERVER_ERROR`
  - **输出截断**：对超大输出做 `_smart_truncate`（优先裁剪 JSON 结果中的列表字段：`memories/issues/tasks/results`）。

---

## 2. Tools 层：8 个 MCP Tools（“方法”）

`aivectormemory/tools/__init__.py` 中定义了 **8 个工具**（`TOOL_DEFINITIONS` + `TOOL_HANDLERS`）。

下面按 “工具名 → 功能 → 关键参数 → 主要实现/数据流” 梳理。

---

### 2.1 `remember`：写入记忆（含自动去重）

- **功能**：写入一条记忆到数据库，支持 `scope=project`（项目级）或 `scope=user`（用户级跨项目）。若与已有记忆向量相似度超过阈值（`DEDUP_THRESHOLD`，默认 0.95），会执行“更新/合并”而不是插入新行。
- **关键参数**
  - `content` *(required)*：记忆内容（最长截断到 5000）
  - `tags` *(required)*：标签（数组或逗号字符串）
  - `scope`：`user|project`（默认 `project`）
- **实现/数据流**
  - `EmbeddingEngine.encode(content)` 生成向量
  - 自动关键词：`extract_keywords(content)` 补充到 tags
  - `UserMemoryRepo.insert(...)` 或 `MemoryRepo.insert(...)`（带 `DEDUP_THRESHOLD`）

---

### 2.2 `recall`：语义检索（可按 scope/tags/source 过滤）

- **功能**：用向量相似度做语义检索；既可 “query + tags” 混合检索，也可仅用 tags 做分类浏览；还支持 `source=experience` 从归档的 track 经验里检索。
- **关键参数**
  - `query`：语义检索文本（可选；但 `query` 与 `tags` 至少要有一个）
  - `scope`：`user|project|all`（默认 `all`）
  - `tags`：标签过滤（数组或字符串）
  - `tags_mode`：`any|all`（默认：`query+tags→any`，仅 tags→`all`）
  - `top_k`：返回条数（默认 `DEFAULT_TOP_K`）
  - `source`：`manual|experience`（不传不过滤；`experience` 走经验检索逻辑）
  - `brief`：只返回 `content,tags`，省略元数据
- **实现/数据流**
  - `EmbeddingEngine.encode(query)` → repo 向量检索
  - **相似度**：把 `distance` 转换为 `similarity` 并排序（tags 检索与纯向量检索的换算略有差异）
  - `source=experience`：`IssueRepo.search_archive_by_vector()` 搜索归档 issue，拼装成 memories 形态返回

---

### 2.3 `forget`：删除记忆（按 id 或 tags 批量）

- **功能**：删除单条/多条记忆，或按 tags 批量删除。
- **关键参数**
  - `memory_id` / `memory_ids`：按 id 删除
  - `tags`：按标签批量删除（会先 list 再 delete）
  - `scope`：`user|project|all`（默认 `all`，决定删哪张表）
- **实现/数据流**
  - 项目记忆：`MemoryRepo.delete(id)`
  - 用户记忆：`UserMemoryRepo.delete(id)`
  - tags 批量：先 `list_by_tags(...limit=10000)` 拉 id，再逐个删

---

### 2.4 `status`：会话状态读写 + 进度聚合

- **功能**
  - **读取/更新** `session_state`（例如 `is_blocked/block_reason/current_task/pending` 等）
  - 自动计算 `progress`（只读）：聚合 **活跃 track** + **未完成 task feature** 的进度摘要
  - 支持 `clear_fields` 清空列表字段（绕过部分 IDE 对空数组的过滤）
- **关键参数**
  - `state`：不传则 read；传则 partial update（会忽略 `progress`）
  - `clear_fields`：如 `["pending"]`、`["recent_changes"]`
- **实现/数据流**
  - 状态表：`StateRepo.get()/upsert()`
  - 进度聚合：
    - `IssueRepo.list_by_date(brief=True)` → `[track #n] ...`
    - SQL 统计 `tasks` 表 `feature_id` 的 `done/total` → `[task fid] done/total completed`

---

### 2.5 `track`：问题追踪（create/update/archive/delete/list）

- **功能**：记录问题的生命周期（排查→根因→方案→自测→归档），并在 archive 时联动归档该 feature 的 tasks。
- **关键参数**
  - `action` *(required)*：`create|update|archive|delete|list`
  - `issue_id`：注意这里的 `issue_id` 实际是 **issue_number**（工具内部会先 `get_by_number()` 解析）
  - `title/content/date/parent_id`：create 用
  - `investigation/root_cause/solution/files_changed/test_result/...`：update 用
  - `brief/limit/status/date`：list 用
- **实现/数据流**
  - repo：`IssueRepo(cm.conn, cm.project_dir, engine=engine)`
  - `_resolve_issue()`：优先活跃表，其次归档表
  - `archive` 联动：
    - 若 issue 绑定 `feature_id` 且该 feature 没有剩余活跃 issue：`TaskRepo.archive_by_feature(feature_id)`

---

### 2.6 `task`：任务管理（batch_create/update/list/delete/archive）

- **功能**
  - 批量创建任务（支持 1 级 children）
  - 更新任务状态，并把变更同步到 `docs/specs/.../tasks.md`（checkbox 勾选）
  - feature 级归档
  - **联动 track**：某 feature 的任务状态变化，会同步更新同 feature 的 issue 状态
- **关键参数**
  - `action` *(required)*：`batch_create|update|list|delete|archive`
  - `feature_id`：batch_create/list/archive 必填
  - `tasks`：batch_create 任务数组
  - `task_id/status/title`：update/delete 用
- **实现/数据流**
  - repo：`TaskRepo`
  - `_sync_tasks_md()`：
    - 在 `_SPEC_DIRS` 多个候选 spec 根目录里查找 `.../<feature_id>/tasks.md`
    - 优先按 “编号前缀（如 5.1）” 匹配 checkbox 行，否则按标题精确匹配
  - 任务状态联动 issue：`IssueRepo.list_by_feature_id(feature_id)` → 对齐到 `repo.get_feature_status(feature_id)`

---

### 2.7 `readme`：从代码/配置生成 README（支持 diff）

- **功能**：读取 `pyproject.toml` + `TOOL_DEFINITIONS` 自动生成 README 内容（含工具清单/依赖），或对比当前 README 与生成版差异。
- **关键参数**
  - `action`：`generate|diff`（默认 `generate`）
  - `lang`：`en/zh-TW/ja/de/fr/es`
  - `sections`：可选指定 `header/tools/deps`
- **实现/数据流**
  - `_extract_tools()` 直接从 `TOOL_DEFINITIONS` 抽取参数 schema 生成工具章节
  - `diff` 会解析现有 README 中的工具标题行，输出 missing/extra 工具集合

---

### 2.8 `auto_save`：自动保存用户偏好（写入 user memories）

- **功能**：把用户偏好（preferences）作为用户级记忆写入，并自动去重；tags 固定含 `preference`，可追加 `extra_tags`。
- **关键参数**
  - `preferences`：数组或逗号字符串（为空则返回 “empty”）
  - `extra_tags`：额外标签（可选）
- **实现/数据流**
  - repo：`UserMemoryRepo.insert(...source="auto_save")`
  - 向量：`EmbeddingEngine.encode(item)`
  - 关键词：`extract_keywords(item)` 补 tags

---

## 3. 代码结构：协议层 / 工具层 / 存储层如何拼起来

- **协议层（stdio JSON-RPC）**
  - `aivectormemory/protocol.py`：`read_message()/write_message()`，以及 `make_result/make_error`
  - `aivectormemory/server.py`：`MCPServer` 负责 method 分发、工具调用、输出截断
- **工具层（tools）**
  - `aivectormemory/tools/__init__.py`：
    - `TOOL_DEFINITIONS`：供 `tools/list`
    - `TOOL_HANDLERS`：供 `tools/call`
  - 单工具文件：`aivectormemory/tools/*.py`（纯函数 handler）
- **存储/检索层（db + embedding）**
  - `ConnectionManager(project_dir)`：绑定项目隔离（`project_dir`）
  - `init_db()`：创建表 + migrations
  - `EmbeddingEngine`：统一编码器（向量写入与检索都依赖它）
  - `*Repo`：MemoryRepo/UserMemoryRepo/IssueRepo/TaskRepo/StateRepo 等

---

## 4. “keywords.py”为何在 tools 目录但不是 tool

`aivectormemory/tools/keywords.py` 存在，但它不是 MCP tool（没有出现在 `TOOL_DEFINITIONS` / `TOOL_HANDLERS`）。它是内部依赖，给 `remember/auto_save` 做关键词提取用。

---

## 5. 总结：你问的“有多少个方法”

- **协议层 JSON-RPC methods**：**4 个**  
`initialize`、`notifications/initialized`、`tools/list`、`tools/call`
- **对外 MCP tools（你在客户端直接调用的“方法”）**：**8 个**  
`remember`、`recall`、`forget`、`status`、`track`、`task`、`readme`、`auto_save`

