#!/usr/bin/env python3
"""NewAPI admin helpers for local bootstrap scripts.

Mirrors Go pkg/baseurl.Origin for HTTP origins; option merge for group/pricing.
"""

from __future__ import annotations

import json
import os
import sys
import urllib.request
from typing import Any, Dict, Optional
from urllib.parse import urlsplit, urlunsplit


def http_origin(raw: str) -> str:
    raw = (raw or "").strip()
    if not raw:
        raise SystemExit("empty url")
    parts = urlsplit(raw)
    if parts.scheme not in ("http", "https") or not parts.netloc:
        raise SystemExit(f"invalid url: {raw}")
    path = parts.path.rstrip("/")
    if path == "/v1":
        path = ""
    if path:
        raise SystemExit(f"base url must not include a path (got {parts.path!r})")
    return urlunsplit((parts.scheme, parts.netloc, "", "", ""))


def _client():
    base = os.environ.get("NEWAPI_URL", "http://localhost:3000").rstrip("/")
    token = os.environ["NEW_API_ADMIN_TOKEN"]
    user_id = os.environ.get("NEW_API_ADMIN_USER_ID", "1")

    def request(method: str, path: str, body: Optional[Dict[str, Any]] = None) -> dict:
        data = None
        headers = {
            "Authorization": f"Bearer {token}",
            "New-Api-User": user_id,
        }
        if body is not None:
            data = json.dumps(body).encode("utf-8")
            headers["Content-Type"] = "application/json"
        req = urllib.request.Request(f"{base}{path}", data=data, headers=headers, method=method)
        with urllib.request.urlopen(req) as resp:
            return json.load(resp)

    return request


def _load_options(request) -> Dict[str, str]:
    return {item["key"]: item.get("value", "") for item in request("GET", "/api/option/").get("data", [])}


def _put_updates(request, updates, already_msg: str) -> None:
    if not updates:
        print(already_msg)
        return
    for item in updates:
        request("PUT", "/api/option/", item)


def ensure_group(group: str, label: str) -> None:
    request = _client()
    options = _load_options(request)
    updates = []
    for key in ("UserUsableGroups", "GroupRatio"):
        raw = options.get(key, "") or "{}"
        try:
            data = json.loads(raw) if raw else {}
        except json.JSONDecodeError:
            data = {}
        if not isinstance(data, dict) or group in data:
            continue
        data[group] = label if key == "UserUsableGroups" else 1
        updates.append({"key": key, "value": json.dumps(data, ensure_ascii=False)})
    _put_updates(request, updates, f"NewAPI group already registered: {group}")
    if updates:
        print(f"Registered NewAPI group: {group}")


def ensure_model_pricing(model: str, model_ratio: float, completion_ratio: float) -> None:
    request = _client()
    options = _load_options(request)
    updates = []
    for key, value in (("ModelRatio", model_ratio), ("CompletionRatio", completion_ratio)):
        raw = options.get(key, "") or "{}"
        try:
            data = json.loads(raw) if raw else {}
        except json.JSONDecodeError:
            data = {}
        if not isinstance(data, dict):
            data = {}
        if data.get(model) == value:
            continue
        data[model] = value
        updates.append({"key": key, "value": json.dumps(data, ensure_ascii=False)})
    _put_updates(request, updates, f"NewAPI pricing already set for {model}")
    if updates:
        print(
            f"Configured NewAPI pricing for {model} "
            f"(ModelRatio={model_ratio}, CompletionRatio={completion_ratio})"
        )


def main(argv: list[str]) -> None:
    if len(argv) < 2:
        raise SystemExit("usage: newapi_admin.py <origin|ensure-group|ensure-model-pricing> ...")
    cmd = argv[1]
    if cmd == "origin":
        print(http_origin(argv[2] if len(argv) > 2 else ""))
        return
    if cmd == "ensure-group":
        ensure_group(argv[2], argv[3] if len(argv) > 3 else argv[2])
        return
    if cmd == "ensure-model-pricing":
        ensure_model_pricing(argv[2], float(argv[3]), float(argv[4]))
        return
    raise SystemExit(f"unknown command: {cmd}")


if __name__ == "__main__":
    main(sys.argv)
