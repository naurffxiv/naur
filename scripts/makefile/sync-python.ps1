<#
.SYNOPSIS
  Sync Python dependencies for all Python services using uv.

.DESCRIPTION
  Reads all Python services from the service registry and runs uv sync
  in each one. Auto-installs uv if missing.
#>

[CmdletBinding()]
param(
    [switch]$Check
)
. "$PSScriptRoot/_config.ps1"
. "$PSScriptRoot/_lib.ps1"

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

function Invoke-PythonSync {
    <#
    .SYNOPSIS
      Core sync logic. Separated for testability.
    #>
    param(
        [string]$ProjectRoot,
        [hashtable]$Registry,
        [switch]$Check
    )

    $pythonServices = $Registry.GetEnumerator() | Where-Object { $_.Value.Type -eq "Python" }

    if (-not $pythonServices) {
        Write-Log -Level Warn -Message "No Python services found in registry."
        return
    }

    Initialize-Uv

    $syncArgs = @("sync", "--all-groups", "--link-mode=copy")
    if ($Check) {
        $syncArgs += "--frozen"
        Write-Log -Level Info -Message "Verifying dependencies match lockfile (frozen mode)..."
    } else {
        Write-Log -Level Info -Message "Synchronizing dependencies (auto-update mode)..."
    }

    foreach ($entry in $pythonServices) {
        $serviceName = $entry.Value.DisplayName
        $serviceDir  = Get-ServicePath -ServiceName $entry.Key -ProjectRoot $ProjectRoot

        if (-not (Test-Path $serviceDir)) {
            throw "Service directory not found: $serviceDir"
        }

        $pyprojectPath = Join-Path $serviceDir "pyproject.toml"
        if (-not (Test-Path $pyprojectPath)) {
            throw "pyproject.toml not found in $serviceDir"
        }

        Write-Log -Level Info -Message "Syncing $serviceName..."
        Enter-LocationSafe $serviceDir

        try {
            & uv @syncArgs
            if ($LASTEXITCODE -ne 0) {
                throw "uv sync failed for $serviceName with exit code $LASTEXITCODE"
            }
            Write-Log -Level Ok -Message "$serviceName dependencies synchronized successfully"
        } finally {
            Exit-LocationSafe
        }
    }

    Write-Log -Level Ok -Message "All Python services synced successfully"
}

try {
    $ProjectRoot = Resolve-ProjectRoot
    Write-Log -Level Debug -Message "Project root: $ProjectRoot"

    Invoke-PythonSync -ProjectRoot $ProjectRoot -Registry (Get-ServiceRegistry) -Check:$Check

} catch {
    Write-Log -Level Error -Message "Unexpected error during sync: $_"
    Write-Host "`nStack trace:" -ForegroundColor $Theme.Muted
    Write-Host $_.ScriptStackTrace -ForegroundColor $Theme.Muted
    exit 1
}
