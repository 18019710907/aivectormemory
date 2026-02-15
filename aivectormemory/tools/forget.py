import json
from aivectormemory.db.memory_repo import MemoryRepo
from aivectormemory.errors import success_response


def handle_forget(args, *, cm, **_):
    mid = args.get("memory_id")
    mids = args.get("memory_ids", [])
    tags = args.get("tags")
    scope = args.get("scope", "all")

    repo = MemoryRepo(cm.conn, cm.project_dir)

    if tags:
        rows = repo.list_by_tags(tags, scope=scope, project_dir=cm.project_dir, limit=10000)
        ids = [r["id"] for r in rows]
    else:
        ids = [mid] if mid else mids

    if not ids:
        raise ValueError("memory_id, memory_ids, or tags is required")

    deleted = [i for i in ids if repo.delete(i)]
    not_found = [i for i in ids if i not in deleted]
    return json.dumps(success_response(deleted_count=len(deleted), ids=deleted, not_found=not_found))
