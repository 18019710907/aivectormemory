from aivectormemory.config import USER_SCOPE_DIR
MEMORIES_TABLE = """
CREATE TABLE IF NOT EXISTS memories (
    id TEXT PRIMARY KEY,
    content TEXT NOT NULL,
    tags TEXT NOT NULL DEFAULT '[]',
    scope TEXT NOT NULL DEFAULT 'project',
    project_dir TEXT NOT NULL DEFAULT '',
    session_id INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
)"""

VEC_MEMORIES_TABLE = """
CREATE VIRTUAL TABLE IF NOT EXISTS vec_memories USING vec0(
    id TEXT PRIMARY KEY,
    embedding FLOAT[384]
)"""

SESSION_STATE_TABLE = """
CREATE TABLE IF NOT EXISTS session_state (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_dir TEXT NOT NULL DEFAULT '',
    is_blocked INTEGER NOT NULL DEFAULT 0,
    block_reason TEXT NOT NULL DEFAULT '',
    next_step TEXT NOT NULL DEFAULT '',
    current_task TEXT NOT NULL DEFAULT '',
    progress TEXT NOT NULL DEFAULT '[]',
    recent_changes TEXT NOT NULL DEFAULT '[]',
    pending TEXT NOT NULL DEFAULT '[]',
    updated_at TEXT NOT NULL,
    UNIQUE(project_dir)
)"""

ISSUES_TABLE = """
CREATE TABLE IF NOT EXISTS issues (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_dir TEXT NOT NULL DEFAULT '',
    issue_number INTEGER NOT NULL,
    date TEXT NOT NULL,
    title TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    content TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
)"""

ISSUES_ARCHIVE_TABLE = """
CREATE TABLE IF NOT EXISTS issues_archive (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_dir TEXT NOT NULL DEFAULT '',
    issue_number INTEGER NOT NULL,
    date TEXT NOT NULL,
    title TEXT NOT NULL,
    content TEXT NOT NULL DEFAULT '',
    archived_at TEXT NOT NULL,
    created_at TEXT NOT NULL
)"""

INDEXES = [
    "CREATE INDEX IF NOT EXISTS idx_memories_project ON memories(project_dir)",
    "CREATE INDEX IF NOT EXISTS idx_memories_scope ON memories(scope)",
    "CREATE INDEX IF NOT EXISTS idx_issues_date ON issues(date)",
    "CREATE INDEX IF NOT EXISTS idx_issues_status ON issues(status)",
    "CREATE INDEX IF NOT EXISTS idx_issues_project ON issues(project_dir)",
    "CREATE INDEX IF NOT EXISTS idx_issues_archive_project ON issues_archive(project_dir)",
    "CREATE INDEX IF NOT EXISTS idx_issues_archive_date ON issues_archive(date)",
]

ALL_TABLES = [MEMORIES_TABLE, VEC_MEMORIES_TABLE, SESSION_STATE_TABLE, ISSUES_TABLE, ISSUES_ARCHIVE_TABLE]


def init_db(conn):
    for sql in ALL_TABLES:
        conn.execute(sql)
    # 迁移：旧版 memories 表可能缺少 project_dir 列
    cols = {row[1] for row in conn.execute("PRAGMA table_info(memories)").fetchall()}
    if "project_dir" not in cols:
        conn.execute("ALTER TABLE memories ADD COLUMN project_dir TEXT NOT NULL DEFAULT ''")
    # 迁移：旧版 issues 表中 archived 记录移到 issues_archive
    issue_cols = {row[1] for row in conn.execute("PRAGMA table_info(issues)").fetchall()}
    if "archive_content" in issue_cols:
        rows = conn.execute("SELECT * FROM issues WHERE status IN ('archived', 'migrated')").fetchall()
        for r in rows:
            conn.execute(
                "INSERT INTO issues_archive (project_dir, issue_number, date, title, content, archived_at, created_at) VALUES (?,?,?,?,?,?,?)",
                (r["project_dir"], r["issue_number"], r["date"], r["title"], r["content"], r["updated_at"], r["created_at"])
            )
            conn.execute("DELETE FROM issues WHERE id=?", (r["id"],))
        # 重建 issues 表去掉废弃字段
        conn.execute("CREATE TABLE IF NOT EXISTS issues_new (id INTEGER PRIMARY KEY AUTOINCREMENT, project_dir TEXT NOT NULL DEFAULT '', issue_number INTEGER NOT NULL, date TEXT NOT NULL, title TEXT NOT NULL, status TEXT NOT NULL DEFAULT 'pending', content TEXT NOT NULL DEFAULT '', created_at TEXT NOT NULL, updated_at TEXT NOT NULL)")
        conn.execute("INSERT INTO issues_new SELECT id, project_dir, issue_number, date, title, status, content, created_at, updated_at FROM issues")
        conn.execute("DROP TABLE issues")
        conn.execute("ALTER TABLE issues_new RENAME TO issues")
    for sql in INDEXES:
        conn.execute(sql)
    # 迁移：session_state 表加 last_session_id 字段
    state_cols = {row[1] for row in conn.execute("PRAGMA table_info(session_state)").fetchall()}
    if "last_session_id" not in state_cols:
        conn.execute("ALTER TABLE session_state ADD COLUMN last_session_id INTEGER NOT NULL DEFAULT 0")
    # 迁移：user scope 记忆的 project_dir 从空字符串改为 @user@
    conn.execute(
        "UPDATE memories SET project_dir=? WHERE project_dir='' AND scope='user'",
        (USER_SCOPE_DIR,)
    )
    conn.commit()
