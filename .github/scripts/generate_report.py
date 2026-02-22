import glob
import json
import os
from datetime import datetime
from typing import Any, Dict, List
from zoneinfo import ZoneInfo

# --- Configuration & Constants ---

ORDERED_SERVICES = [
    "naurffxiv",
    "moddingway",
    "authingway",
    "findingway",
    "clearingway",
]

CHECK_MAPPING = {
    "ESLint": "Lint",
    "Ruff Lint": "Lint",
    "Lint": "Lint",
    "GolangCI-Lint": "Lint",
    "Prettier": "Format",
    "Ruff Format": "Format",
    "Black": "Format",
    "Build": "Build",
    "TypeScript": "Build",
    "Type Check": "Build",
    "Unit Tests": "Unit Test",
    "E2E": "E2E",
}

COLUMN_ORDER = {
    "Format": 10,
    "Lint": 11,
    "Build": 20,
    "Unit Test": 30,
    "E2E": 40,
}

REPORT_COLUMNS = ["Lint", "Format", "Build", "Unit Test"]
if os.environ.get("RUN_E2E", "true").lower() == "true":
    REPORT_COLUMNS.append("E2E")

STATUS_MAP = {"failure": "❌", "warning": "⚠️", "skipped": "🚫", "success": "✅"}
STATUS_BADGE = {"failure": "❌ FAILED", "warning": "⚠️ WARNING"}
SERVICE_STATUS_ICON = {"Failed": "🔴", "Warning": "🟡", "Passed": "🟢"}


def _td(content: str, align: str = "center") -> str:
    return f'<td align="{align}">{content}</td>'


def _th(content: str, align: str = "center") -> str:
    return f'<th align="{align}">{content}</th>'


class CIReportGenerator:
    def __init__(self, root_dir: str = ".") -> None:
        self.root_dir = root_dir
        self.data: List[Dict[str, Any]] = []
        self.services: Dict[str, Dict[str, Dict[str, Any]]] = {}
        self.service_status: Dict[str, str] = {}

    def load_data(self, pattern: str = "**/report-*.json") -> None:
        search_path = os.path.join(self.root_dir, pattern)
        files = glob.glob(search_path, recursive=True)

        for f in files:
            try:
                self._process_report_file(f)
            except Exception as e:
                print(f"Error reading {f}: {e}")

        for entry in self.data:
            self.services.setdefault(entry["service"], {})[entry["check_name"]] = entry

        self._calculate_service_status()

    def _process_report_file(self, file_path: str) -> None:
        with open(file_path, "r", encoding="utf-8") as json_file:
            self.data.append(json.load(json_file))

    def _calculate_service_status(self) -> None:
        for svc, checks in self.services.items():
            status = "Passed"
            for entry in checks.values():
                if entry["status"] == "failure":
                    status = "Failed"
                    break
                if entry["status"] == "warning" and status != "Failed":
                    status = "Warning"
            self.service_status[svc] = status

    def _ordered_services(self) -> List[str]:
        return [s for s in ORDERED_SERVICES if s in self.services] + [
            s for s in self.services if s not in ORDERED_SERVICES and s != "global"
        ]

    def generate_markdown(self) -> str:
        parts = [self._render_header()]
        if "global" in self.services:
            parts.append(self._render_global_validation(self.services["global"]))
        parts += [
            self._render_service_table(),
            self._render_failure_details(),
            self._render_footer(),
        ]
        return "\n".join(parts)

    def _render_header(self) -> str:
        now = datetime.now(ZoneInfo("America/Los_Angeles")).strftime(
            "%A, %B %d, %Y at %I:%M %p %Z"
        )
        return f"# NAUR Ecosystem CI Report\n\n**Updated:** {now}\n"

    def _render_global_validation(self, global_checks: Dict[str, Any]) -> str:
        md = "### Global Validation\n\n"
        for name, entry in sorted(
            global_checks.items(), key=lambda x: x[1].get("order", 99)
        ):
            icon = STATUS_MAP.get(entry["status"], "✅")
            message = entry.get("message", "")
            line = (
                f"[Visit Preview]({message})"
                if name.startswith("Preview") and entry["status"] == "success"
                else message
            )
            md += f"- {icon} **{name}**: {line}\n"
        return md + "\n"

    def _render_service_table(self) -> str:
        headers = (
            _th("Service") + "".join(_th(col) for col in REPORT_COLUMNS) + _th("Status")
        )
        md = "### Service Validation Dashboard\n\n"
        md += f"<table>\n<thead>\n<tr>{headers}</tr>\n</thead>\n<tbody>\n"

        for svc in self._ordered_services():
            checks = self.services[svc]
            failure_order = min(
                (
                    e.get("order", 999)
                    for e in checks.values()
                    if e["status"] == "failure"
                ),
                default=float("inf"),
            )
            cells = _td(f"<b>{svc}</b>", align="left")

            for col in REPORT_COLUMNS:
                entry = next(
                    (e for n, e in checks.items() if CHECK_MAPPING.get(n) == col),
                    checks.get(col),
                )
                if entry:
                    cells += _td(STATUS_MAP.get(entry["status"], "✅"))
                else:
                    fallback = (
                        STATUS_MAP["skipped"]
                        if COLUMN_ORDER.get(col, 999) > failure_order
                        else "-"
                    )
                    cells += _td(fallback)

            status = self.service_status[svc]
            cells += _td(
                f"{SERVICE_STATUS_ICON.get(status, '🟢')} {status}", align="left"
            )
            md += f"<tr>{cells}</tr>\n"

        return md + "</tbody>\n</table>\n\n---\n"

    def _render_failure_details(self) -> str:
        md = "### Failure & Warning Details\n\n"
        found_issues = False

        for svc, checks in self.services.items():
            issues = [
                e for e in checks.values() if e["status"] in ("failure", "warning")
            ]
            if not issues:
                continue

            found_issues = True
            icon = SERVICE_STATUS_ICON.get(self.service_status[svc], "🟡")
            md += f"#### {icon} Service: {svc}\n\n"

            for issue in issues:
                badge = STATUS_BADGE.get(issue["status"], issue["status"].upper())
                log = issue.get("log_content", "No log captured.")
                md += f"<details>\n<summary><b>{badge} - {issue['check_name']}</b></summary>\n\n"
                md += f"```text\n{log}\n```\n\n</details>\n\n"

        return (
            md if found_issues else md + "✅ **No failures or warnings detected.**\n\n"
        )

    def _render_footer(self) -> str:
        server = os.environ.get("GITHUB_SERVER_URL", "https://github.com")
        repo = os.environ.get("GITHUB_REPOSITORY", "")
        run_id = os.environ.get("GITHUB_RUN_ID", "")
        run_link = (
            f" | [View Run]({server}/{repo}/actions/runs/{run_id})" if run_id else ""
        )
        return f"---\n###### Generated by **NAUR CI Bot**{run_link}\n"


if __name__ == "__main__":
    generator = CIReportGenerator()
    generator.load_data()
    report_content = generator.generate_markdown()

    with open("report.md", "w", encoding="utf-8") as f:
        f.write(report_content)

    print("Report generated: report.md")
