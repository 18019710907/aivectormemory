import sys
from http.server import HTTPServer, SimpleHTTPRequestHandler
from pathlib import Path
from aivectormemory.db import ConnectionManager, init_db
from aivectormemory.web.api import handle_api_request

STATIC_DIR = Path(__file__).parent / "static"


class NoFQDNHTTPServer(HTTPServer):
    def server_bind(self):
        self.socket.bind(self.server_address)
        self.server_address = self.socket.getsockname()


class WebHandler(SimpleHTTPRequestHandler):
    cm = None
    auth_token = None

    def address_string(self):
        return self.client_address[0]

    def _check_auth(self):
        if not self.auth_token:
            return True
        from urllib.parse import urlparse, parse_qs
        params = parse_qs(urlparse(self.path).query)
        return params.get("token", [None])[0] == self.auth_token

    def do_GET(self):
        if self.path.startswith("/api/"):
            if not self._check_auth():
                self.send_error(403, "Forbidden: invalid token")
                return
            handle_api_request(self, self.cm)
        else:
            self._serve_static()

    def do_PUT(self):
        if self.path.startswith("/api/"):
            if not self._check_auth():
                self.send_error(403, "Forbidden: invalid token")
                return
            handle_api_request(self, self.cm)
        else:
            self.send_error(405)

    def do_DELETE(self):
        if self.path.startswith("/api/"):
            if not self._check_auth():
                self.send_error(403, "Forbidden: invalid token")
                return
            handle_api_request(self, self.cm)
        else:
            self.send_error(405)

    def do_POST(self):
        if self.path.startswith("/api/"):
            if not self._check_auth():
                self.send_error(403, "Forbidden: invalid token")
                return
            handle_api_request(self, self.cm)
        else:
            self.send_error(405)

    def _serve_static(self):
        path = self.path.split("?")[0].lstrip("/") or "index.html"
        file_path = STATIC_DIR / path
        if not file_path.exists() or not file_path.is_file():
            file_path = STATIC_DIR / "index.html"
        if not file_path.exists():
            self.send_error(404)
            return
        content = file_path.read_bytes()
        content_type = {
            ".html": "text/html; charset=utf-8",
            ".css": "text/css; charset=utf-8",
            ".js": "application/javascript; charset=utf-8",
            ".json": "application/json",
            ".svg": "image/svg+xml",
        }.get(file_path.suffix, "application/octet-stream")
        self.send_response(200)
        self.send_header("Content-Type", content_type)
        self.send_header("Content-Length", len(content))
        self.end_headers()
        self.wfile.write(content)

    def log_message(self, format, *args):
        print(f"[aivectormemory-web] {args[0]}", file=sys.stderr)


def run_web(project_dir: str | None = None, port: int = 9080, bind: str = "127.0.0.1", token: str | None = None):
    cm = ConnectionManager(project_dir=project_dir)
    init_db(cm.conn)

    try:
        from aivectormemory.embedding.engine import EmbeddingEngine
        engine = EmbeddingEngine()
        engine.load()
        cm._embedding_engine = engine
        print("[aivectormemory] Semantic search enabled", file=sys.stderr)
    except Exception as e:
        cm._embedding_engine = None
        print(f"[aivectormemory] Semantic search disabled: {e}", file=sys.stderr)

    WebHandler.cm = cm
    WebHandler.auth_token = token

    server = NoFQDNHTTPServer((bind, port), WebHandler)
    print(f"[aivectormemory] Web dashboard: http://{bind}:{port}", file=sys.stderr)
    if token:
        print(f"[aivectormemory] Token auth enabled", file=sys.stderr)
    try:
        server.serve_forever()
    except KeyboardInterrupt:
        pass
    finally:
        cm.close()
        server.server_close()
