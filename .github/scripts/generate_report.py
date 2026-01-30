import collections
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
    "Preview": "Preview",
}

REPORT_COLUMNS = ["Lint", "Format", "Build", "Unit Test"]
if os.environ.get("RUN_E2E", "true").lower() == "true":
    REPORT_COLUMNS.append("E2E")

STATUS_MAP            = {"failure": "❌", "warning": "⚠️", "skipped": "🚫"}
SERVICE_STATUS_ICON   = {"Failed": "🔴", "Warning": "🟡", "Passed": "🟢"}


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
            entry = json.load(json_file)

            # If the log content is already in the JSON, use it.
            # Otherwise, try to read it from the file path if provided.
            if entry.get("status") != "success" and not entry.get("log_content"):
                if entry.get("log_file"):
                    entry["log_content"] = self._read_log_tail(entry["log_file"])

            self.data.append(entry)

    def _read_log_tail(self, log_path: str, max_lines: int = 100) -> str:
        if not log_path:
            return "[No log path provided]"

        final_path = log_path
        if not os.path.exists(final_path):
            basename = os.path.basename(log_path)
            found_logs = glob.glob(
                f"**/{basename}", root_dir=self.root_dir, recursive=True
            )
            if found_logs:
                final_path = os.path.join(self.root_dir, found_logs[0])

        if os.path.exists(final_path):
            try:
                stats = os.stat(final_path)
                if stats.st_size == 0:
                    return "[Log file is empty]"

                with open(final_path, "r", encoding="utf-8", errors="replace") as lf:
                    tail = collections.deque(lf, maxlen=max_lines)
                    return "".join(tail)
            except Exception as e:
                return f"[Error reading log file: {e}]"

        return f"[Log file not found: {log_path}]"

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

    def generate_markdown(self) -> str:
        parts = []
        parts.append(self._render_header())

        if "global" in self.services:
            parts.append(self._render_global_validation(self.services["global"]))

        parts.append(self._render_service_table())
        parts.append(self._render_failure_details())
        parts.append(self._render_footer())

        return "\n".join(parts)

    def _render_header(self) -> str:
        now = datetime.now(ZoneInfo("America/New_York")).strftime(
            "%A, %B %d, %Y at %I:%M %p EST"
        )
        return f"# NAUR Ecosystem CI Report\n\n**Updated:** {now}\n"

    def _render_global_validation(self, global_checks: Dict[str, Any]) -> str:
        md = "### Global Validation\n\n"
        # Sort global checks by their 'order' field
        sorted_checks = sorted(global_checks.items(), key=lambda x: x[1].get("order", 99))

        for name, entry in sorted_checks:
            icon = STATUS_MAP.get(entry["status"], "✅")
            message = entry.get("message", "")
            if name.startswith("Preview") and entry["status"] == "success":
                md += f"- {icon} **{name}**: [Visit Preview]({message})\n"
            else:
                md += f"- {icon} **{name}**: {message}\n"
        return md + "\n"

    def _render_service_table(self) -> str:
        md = "### Service Validation Dashboard\n\n"
        md += "| Service | " + " | ".join(REPORT_COLUMNS) + " | Status |\n"
        md += "| :--- | " + " | ".join([":---:"] * len(REPORT_COLUMNS)) + " | :--- |\n"

        all_services = [s for s in ORDERED_SERVICES if s in self.services] + [
            s for s in self.services if s not in ORDERED_SERVICES and s != "global"
        ]

        for svc in all_services:
            checks = self.services[svc]
            row = f"| **{svc}** |"

            for col in REPORT_COLUMNS:
                entry = next(
                    (e for n, e in checks.items() if CHECK_MAPPING.get(n) == col),
                    checks.get(col),
                )

                if entry:
                    icon = STATUS_MAP.get(entry["status"], "✅")
                    row += f" {icon} |"
                else:
                    row += " - |"

            status = self.service_status[svc]
            status_icon = SERVICE_STATUS_ICON.get(status, "🟢")

            row += f" {status_icon} {status} |"
            md += row + "\n"

        return md + "\n---\n"

    def _render_failure_details(self) -> str:
        md = "### Failure & Warning Details\n\n"
        found_issues = False

        for svc, checks in self.services.items():
            issues = [
                e for e in checks.values() if e["status"] in ["failure", "warning"]
            ]

            if issues:
                found_issues = True
                icon = "🔴" if self.service_status[svc] == "Failed" else "🟡"
                md += f"#### {icon} Service: {svc}\n\n"

                for issue in issues:
                    name = issue["check_name"]
                    log = issue.get("log_content", "No log captured.")
                    status_badge = "❌ FAILED" if issue["status"] == "failure" else "⚠️ WARNING"
                    md += f"<details>\n<summary><b>{status_badge} - {name}</b></summary>\n\n"
                    md += f"```text\n{log}\n```\n\n</details>\n\n"

        if not found_issues:
            return md + "✅ **No failures or warnings detected.**\n\n"

        return md

    def _render_footer(self) -> str:
        github_server_url = os.environ.get("GITHUB_SERVER_URL", "https://github.com")
        github_repository = os.environ.get("GITHUB_REPOSITORY", "")
        github_run_id = os.environ.get("GITHUB_RUN_ID", "")

        run_link = (
            f"[View Run]({github_server_url}/{github_repository}/actions/runs/{github_run_id})"
            if github_run_id
            else ""
        )

        footer = "---\n"
        footer += "###### Generated by **NAUR CI Bot**"
        if run_link:
            footer += f" | {run_link}"
        footer += "\n"

        return footer


if __name__ == "__main__":
    generator = CIReportGenerator()
    generator.load_data()
    report_content = generator.generate_markdown()

    with open("report.md", "w", encoding="utf-8") as f:
        f.write(report_content)

    print("Report generated: report.md")
