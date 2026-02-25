<#
.SYNOPSIS
  Parallel install of monorepo dependencies.

.DESCRIPTION
  Restores .NET, downloads Go modules, runs package manager install at the repo root and prepares Python venvs.
#>

[CmdletBinding()]
param()

. "$PSScriptRoot/_lib.ps1"
. "$PSScriptRoot/_config.ps1"

$ProjectRoot = Resolve-ProjectRoot
$pmInstall = Get-PackageManagerCmd -CommandType "InstallCmd"
$pmExec = Get-PackageManagerCmd -CommandType "ExecCmd"

$Tasks = @(
    @{ Name = "AppHost";     Action = { param($p) dotnet restore $p --verbosity quiet }; Arg = Get-ServicePath -ServiceName "AppHost" -ProjectRoot $ProjectRoot }
    @{ Name = "Authingway";  Action = { param($p) dotnet restore $p --verbosity quiet }; Arg = Get-ServicePath -ServiceName "Authingway" -ProjectRoot $ProjectRoot }
    @{ Name = "Root-pnpm";   Action = { param($p, $cmd) Set-Location $p; Invoke-Expression $cmd }; Arg = @($ProjectRoot, $pmInstall) }
    @{ Name = "Findingway";  Action = { param($p) Set-Location $p; go mod download }; Arg = Get-ServicePath -ServiceName "Findingway" -ProjectRoot $ProjectRoot }
    @{ Name = "Clearingway"; Action = { param($p) Set-Location $p; go mod download }; Arg = Get-ServicePath -ServiceName "Clearingway" -ProjectRoot $ProjectRoot }
    @{ Name = "Moddingway";  Action = {
        param($p)
        Push-Location $p
        try {
            $venvPython = Join-Path $p "venv\Scripts\python.exe"
            if (-not (Test-Path $venvPython)) {
                Write-Log -Level Info -Message "Venv missing in $p, recreating..."
                if (Test-Path venv) { Remove-Item -Recurse -Force venv }
                python -m venv venv
                $venvPython = Join-Path $p "venv\Scripts\python.exe"
            }
            & $venvPython -m pip install -q --upgrade pip
            if (Test-Path requirements.txt) { & $venvPython -m pip install -q -r requirements.txt }
            if (Test-Path requirements-dev.txt) { & $venvPython -m pip install -q -r requirements-dev.txt }
        } finally {
            Pop-Location
        }
    }; Arg = Get-ServicePath -ServiceName "Moddingway" -ProjectRoot $ProjectRoot }
    @{ Name = "Playwright";  Action = {
        param($p, $cmd)
        Set-Location $p
        if (Test-Path "package.json") {
            Invoke-Expression "$cmd playwright install chromium --with-deps"
        }
    }; Arg = @((Get-ServicePath -ServiceName "E2ETests" -ProjectRoot $ProjectRoot), $pmExec) }
)

$jobs = foreach ($t in $Tasks) {
    $taskArgs = if ($t.Arg -is [array]) { $t.Arg } else { @($t.Arg) }
    Start-TaskJob -Name $t.Name -ScriptBlock $t.Action -ArgumentList $taskArgs
}

$results = Wait-TaskJobs -Jobs $jobs

$successCount = ($results | Where-Object { $_.Success }).Count
Write-Host "($successCount) services installation finished." -ForegroundColor $Theme.Ok

$failed = $results | Where-Object { -not $_.Success }
if ($failed) {
    Write-Log -Level Error -Message "One or more install tasks failed."
    foreach ($f in $failed) {
        Write-Host "`n--- Output for $($f.Name) ---" -ForegroundColor $Theme.Warn
        $f.Output | ForEach-Object { Write-Host $_ }
    }
    exit 1
}

# Verify artifacts
Write-Log -Level Info -Message "Verifying Service Artifacts"

$registry = Get-ServiceRegistry
$services = foreach ($key in $registry.Keys) {
    $svc = $registry[$key]
    @{
        Name = $svc.DisplayName
        Path = Get-ServicePath -ServiceName $key -ProjectRoot $ProjectRoot
        Type = $svc.Type
    }
}

$failedChecks = @()
foreach ($s in $services) {
    if (-not (Test-Path $s.Path)) {
        Write-Log -Level Error -Message "$($s.Name): Directory not found"
        $failedChecks += $s.Name; continue
    }

    $isInstalled = Test-ServiceHealth -Type $s.Type -Path $s.Path -ProjectRoot $ProjectRoot

    if ($isInstalled) { Write-Log -Level Ok -Message "$($s.Name) ($($s.Type)) ready" } else { Write-Log -Level Warn -Message "$($s.Name): Verification failed"; $failedChecks += $s.Name }
}

# Pre-commit hooks
$pmName = Get-PMName
$prekCheckFailed = $false
try {
    & $pmName prek --version > $null 2>&1
    if ($LASTEXITCODE -ne 0) {
        $prekCheckFailed = $true
    }
} catch {
    $prekCheckFailed = $true
}

if ($prekCheckFailed) {
    Write-Log -Level Error -Message "prek: Validation failed. Run '$pmName install' manually."
    $failedChecks += "prek"
} else {
    Write-Log -Level Ok -Message "prek (pre-commit hooks) installed"
}
if ($failedChecks.Count -gt 0) {
    Write-Log -Level Error -Message "Monorepo setup is incomplete."
    exit 1
}

Write-Host "`n  [SUCCESS] Monorepo is fully installed and ready!" -ForegroundColor $Theme.Info
