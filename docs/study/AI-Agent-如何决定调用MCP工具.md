# AI Agent 如何决定调用 MCP 工具（含流程图）

本文解释：当一个 AI agent（如 Cursor / Claude / Windsurf 内的 agent）连接到 MCP Server 后，它**为什么会“知道”该调用哪个工具（tool）**，以及这个过程在协议层是怎么发生的。

---

## 核心结论

AI agent 并不是“天生知道”你有哪些工具；它能做出正确调用，通常来自三类信息的叠加：

- **工具清单（Tool Catalog）**：客户端通过 MCP 的 `tools/list` 把工具的 `name/description/inputSchema` 提供给模型，等价于给了它一份 API 文档。
- **行为规则/提示词（Steering / Prompting）**：系统提示词或项目规则会约束“在什么场景必须调用什么工具”（例如先读状态、发现问题先建追踪、完成后归档等）。
- **意图匹配（Intent → Tool）**：模型根据用户话术与工具描述/参数 schema 的匹配度，选择最合适的工具并生成调用参数。

---

## MCP 层面发生了什么

在 AIVectorMemory 的 MCP 实现中（`aivectormemory/server.py`），协议层对外只暴露少量 JSON-RPC methods，其中与“工具”相关的是：

- **`tools/list`**：返回 `TOOL_DEFINITIONS`（工具元信息：name/description/inputSchema）
- **`tools/call`**：执行 `TOOL_HANDLERS[name]`，把结果包装成 MCP `content` 返回

所以：**模型想调用工具，前提是客户端已把 `tools/list` 的结果作为“可用工具集合”提供给模型上下文**。

---

## 为什么提示词很关键

即使模型“知道工具存在”，也未必会在正确时机调用。提示词/规则的作用在于：

- **触发时机**：例如“新会话必须先读 status / recall”，模型就会在收到消息时优先做初始化读取。
- **流程约束**：例如“提出方案后要阻塞等待确认”，模型会倾向在响应里更新 `status.is_blocked=true`。
- **产物规范**：例如“个人学习文档放到 `docs/study/`”，模型会把输出落到指定目录。

换句话说：**tools/list 解决“有什么能用”，提示词解决“什么时候用、怎么用才算对”。**

---

## 流程图（从用户输入到工具执行）

```mermaid
flowchart TD
  U[用户输入自然语言请求] --> A[Agent 解析意图/任务类型]

  A -->|需要外部能力/持久化/结构化输出| T0{是否已知可用工具?}
  T0 -->|否| L[调用 MCP: tools/list\n获取 TOOL_DEFINITIONS]
  L --> C[客户端把工具清单\n注入模型上下文]
  C --> T1[模型选择 tool + 生成 arguments]

  T0 -->|是| T1[模型选择 tool + 生成 arguments]

  T1 --> K[客户端发起 MCP: tools/call\n{name, arguments}]
  K --> S[MCP Server 分发到 TOOL_HANDLERS[name]\n执行业务逻辑]
  S --> DB[(DB/向量引擎/文件系统等)]
  S --> R[返回 result\ncontent: text/json]
  R --> M[模型读取工具返回结果]
  M --> O[输出给用户 / 或继续下一次工具调用]

  A -->|纯闲聊/无需外部能力| O
```

---

## 常见“看起来像自动，其实是规则 + 工具定义”场景

- **“归档一下”**  
  - 规则/流程：问题生命周期走到最后应归档  
  - 工具匹配：`track` 的 `archive` action 最贴合 “归档”

- **“查下当前是否阻塞/进度”**  
  - 工具匹配：`status`（read）直接返回 `is_blocked/progress/pending`

- **“把偏好记下来/以后默认这样做”**  
  - 工具匹配：`auto_save` / `remember(scope=user)`

---

## 你在本项目里能直接验证的点

- 工具是否“可被知道”：看 `tools/list` 返回的 `TOOL_DEFINITIONS`
- 工具是否“被正确调用”：看 `tools/call` 是否按 `TOOL_HANDLERS` 分发
- 调用是否“被规则触发”：看项目的 steering/规则文件是否明确规定了调用顺序与阻塞策略

