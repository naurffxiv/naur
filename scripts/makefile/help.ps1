<#
.SYNOPSIS
  Display help for common make targets (mapped to scripts).
#>

[CmdletBinding()]
param()

. "$PSScriptRoot/_lib.ps1"

Write-Host "`n--- NAUR Monorepo ---`n" -ForegroundColor Cyan

$commands       = @(
    @{ Group    = "Setup & Installation"; Target = "setup";         Desc = "Install system compilers (.NET 10, Node, Python, Go)" }
    @{ Group    = "Setup & Installation"; Target = "install";       Desc = "Parallel install of all service dependencies" }
    @{ Group    = "Setup & Installation"; Target = "check";         Desc = "Validate environment prerequisite versions" }

    @{ Group    = "Development";          Target = "lint";          Desc = "Run linters and format checks (Layer 0)" }
    @{ Group    = "Development";          Target = "build";         Desc = "Build and validate all services (Layer 1)" }
    @{ Group    = "Development";          Target = "format";        Desc = "Auto-fix linting and formatting issues" }

    @{ Group    = "Aspire";               Target = "aspire-run";    Desc = "Launch the full stack with .NET Aspire" }
    @{ Group    = "Aspire";               Target = "aspire-watch";  Desc = "Launch the full stack with .NET Aspire (Hot-Reload)" }

    @{ Group    = "Maintenance";          Target = "clean";         Desc = "Remove build artifacts" }
    @{ Group    = "Maintenance";          Target = "clean-all";     Desc = "Deep clean including Aspire dashboard cache" }
    @{ Group    = "Maintenance";          Target = "kill-dev";      Desc = "Force-stop all background dev processes" }

    @{ Group    = "Troubleshooting";      Target = "troubleshoot";  Desc = "Run health checks and dependency diagnostics" }
)

$groups = $commands | Group-Object Group
foreach ($g in $groups) {
    Write-Host " $($g.Name):" -ForegroundColor Yellow
    foreach ($cmd in $g.Group) {
        $t = $cmd.Target.PadRight(15)
        Write-Host "   make " -NoNewline -ForegroundColor Gray
        Write-Host $t -NoNewline -ForegroundColor White
        Write-Host " - $($cmd.Desc)" -ForegroundColor Gray
    }
    Write-Host ""
}

Write-Host "[TIP] Run 'make troubleshoot' if a build fails to see a health report." -ForegroundColor Cyan
