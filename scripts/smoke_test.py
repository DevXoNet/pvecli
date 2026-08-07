#!/usr/bin/env python3
"""Safe smoke tests for pvecli.

The script executes every command with --help, but functionally executes only
an explicit read-only allowlist. It never starts, stops, reboots, migrates,
clones, resizes, updates, deletes, opens consoles, or executes guest commands.
"""

from __future__ import annotations

import argparse
import json
import re
import subprocess
import sys
import time
from dataclasses import asdict, dataclass
from pathlib import Path
from typing import Any

import yaml


HELP_COMMANDS = [
    [],
    ["agent"], ["agent", "ping"], ["agent", "network"], ["agent", "osinfo"],
    ["agent", "fsinfo"], ["agent", "exec"],
    ["backup"], ["backup", "list"], ["backup", "delete"],
    ["ceph"], ["ceph", "health"], ["ceph", "status"], ["ceph", "osd"], ["ceph", "pools"],
    ["cluster"], ["cluster", "update"],
    ["config"], ["config", "get"], ["config", "list"], ["config", "set"],
    ["console"], ["info"], ["list"],
    ["lxc"], ["lxc", "templates"],
    ["oci"], ["oci", "images"], ["oci", "tags"],
    ["reboot"], ["shutdown"], ["snapshot"], ["start"], ["status"], ["stop"],
    ["task"], ["task", "list"], ["task", "log"], ["task", "status"], ["task", "wait"],
    ["top"],
    ["vm"], ["vm", "clone"], ["vm", "disk"], ["vm", "disk", "resize"],
    ["vm", "migrate"], ["vm", "resume"], ["vm", "suspend"], ["vm", "template"],
]

SKIPPED_ACTIONS = {
    "agent exec": "executes a command inside a guest",
    "backup delete": "deletes backup data",
    "cluster update": "updates cluster nodes over SSH",
    "config": "interactive and may rewrite configuration",
    "config set": "changes persistent configuration",
    "console": "opens an interactive VM/LXC console",
    "reboot": "reboots a guest",
    "shutdown": "shuts down a guest",
    "snapshot": "creates a snapshot",
    "start": "starts a guest",
    "stop": "stops a guest",
    "task log/status/wait": "requires a specific task and may block",
    "top": "interactive long-running monitor",
    "vm clone": "creates a VM",
    "vm disk resize": "changes a VM disk",
    "vm migrate": "migrates a guest",
    "vm resume": "changes VM power state",
    "vm suspend": "changes VM power state",
    "vm template": "converts a VM to a template",
    "oci tags": "requires a registry/repository and performs an external registry request",
}

SECRET_PATTERNS = [
    re.compile(r'(?i)(token_secret|password|authorization)(["\s:=]+)([^\s",}]+)'),
    re.compile(r'(?i)(PVEAPIToken=)[^\s"]+'),
]


@dataclass
class Result:
    name: str
    command: list[str]
    status: str
    duration_seconds: float
    return_code: int | None = None
    message: str = ""
    output_excerpt: str = ""


def redact(text: str) -> str:
    for pattern in SECRET_PATTERNS:
        text = pattern.sub(lambda match: f"{match.group(1)}{match.group(2) if match.lastindex and match.lastindex >= 2 else ''}<redacted>", text)
    return text


def run_command(
    binary: Path,
    args: list[str],
    timeout: int,
    output_format: str | None = None,
    name: str | None = None,
) -> tuple[Result, Any | None]:
    command = [str(binary), *args]
    started = time.monotonic()
    try:
        process = subprocess.run(
            command,
            text=True,
            stdout=subprocess.PIPE,
            stderr=subprocess.STDOUT,
            timeout=timeout,
            check=False,
        )
    except subprocess.TimeoutExpired as error:
        raw_output = error.stdout or ""
        if isinstance(raw_output, bytes):
            raw_output = raw_output.decode(errors="replace")
        output = redact(raw_output)
        return Result(
            name=name or " ".join(args), command=command, status="failed",
            duration_seconds=round(time.monotonic() - started, 3),
            message=f"timed out after {timeout}s", output_excerpt=output[-2000:],
        ), None
    except OSError as error:
        return Result(
            name=name or " ".join(args), command=command, status="failed",
            duration_seconds=round(time.monotonic() - started, 3), message=str(error),
        ), None

    output = redact(process.stdout)
    parsed = None
    message = ""
    status = "passed" if process.returncode == 0 else "failed"
    stripped = process.stdout.strip()
    if status == "passed" and not stripped:
        status = "failed"
        message = "command returned empty output"
    elif status == "passed" and output_format == "json":
        try:
            parsed = json.loads(process.stdout)
        except json.JSONDecodeError as error:
            status = "failed"
            message = f"invalid JSON output: {error}"
    elif status == "passed" and output_format == "yaml":
        if stripped.startswith(("{", "[")):
            status = "failed"
            message = "requested YAML but command returned JSON"
        else:
            try:
                parsed = yaml.safe_load(process.stdout)
                if parsed is None:
                    raise ValueError("empty YAML document")
            except (yaml.YAMLError, ValueError) as error:
                status = "failed"
                message = f"invalid YAML output: {error}"
    elif status == "passed" and output_format == "text":
        if stripped.startswith(("{", "[")):
            status = "failed"
            message = "requested text but command returned structured JSON"

    return Result(
        name=name or ("root" if not args else " ".join(args)),
        command=command,
        status=status,
        duration_seconds=round(time.monotonic() - started, 3),
        return_code=process.returncode,
        message=message or ("" if process.returncode == 0 else f"exit code {process.returncode}"),
        output_excerpt=output[-2000:],
    ), parsed


def skip(name: str, reason: str) -> Result:
    return Result(name=name, command=[], status="skipped", duration_seconds=0, message=reason)


def discover_guest(data: Any, requested_vmid: str | None) -> tuple[str | None, str | None]:
    entries = data.get("vms", []) if isinstance(data, dict) else []
    if requested_vmid:
        for entry in entries:
            if str(entry.get("VMID")) == requested_vmid:
                return requested_vmid, "requested VMID"
        return requested_vmid, "requested VMID was not present in list output"

    for entry in entries:
        if entry.get("Type") == "vm" and entry.get("Status") == "running" and entry.get("GuestAgent"):
            return str(entry.get("VMID")), "auto-detected running VM with Guest Agent"
    for entry in entries:
        if entry.get("Status") == "running":
            return str(entry.get("VMID")), "auto-detected running guest"
    return None, None


def main() -> int:
    project_root = Path(__file__).resolve().parent.parent
    parser = argparse.ArgumentParser(description="Run non-destructive pvecli smoke tests")
    parser.add_argument("--binary", type=Path, default=project_root / "pvecli", help="Path to pvecli binary")
    parser.add_argument("--env", help="Configured pvecli environment to test")
    parser.add_argument("--vmid", help="VMID for read-only info/status/Guest Agent tests")
    parser.add_argument("--storage", help="Storage name for read-only backup list test")
    parser.add_argument("--timeout", type=int, default=30, help="Timeout per command in seconds")
    parser.add_argument("--report", type=Path, default=project_root / "smoke-report.json", help="JSON report path")
    parser.add_argument("--skip-agent", action="store_true", help="Skip Guest Agent read-only queries")
    options = parser.parse_args()

    binary = options.binary.resolve()
    if not binary.is_file():
        print(f"ERROR: binary not found: {binary}", file=sys.stderr)
        return 2

    env_args = ["--env", options.env] if options.env else []
    results: list[Result] = []

    print("[1/3] Checking command registration and help...")
    for path in HELP_COMMANDS:
        label = "root help" if not path else f"{' '.join(path)} help"
        result, _ = run_command(binary, [*path, "--help"], options.timeout, name=label)
        results.append(result)
        print(f"  {result.status.upper():7} {label}")

    print("[2/3] Running real read-only tests in text, JSON, and YAML...")
    formats = ("text", "json", "yaml")
    list_data = None

    # Configuration inspection is text-oriented and does not use output.Print.
    for name, args in [
        ("config list", ["config", "list"]),
        ("config get output_format", ["config", "get", "output_format"]),
    ]:
        result, _ = run_command(binary, [*env_args, *args], options.timeout, "text", name)
        results.append(result)
        print(f"  {result.status.upper():7} {name}")

    read_only_commands: list[tuple[str, list[str]]] = [
        ("cluster", ["cluster"]),
        ("list", ["list"]),
        ("task list", ["task", "list", "--limit", "5"]),
        ("ceph health", ["ceph", "health"]),
        ("ceph status", ["ceph", "status"]),
        ("ceph osd", ["ceph", "osd"]),
        ("ceph pools", ["ceph", "pools"]),
        ("lxc templates", ["lxc", "templates"]),
    ]

    for command_name, command_args in read_only_commands:
        for output_format in formats:
            name = f"{command_name} [{output_format}]"
            result, parsed = run_command(
                binary, [*env_args, *command_args, "-o", output_format],
                options.timeout, output_format, name,
            )
            results.append(result)
            if command_name == "list" and output_format == "json" and result.status == "passed":
                list_data = parsed
            print(f"  {result.status.upper():7} {name}")

    vmid, vmid_reason = discover_guest(list_data, options.vmid)
    if vmid:
        guest_commands: list[tuple[str, list[str]]] = [
            (f"info {vmid}", ["info", vmid]),
            (f"status {vmid}", ["status", vmid]),
        ]
        if options.skip_agent:
            results.append(skip("Guest Agent queries", "disabled by --skip-agent"))
        else:
            guest_commands.extend(
                (f"agent {action} {vmid}", ["agent", action, vmid])
                for action in ("ping", "network", "osinfo", "fsinfo")
            )

        for command_name, command_args in guest_commands:
            for output_format in formats:
                name = f"{command_name} [{output_format}]"
                result, _ = run_command(
                    binary, [*env_args, *command_args, "-o", output_format],
                    options.timeout, output_format, name,
                )
                results.append(result)
                print(f"  {result.status.upper():7} {name} ({vmid_reason})")
    else:
        results.append(skip("info/status/Guest Agent", "no running guest could be discovered"))

    if options.storage:
        for output_format in formats:
            name = f"backup list {options.storage} [{output_format}]"
            result, _ = run_command(
                binary,
                [*env_args, "backup", "list", "--storage", options.storage, "-o", output_format],
                options.timeout, output_format, name,
            )
            results.append(result)
            print(f"  {result.status.upper():7} {name}")
    else:
        results.append(skip("backup list", "provide --storage to test a specific storage"))

    print("[3/3] Recording intentionally skipped actions...")
    for name, reason in SKIPPED_ACTIONS.items():
        results.append(skip(name, reason))

    counts = {status: sum(result.status == status for result in results) for status in ("passed", "failed", "skipped")}
    report = {
        "binary": str(binary),
        "environment": options.env or "default",
        "safe_mode": True,
        "summary": counts,
        "results": [asdict(result) for result in results],
    }
    options.report.parent.mkdir(parents=True, exist_ok=True)
    options.report.write_text(json.dumps(report, indent=2) + "\n")

    print(f"\nSummary: {counts['passed']} passed, {counts['failed']} failed, {counts['skipped']} skipped")
    print(f"Report: {options.report}")
    if counts["failed"]:
        print("\nFailures:")
        for result in results:
            if result.status == "failed":
                print(f"  - {result.name}: {result.message}")
        return 1
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
