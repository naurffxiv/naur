<#
.SYNOPSIS
  Shared helper library for monorepo scripts.

.DESCRIPTION
  Provides Resolve-ProjectRoot, unified logging, job helpers and safe push/pop helpers.
#>

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

# Centralized Color Theme
$Theme = @{
    Ok    = 'Green'
    Info  = 'Cyan'
    Warn  = 'Yellow'
    Error = 'Red'
    Debug = 'DarkGray'
}

function Resolve-ProjectRoot {
    <#
    .SYNOPSIS
      Return repository root (two levels above scripts folder).
    #>
    param()
    $scriptRoot = $PSScriptRoot
    if (-not $scriptRoot) { $scriptRoot = (Get-Location).Path }
    $root = Split-Path (Split-Path $scriptRoot -Parent) -Parent
    if ([string]::IsNullOrWhiteSpace($root)) { $root = (Resolve-Path .).Path }
    return $root
}

function Write-Log {
    <#
    .SYNOPSIS
      Unified logging for Info/Ok/Warn/Error/Debug.
    #>
    param(
        [Parameter(Mandatory)][ValidateSet('Info', 'Ok', 'Warn', 'Error', 'Debug')][string]$Level,
        [Parameter(Mandatory)][string]$Message
    )
    switch ($Level) {
        'Ok'    { Write-Host "  [+]   $Message" -ForegroundColor $Theme.Ok }
        'Info'  { Write-Host "  [i]   $Message" -ForegroundColor $Theme.Info }
        'Warn'  { Write-Host "  [!]   $Message" -ForegroundColor $Theme.Warn }
        'Error' { Write-Host "  [X]   $Message" -ForegroundColor $Theme.Error }
        'Debug' { Write-Host "  [DBG] $Message" -ForegroundColor $Theme.Debug }
    }
}

function Start-TaskJob {
    <#
    .SYNOPSIS
      Start a named background job that captures output and error state.
      Automatically injects Write-Log and $Theme into the job context.
    #>
    param(
        [Parameter(Mandatory)][string]$Name,
        [Parameter(Mandatory)][scriptblock]$ScriptBlock,
        [object[]]$ArgumentList = @()
    )

    # Serialize theme and Write-Log definition for injection
    $themeEntries = foreach ($k in $Theme.Keys) { "$k='$($Theme[$k])'" }
    $themeStr = "`$global:Theme = @{ $($themeEntries -join ';') }"
    $writeLogDef = (Get-Item function:Write-Log).Definition
    $writeLogStr = "function Write-Log { " + $writeLogDef + " }"

    $initScript = [scriptblock]::Create($themeStr + "`n" + $writeLogStr)

    return Start-Job -Name $Name -InitializationScript $initScript -ScriptBlock $ScriptBlock -ArgumentList $ArgumentList
}

function Wait-TaskJobs {
    <#
    .SYNOPSIS
      Wait for a collection of jobs and show a progress bar. Returns PSCustomObject[]: Name, Success, Output.
    .PARAMETER Jobs
    .PARAMETER RefreshMs
    #>
    param(
        [Parameter(Mandatory)][System.Collections.IEnumerable]$Jobs,
        [int]$RefreshMs = 200
    )

    $jobs = @($Jobs)
    $completed = @{}
    $results = @()

    while ($jobs | Where-Object { $_.State -eq 'Running' }) {
        $done = @($jobs | Where-Object { $_.State -ne 'Running' }).Count
        $percent = if ($jobs.Count -gt 0) { [int](($done / $jobs.Count) * 100) } else { 100 }
        Write-Progress -Activity "Processing tasks" -Status "Completed $done/$($jobs.Count)" -PercentComplete $percent

        $jobs | Where-Object { $_.State -ne 'Running' -and -not $completed.ContainsKey($_.Id) } | ForEach-Object {
            $completed[$_.Id] = $true
            $job = $_
            $childFailed = @($job.ChildJobs | Where-Object { 'Failed' -eq $_.JobStateInfo.State })
            $isFailed = ('Failed' -eq $job.State) -or ($childFailed.Count -gt 0)
            $out = Receive-Job $job -Keep -ErrorAction SilentlyContinue
            if ($isFailed) {
                Write-Log -Level Error -Message "$($job.Name) failed."
            }
            else {
                Write-Log -Level Ok -Message "$($job.Name) finished."
            }
            $results += [PSCustomObject]@{ Name = $job.Name; Success = -not $isFailed; Output = $out }
        }

        Start-Sleep -Milliseconds $RefreshMs
    }

    # Finalize any remaining outputs (jobs that completed exactly as loop ended)
    $jobs | Where-Object { -not $completed.ContainsKey($_.Id) } | ForEach-Object {
        $job = $_
        $childFailed = @($job.ChildJobs | Where-Object { 'Failed' -eq $_.JobStateInfo.State })
        $isFailed = ('Failed' -eq $job.State) -or ($childFailed.Count -gt 0)
        $out = Receive-Job $job -Keep -ErrorAction SilentlyContinue
        if ($isFailed) { Write-Log -Level Error -Message "$($job.Name) failed." } else { Write-Log -Level Ok -Message "$($job.Name) finished." }
        $results += [PSCustomObject]@{ Name = $job.Name; Success = -not $isFailed; Output = $out }
    }

    Write-Progress -Activity "Processing tasks" -Completed
    return $results
}

function Enter-LocationSafe {
    param([string]$Path)
    Push-Location $Path
    return $true
}

function Exit-LocationSafe {
    param()
    Pop-Location
}

function Remove-DirectoryRobust {
    <#
    .SYNOPSIS
      Uses robocopy /purge to fast-delete a directory.
    #>
    param([Parameter(Mandatory)][string]$Path)

    if (-not (Test-Path $Path)) { return }

    try {
        $temp = Join-Path $env:TEMP ([Guid]::NewGuid().ToString())
        New-Item -Path $temp -ItemType Directory -Force | Out-Null

        # Robocopy to empty 'mirrors' the empty state to the target, deleting everything
        robocopy $temp $Path /purge /quiet /njh /njs /nfl /ndl /np /MT:32 > $null

        Remove-Item -Path $Path -Force -Recurse -ErrorAction SilentlyContinue
        Remove-Item -Path $temp -Force -ErrorAction SilentlyContinue
    }
    catch {
        Write-Log -Level Warn -Message "Robust delete failed for $Path, falling back to Remove-Item."
        Remove-Item -Path $Path -Force -Recurse -ErrorAction SilentlyContinue
    }
}
