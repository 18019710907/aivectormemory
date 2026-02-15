import json
from aivectormemory.config import DEDUP_THRESHOLD
from aivectormemory.db.memory_repo import MemoryRepo
from aivectormemory.errors import success_response

CATEGORY_TAG_MAP = {
    "decisions": "decision",
    "modifications": "modification",
    "pitfalls": "pitfall",
    "todos": "todo",
    "preferences": "preference",
}

CATEGORY_SCOPE_OVERRIDE = {
    "preferences": "user",
}


def handle_auto_save(args, *, cm, engine, session_id, **_):
    scope = args.get("scope", "project")
    repo = MemoryRepo(cm.conn, cm.project_dir)
    saved = []

    for category, tag in CATEGORY_TAG_MAP.items():
        items = args.get(category, [])
        if not isinstance(items, list):
            continue
        cat_scope = CATEGORY_SCOPE_OVERRIDE.get(category, scope)
        for item in items:
            if not item or not isinstance(item, str):
                continue
            item = item[:5000] if len(item) > 5000 else item
            embedding = engine.encode(item)
            tags = [tag] + args.get("extra_tags", [])
            result = repo.insert(item, tags, cat_scope, session_id, embedding, DEDUP_THRESHOLD)
            saved.append({"id": result["id"], "action": result["action"], "category": category})

    return json.dumps(success_response(saved=saved, count=len(saved)))
