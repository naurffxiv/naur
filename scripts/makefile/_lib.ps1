<#
.SYNOPSIS
  Shared helper library for monorepo scripts.

.DESCRIPTION
  Provides Resolve-ProjectRoot, unified logging, job helpers and safe push/pop helpers.
#>

$ErrorActionPreference = 'Stop'

# Centralized Color Theme
$Theme = @{
    Ok       = 'Green'
    Info     = 'Cyan'
    Warn     = 'Yellow'
    Error    = 'Red'
    Debug    = 'DarkGray'
    Primary  = 'White'
    Emphasis = 'White'
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

function Test-ToolPresence {
    <#
    .SYNOPSIS
      Check if a tool is present on the system.
    .DESCRIPTION
      Accepts a tool hashtable. If it contains 'Check' and it's a scriptblock, invokes it.
      Otherwise, uses Get-Command to check for the presence of tool.Cmd.
      Optionally supports specialized 'Type' checks via Test-ServiceHealth.
    #>
    param(
        [Parameter(Mandatory)]
        [hashtable]$Tool
    )

    if ($Tool.Type) {
        # Check specialized types using the shared library
        return Test-ServiceHealth -Type $Tool.Type -Path "." # Path is often a dummy or root context
    }

    if ($Tool.ContainsKey('Check') -and $Tool.Check -is [scriptblock]) {
        return & $Tool.Check
    }

    if ($Tool.Cmd) {
        return [bool](Get-Command $Tool.Cmd -ErrorAction SilentlyContinue)
    }

    return $false
}

function Test-ServiceHealth {
    <#
    .SYNOPSIS
      Centralized check for service readiness.
    #>
    param(
        [Parameter(Mandatory)][string]$Type,
        [Parameter(Mandatory)][string]$Path,
        [string]$ProjectRoot,
        [string]$DotNetProjName
    )

    if (-not (Test-Path $Path)) { return $false }

    switch ($Type) {
        "Node" { return Test-Path (Join-Path $Path "node_modules") }
        "Go" { return Test-Path (Join-Path $Path "go.mod") }
        "Playwright" {
            $pwPath = Join-Path $env:LOCALAPPDATA "ms-playwright"
            return (Test-Path $pwPath) -and (Get-ChildItem $pwPath -Filter "chromium-*" -ErrorAction SilentlyContinue)
        }
        "Python" {
            $venvPython = Join-Path $Path "venv\Scripts\python.exe"
            if (-not (Test-Path $venvPython)) { $venvPython = Join-Path $Path "venv/bin/python" }
            if (Test-Path $venvPython) {
                # Heuristic: Check if pip can list anything (means venv is actually working)
                $installed = & $venvPython -m pip list --format=freeze 2>$null
                return ($null -ne $installed) -and (@($installed).Count -gt 0)
            }
            return $false
        }
        "DotNet" {
            $root = if ($ProjectRoot) { $ProjectRoot } else { Resolve-ProjectRoot }
            $projName = if ($DotNetProjName) { $DotNetProjName } else { [System.IO.Path]::GetFileNameWithoutExtension($Path) }
            $artifactsObjPath = Join-Path $root "artifacts/obj/$projName"
            return (Test-Path $artifactsObjPath) -and (Get-ChildItem $artifactsObjPath -ErrorAction SilentlyContinue)
        }
        Default { return $false }
    }
}

function Get-PythonPath {
    <#
    .SYNOPSIS
      Get the Python executable path from a venv directory.
    .PARAMETER Path
      Path to the service or venv directory.
    #>
    param([Parameter(Mandatory)][string]$Path)

    $windowsPath = Join-Path $Path "venv\Scripts\python.exe"
    if (Test-Path $windowsPath) { return $windowsPath }

    $unixPath = Join-Path $Path "venv/bin/python"
    if (Test-Path $unixPath) { return $unixPath }

    return $null
}
