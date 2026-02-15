import json
from aivectormemory.config import DEDUP_THRESHOLD
from aivectormemory.db.memory_repo import MemoryRepo
from aivectormemory.errors import success_response


def handle_remember(args, *, cm, engine, session_id, **_):
    content = args.get("content")
    tags = args.get("tags", [])
    scope = args.get("scope", "project")

    if not content:
        raise ValueError("content is required")
    if not isinstance(tags, list):
        raise ValueError("tags must be a list")
    if len(content) > 5000:
        content = content[:5000]

    repo = MemoryRepo(cm.conn, cm.project_dir)
    embedding = engine.encode(content)
    result = repo.insert(content, tags, scope, session_id, embedding, DEDUP_THRESHOLD)

    return json.dumps(success_response(
        id=result["id"], action=result["action"],
        tags=tags, scope=scope
    ))
