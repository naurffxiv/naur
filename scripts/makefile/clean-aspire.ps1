<#
.SYNOPSIS
  Remove .aspire cache folders created by the Aspire dashboard.
#>

[CmdletBinding()]
param()

. "$PSScriptRoot/_lib.ps1"

$ProjectRoot = Resolve-ProjectRoot

$aspireFolders = Get-ChildItem -Path $ProjectRoot -Filter .aspire -Recurse -Directory -ErrorAction SilentlyContinue |
    Where-Object { $_.FullName.StartsWith($ProjectRoot) }

if ($aspireFolders.Count -eq 0) {
    Write-Log -Level Warn -Message 'No Aspire cache found'
    exit 0
}

Write-Log -Level Info -Message "Found $($aspireFolders.Count) Aspire cache folder(s)..."

$aspireFolders | ForEach-Object -Parallel {
    $path = $_.FullName
    $root = $using:ProjectRoot
    $displayPath = $path.Replace($root, '.').Replace('\', '/')
    Write-Host "  Removing: $displayPath" -ForegroundColor Gray

    . "$using:PSScriptRoot/_lib.ps1"
    Remove-DirectoryRobust -Path $path
} -ThrottleLimit ([Environment]::ProcessorCount)

Write-Log -Level Ok -Message "Aspire cache removed"
