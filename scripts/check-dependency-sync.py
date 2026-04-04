import argparse
import json
import re
import sys
import xml.etree.ElementTree as ET
from dataclasses import dataclass
from pathlib import Path
from typing import Dict, List, Optional

try:
    from ruamel.yaml import YAML
except ImportError:
    try:
        from ruamel_yaml import YAML  # type: ignore
    except ImportError:
        # Fallback for when ruamel.yaml is not available (only needed for --fix)
        YAML = None  # type: ignore

try:
    import yaml
except ImportError:
    print("Error: PyYAML required. Run: uv run --with pyyaml python scripts/check-dependency-sync.py")
    sys.exit(2)


@dataclass
class ValidationResult:
    service           : str
    package           : str
    message           : str
    expected          : str
    actual            : str
    is_system_error   : bool = False


@dataclass
class FixResult:
    service           : str
    package           : str
    old_version       : str
    new_version       : str


class DependencyValidator:
    """
    Validates and syncs package dependency versions across a monorepo.

    Compares versions in dependencies.yml against actual manifest files
    (package.json, pyproject.toml, go.mod, .csproj) and reports mismatches.
    """
    def __init__(self, repo_root: Path, args: argparse.Namespace):
        self.repo_root = repo_root.resolve()
        self.args = args
        self.results: List[ValidationResult] = []
        self.fixes: List[FixResult] = []
        self.config: Dict = {}

    def _get_path(self, rel_path: str) -> Optional[Path]:
        try:
            path = (self.repo_root / rel_path).resolve()
            if not path.is_relative_to(self.repo_root):
                # Prevent directory traversal attacks
                self._log_system_error(f"Path traversal detected: {rel_path}")
                return None
            return path
        except Exception as e:
            self._log_system_error(f"Invalid path {rel_path}: {e}")
            return None

    def _log_system_error(self, message: str):
        self.results.append(
            ValidationResult("SYS", "N/A", message, "N/A", "N/A", is_system_error=True)
        )

    def _log_mismatch(self, service: str, pkg: str, exp: str, act: str, msg: str):
        self.results.append(ValidationResult(service, pkg, msg, exp, act))

    @staticmethod
    def parse_node(path: Path) -> Dict[str, str]:
        data = json.loads(path.read_text(encoding="utf-8"))
        deps = {**data.get("dependencies", {}), **data.get("devDependencies", {})}
        return {
            # Strip common version prefixes (^, ~, >=, etc.) and 'workspace:' protocol
            k: re.sub(r"^(\^|~|>=|<=|>|<|v|workspace:)", "", str(v)).strip()
            for k, v in deps.items()
        }

    @staticmethod
    def parse_python(path: Path) -> Dict[str, str]:
        pkgs = {}
        content = path.read_text(encoding="utf-8")
        if path.suffix == ".toml":
            # Simple regex-based parser for pinned dependencies in pyproject.toml
            # Matches "package==version" or "package[extra]==version" inside quotes
            for line in content.splitlines():
                if m := re.search(r'["\']([a-zA-Z0-9\._-]+)(?:\[.+\])?\s*==\s*([a-zA-Z0-9\._\-+]+)["\']', line):
                    pkgs[m.group(1).lower()] = m.group(2)
        else:
            for line in content.splitlines():
                # Matches package names with optional extras (e.g., package[extra])
                # followed by exact version pins (==1.2.3)
                if m := re.match(
                    r"^([a-zA-Z0-9\._-]+)(?:\[.+\])?\s*==\s*([a-zA-Z0-9\._\-+]+)",
                    line.split("#")[0].strip(),
                ):
                    pkgs[m.group(1).lower()] = m.group(2)
        return pkgs

    @staticmethod
    def parse_go(path: Path) -> Dict[str, str]:
        pkgs = {}
        in_block = False
        for line in path.read_text(encoding="utf-8").splitlines():
            clean = line.split("//")[0].strip()
            # Skip empty lines and indirect dependencies (transitive deps we don't directly control)
            if not clean or "indirect" in line:
                continue
            # Parse both block-style 'require (...)' and single-line 'require pkg version'
            if clean.startswith("require ("):
                in_block = True
                continue
            if clean == ")" and in_block:
                in_block = False
                continue

            # Handle both single-line 'require x v1' and block entries 'x v1'
            tokens = clean.replace("require ", "").split() if not in_block else clean.split()
            if len(tokens) >= 2:
                 pkgs[tokens[0]] = tokens[1]
        return pkgs

    def parse_dotnet(self, path: Path) -> Dict[str, str]:
        pkgs = {}
        cpm_file = self.repo_root / "Directory.Packages.props"
        if cpm_file.exists():
            # Check Directory.Packages.props for Central Package Management (CPM)
            # CPM allows .NET projects to manage NuGet package versions centrally
            for p in ET.parse(cpm_file).findall(".//PackageVersion"):
                if name := p.get("Include"):
                    pkgs[name] = p.get("Version", "")

        for p in ET.parse(path).findall(".//PackageReference"):
            if name := p.get("Include"):
                # Use local version if present, otherwise fallback to CPM or existing
                pkgs[name] = p.get("Version") or pkgs.get(name, "")
        return pkgs

    def run(self) -> bool:
        deps_path = self.repo_root / "dependencies.yml"
        if not deps_path.exists():
            self._log_system_error("dependencies.yml not found")
            return False

        # Load config - prefer ruamel for preserving comments during fixes
        if self.args.fix and YAML:
            y = YAML()

            with open(deps_path, encoding="utf-8") as f:
                self.config = y.load(f) or {}
        else:
            with open(deps_path, encoding="utf-8") as f:
                self.config = yaml.safe_load(f) or {}

        parsers = {
            "node"    : self.parse_node,
            "python"  : self.parse_python,
            "go"      : self.parse_go,
            "dotnet"  : self.parse_dotnet,
        }

        svc_updates = {}

        for name, cfg in self.config.get("services", {}).items():
            if self.args.audit_only and name != self.args.audit_only:
                continue

            manifest = self._get_path(cfg["manifest_file"])
            if not manifest or not manifest.exists():
                self._log_system_error(f"Manifest not found for {name}: {cfg['manifest_file']}")
                continue

            try:
                parser = parsers.get(cfg["runtime"])
                if not parser:
                    continue

                actual_deps = parser(manifest)
                svc_updates[name] = self._audit_service(name, cfg.get("audited_packages", []), actual_deps)
            except Exception as e:
                self._log_system_error(f"Failed to parse {name}: {e}")

        if self.args.fix:
            return self._apply_fixes(deps_path, svc_updates)

        # Returns True if validation passed (no non-system errors found)
        # or if running in --fix mode (fixes always return True)
        return not any(not r.is_system_error for r in self.results)

    def _audit_service(self, service: str, audited_pkgs: List[Dict], actual_deps: Dict[str, str]) -> Dict[str, str]:
        updates = {}
        for pkg in audited_pkgs:
            name = pkg["name"]
            expected = str(pkg["version"])
            # Use "MISSING" string instead of None to distinguish between
            # "package not found" vs "version is null/empty"
            actual = actual_deps.get(name, "MISSING")

            if actual == "MISSING":
                self._log_mismatch(service, name, expected, actual, "Package not found")
            elif actual != expected:
                self._log_mismatch(service, name, expected, actual, "Version mismatch")

            updates[name] = actual
        return updates

    def _apply_fixes(self, deps_path: Path, updates_map: Dict[str, Dict[str, str]]) -> bool:
        if not YAML:
             print("Error: ruamel.yaml required for formatting fixes.")
             return False

        made_changes = False
        for s_name, updates in updates_map.items():
            service_cfg = self.config["services"][s_name]
            for pkg in service_cfg.get("audited_packages", []):
                name = pkg["name"]
                new_ver = updates.get(name)

                if new_ver and new_ver != "MISSING" and str(pkg["version"]) != new_ver:
                    self.fixes.append(FixResult(s_name, name, str(pkg["version"]), new_ver))
                    pkg["version"] = new_ver
                    made_changes = True

        if made_changes:
            y = YAML()
            # Configure YAML formatting to match common conventions:
            # 2-space mapping indent, 4-space sequence indent, 2-space offset
            y.indent(mapping=2, sequence=4, offset=2)
            y.preserve_quotes = True # Keep existing quote style
            with open(deps_path, "w", encoding="utf-8") as f:
                y.dump(self.config, f)

        return True

    def print_report(self):
        if self.args.json:
            report = {
                "success": not self.results, # Success = no validation errors
                "errors": [
                    {
                        "service"   : r.service,
                        "package"   : r.package,
                        "message"   : r.message,
                        "expected"  : r.expected,
                        "actual"    : r.actual,
                        "type": "system" if r.is_system_error else "validation"
                    }
                    for r in self.results
                ],
                "fixes": [
                    {
                        "svc": f.service,
                        "pkg": f.package,
                        "old": f.old_version,
                        "new": f.new_version
                    }
                    for f in self.fixes
                ]
            }
            print(json.dumps(report, indent=2))
            return

        # Console Output
        BLUE    = "\033[94m"
        GREEN   = "\033[92m"
        WARNING = "\033[93m"
        FAIL    = "\033[91m"
        END     = "\033[0m"

        for f in self.fixes:
            print(f"{GREEN}Fixed {f.service}/{f.package}: "
                  f"{FAIL}{f.old_version}{END} -> {GREEN}{f.new_version}{END}")

        for r in self.results:
            color = FAIL if not r.is_system_error else WARNING
            print(f"{color}[{r.service}] {r.package}: {r.message} "
                  f"(Exp: {r.expected}, Act: {r.actual}){END}")

        if self.results and not any(r.is_system_error for r in self.results):
             script_name = Path(__file__).name
             print(f"\n{WARNING}Tip: Run '{script_name} --fix' to update dependencies.yml{END}")

        if not self.results and not self.fixes:
            print(f"{GREEN}Sync Validation Passed!{END}")


def main():
    parser = argparse.ArgumentParser(description="Dependency Sync Validator")
    parser.add_argument("--fix",        action="store_true", help="Update dependencies.yml to match manifest versions")
    parser.add_argument("--json",       action="store_true", help="Output results as JSON")
    parser.add_argument("--audit-only",                      help="Audit only a specific service")
    args = parser.parse_args()

    validator = DependencyValidator(Path(__file__).parent.parent, args)
    success = validator.run()
    validator.print_report()

    sys.exit(0 if success else 1)


if __name__ == "__main__":
    main()
