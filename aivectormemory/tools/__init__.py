TOOL_DEFINITIONS = [
    {
        "name": "remember",
        "description": "存入一条记忆。支持用户级（跨项目）和项目级存储，自动去重（相似度>0.95则更新）。",
        "inputSchema": {
            "type": "object",
            "properties": {
                "content": {"type": "string", "description": "记忆内容，Markdown 格式"},
                "tags": {"type": "array", "items": {"type": "string"}, "description": "标签列表"},
                "scope": {"type": "string", "enum": ["user", "project"], "default": "project", "description": "作用域"}
            },
            "required": ["content", "tags"]
        }
    },
    {
        "name": "recall",
        "description": "语义搜索回忆记忆。通过向量相似度匹配，即使用词不同也能找到相关记忆。",
        "inputSchema": {
            "type": "object",
            "properties": {
                "query": {"type": "string", "description": "搜索内容（语义搜索，可选）"},
                "scope": {"type": "string", "enum": ["user", "project", "all"], "default": "all"},
                "tags": {"type": "array", "items": {"type": "string"}, "description": "按标签过滤（无 query 时走纯标签精确查询）"},
                "top_k": {"type": "integer", "default": 5, "description": "返回结果数量"}
            }
        }
    },
    {
        "name": "forget",
        "description": "删除一条或多条记忆。",
        "inputSchema": {
            "type": "object",
            "properties": {
                "memory_id": {"type": "string", "description": "单个记忆 ID"},
                "memory_ids": {"type": "array", "items": {"type": "string"}, "description": "多个记忆 ID"},
                "tags": {"type": "array", "items": {"type": "string"}, "description": "按标签批量删除，删除所有匹配标签的记忆"},
                "scope": {"type": "string", "enum": ["user", "project", "all"], "default": "all", "description": "配合 tags 使用，限定删除范围"}
            }
        }
    },
    {
        "name": "status",
        "description": "读取或更新会话状态（阻塞状态、当前任务、进度等）。不传 state 参数则读取，传则部分更新。",
        "inputSchema": {
            "type": "object",
            "properties": {
                "state": {
                    "type": "object",
                    "description": "要更新的字段（部分更新）",
                    "properties": {
                        "is_blocked": {"type": "boolean"},
                        "block_reason": {"type": "string"},
                        "next_step": {"type": "string"},
                        "current_task": {"type": "string"},
                        "progress": {"type": "array", "items": {"type": "string"}},
                        "recent_changes": {"type": "array", "items": {"type": "string"}},
                        "pending": {"type": "array", "items": {"type": "string"}}
                    }
                }
            }
        }
    },
    {
        "name": "track",
        "description": "问题跟踪：create/update/archive/list 四个 action。",
        "inputSchema": {
            "type": "object",
            "properties": {
                "action": {"type": "string", "enum": ["create", "update", "archive", "list"]},
                "title": {"type": "string", "description": "问题标题（create）"},
                "date": {"type": "string", "description": "日期 YYYY-MM-DD"},
                "issue_id": {"type": "integer", "description": "问题 ID（update/archive）"},
                "status": {"type": "string", "enum": ["pending", "in_progress", "completed"]},
                "content": {"type": "string", "description": "排查内容"},
                "include_archived": {"type": "boolean", "default": False, "description": "list 时是否包含已归档问题"}
            },
            "required": ["action"]
        }
    },
    {
        "name": "digest",
        "description": "提取待整理的记忆列表，按 session 范围和标签过滤，由 AI 端归纳总结。",
        "inputSchema": {
            "type": "object",
            "properties": {
                "scope": {"type": "string", "enum": ["user", "project", "all"], "default": "project"},
                "since_sessions": {"type": "integer", "default": 20, "description": "最近 N 次会话"},
                "tags": {"type": "array", "items": {"type": "string"}, "description": "按标签过滤"},
                "compress": {"type": "boolean", "default": False, "description": "是否触发智能归纳压缩，合并同主题碎片、清理过时记忆"}
            }
        }
    },
    {
        "name": "auto_save",
        "description": "【每次对话结束前必须调用】自动保存本次对话的关键信息。将决策、修改、踩坑、待办、偏好分类存储为独立记忆，自动打标签和去重。偏好类记忆固定 scope=user（跨项目通用）。",
        "inputSchema": {
            "type": "object",
            "properties": {
                "decisions": {"type": "array", "items": {"type": "string"}, "description": "本次对话做出的关键决策"},
                "modifications": {"type": "array", "items": {"type": "string"}, "description": "本次对话修改的文件和内容摘要"},
                "pitfalls": {"type": "array", "items": {"type": "string"}, "description": "本次对话遇到的坑和解决方案"},
                "todos": {"type": "array", "items": {"type": "string"}, "description": "本次对话产生的待办事项"},
                "preferences": {"type": "array", "items": {"type": "string"}, "description": "用户表达的技术偏好、设计风格倾向、架构选择习惯（固定 scope=user，跨项目通用）"},
                "scope": {"type": "string", "enum": ["user", "project"], "default": "project", "description": "作用域，默认项目级（preferences 固定 user）"},
                "extra_tags": {"type": "array", "items": {"type": "string"}, "description": "额外标签"}
            }
        }
    }
]

from aivectormemory.tools.remember import handle_remember
from aivectormemory.tools.recall import handle_recall
from aivectormemory.tools.forget import handle_forget
from aivectormemory.tools.status import handle_status
from aivectormemory.tools.track import handle_track
from aivectormemory.tools.digest import handle_digest
from aivectormemory.tools.auto_save import handle_auto_save

TOOL_HANDLERS = {
    "remember": handle_remember,
    "recall": handle_recall,
    "forget": handle_forget,
    "status": handle_status,
    "track": handle_track,
    "digest": handle_digest,
    "auto_save": handle_auto_save,
}
