#!/usr/bin/env python3
import argparse
import json
import os
from pathlib import Path


def matching_sessions(workspace, sessions_root, include_subagents=False):
    workspace = Path(workspace).expanduser().resolve()
    sessions = []
    for path in Path(sessions_root).expanduser().glob("**/*.jsonl"):
        try:
            with path.open(encoding="utf-8") as source:
                first = source.readline(1024 * 1024)
            metadata = json.loads(first)
            if metadata.get("type") != "session_meta":
                continue
            payload = metadata.get("payload", {})
            cwd = payload.get("cwd")
            if not isinstance(cwd, str) or not Path(cwd).is_absolute():
                continue
            if Path(cwd).resolve() != workspace:
                continue
            origin = payload.get("source")
            if not include_subagents and (
                origin == "subagent" or isinstance(origin, dict) and "subagent" in origin
            ):
                continue
            sessions.append({
                "id": payload.get("id"),
                "workspace": str(workspace),
                "path": str(path.resolve()),
                "modified": path.stat().st_mtime,
            })
        except (OSError, UnicodeError, ValueError, TypeError, AttributeError):
            continue
    return sorted(sessions, key=lambda row: row["modified"], reverse=True)


def main():
    parser = argparse.ArgumentParser(description="List Codex session metadata for one exact workspace")
    parser.add_argument("--workspace", required=True)
    parser.add_argument("--sessions-root", default=str(
        Path(os.environ.get("CODEX_HOME") or Path.home() / ".codex") / "sessions"
    ))
    parser.add_argument("--include-subagents", action="store_true")
    parser.add_argument("--latest-activity", action="store_true")
    args = parser.parse_args()
    sessions = matching_sessions(args.workspace, args.sessions_root, args.include_subagents)
    if args.latest_activity:
        print(int(sessions[0]["modified"]) if sessions else 0)
    else:
        print(json.dumps(sessions, ensure_ascii=False, indent=2))


if __name__ == "__main__":
    main()
