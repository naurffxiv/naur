<#
.SYNOPSIS
  Stop background dev processes that reference the repository root.

.DESCRIPTION
  Finds running node/dotnet/python/go processes whose command line references the project root and stops them.
#>

[CmdletBinding()]
param()

. "$PSScriptRoot/_lib.ps1"

$ProjectRoot = Resolve-ProjectRoot

$killedCount = 0
Get-CimInstance Win32_Process | Where-Object {
    $_.CommandLine -and $_.CommandLine -like "*$ProjectRoot*" -and
    $_.Name -match '^(node|dotnet|python|go)\.exe$'
} | ForEach-Object {
    Write-Log -Level Info -Message "Killing $($_.Name) (PID: $($_.ProcessId))"
    try {
        Stop-Process -Id $_.ProcessId -Force -ErrorAction SilentlyContinue
        $killedCount++
    } catch {
        Write-Log -Level Warn -Message "Failed to stop PID $($_.ProcessId): $_"
    }
}

if ($killedCount -eq 0) {
    Write-Log -Level Warn -Message 'No development processes found'
} else {
    Write-Log -Level Ok -Message "Stopped $killedCount process(es)"
}
