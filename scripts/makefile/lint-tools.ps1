<#
.SYNOPSIS
  Ensure linting tools are installed (Windows winget-based installer).
#>

[CmdletBinding()]
param()

. "$PSScriptRoot/_lib.ps1"

Write-Log -Level Info -Message "Syncing linting tools..."

$tools = @(
    @{ Name = "Ruff";          Cmd = "ruff";          Id = "astral-sh.ruff" }
    @{ Name = "golangci-lint"; Cmd = "golangci-lint"; Id = "GolangCI.golangci-lint" }
    @{ Name = "ty";            Cmd = "ty";            Id = "astral-sh.ty" }
)

$installTasks = @()

foreach ($tool in $tools) {
    if (-not (Get-Command $tool.Cmd -ErrorAction SilentlyContinue)) {
        Write-Log -Level Warn -Message "$($tool.Name) missing"
        $installTasks += $tool.Id
    } else {
        Write-Log -Level Ok -Message "$($tool.Name) is already installed."
    }
}

if ($installTasks.Count -gt 0) {
    if (-not (Get-Command winget -ErrorAction SilentlyContinue)) {
        Write-Log -Level Error -Message "winget not available; install tools manually"
        exit 1
    }

    foreach ($id in $installTasks) {
        Write-Log -Level Info -Message "Running: winget install $id"
        winget install $id --silent --accept-source-agreements --accept-package-agreements | Out-Null
    }

    Write-Log -Level Info -Message "Refreshing PATH..."
    $env:Path = [System.Environment]::GetEnvironmentVariable("Path", "Machine") + ";" + [System.Environment]::GetEnvironmentVariable("Path", "User")
    Write-Log -Level Ok -Message "Tools installed. You can now use them in this session!"
} else {
    Write-Log -Level Info -Message "All linting tools are ready."
}
