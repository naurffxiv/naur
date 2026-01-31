<#
.SYNOPSIS
  Parallel install of monorepo dependencies.

.DESCRIPTION
  Restores .NET, downloads Go modules, runs npm install at the repo root and prepares Python venvs.
#>

[CmdletBinding()]
param(
    [string]$AppHostPrj = "services/AppHost/Naur.AppHost.csproj",
    [string]$AuthPrj    = "services/authingway/Authingway.csproj",
    [string]$NaurDir    = "services/naurffxiv",
    [string]$ModDir     = "services/moddingway",
    [string]$FindDir    = "services/findingway",
    [string]$ClearDir   = "services/clearingway"
)

. "$PSScriptRoot/_lib.ps1"

$ProjectRoot = Resolve-ProjectRoot

$Tasks = @(
    @{ Name = "AppHost";     Action = { param($p) dotnet restore $p --verbosity quiet }; Arg = Join-Path $ProjectRoot $AppHostPrj }
    @{ Name = "Authingway";  Action = { param($p) dotnet restore $p --verbosity quiet }; Arg = Join-Path $ProjectRoot $AuthPrj }
    @{ Name = "Root-npm";    Action = { param($p) Set-Location $p; npm install --loglevel error }; Arg = $ProjectRoot }
    @{ Name = "Findingway";  Action = { param($p) Set-Location $p; go mod download }; Arg = Join-Path $ProjectRoot $FindDir }
    @{ Name = "Clearingway"; Action = { param($p) Set-Location $p; go mod download }; Arg = Join-Path $ProjectRoot $ClearDir }
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
    }; Arg = Join-Path $ProjectRoot $ModDir }
)

$jobs = foreach ($t in $Tasks) {
    if ($t.Arg) { Start-TaskJob -Name $t.Name -ScriptBlock $t.Action -ArgumentList @($t.Arg) }
    else       { Start-TaskJob -Name $t.Name -ScriptBlock $t.Action }
}

$results = Wait-TaskJobs -Jobs $jobs

$successCount = ($results | Where-Object { $_.Success }).Count
Write-Host "($successCount) services installation finished." -ForegroundColor Green

$failed = $results | Where-Object { -not $_.Success }
if ($failed) {
    Write-Log -Level Error -Message "One or more install tasks failed."
    foreach ($f in $failed) {
        Write-Host "`n--- Output for $($f.Name) ---" -ForegroundColor Yellow
        $f.Output | ForEach-Object { Write-Host $_ }
    }
    exit 1
}

# Verify artifacts
Write-Log -Level Info -Message "Verifying Service Artifacts"

$services = @(
    @{ Name = "AppHost";     Path = Join-Path $ProjectRoot $AppHostPrj; Type = "DotNet" }
    @{ Name = "Authingway";  Path = Join-Path $ProjectRoot $AuthPrj;  Type = "DotNet" }
    @{ Name = "Naurffxiv";   Path = Join-Path $ProjectRoot $NaurDir;  Type = "Node" }
    @{ Name = "Moddingway";  Path = Join-Path $ProjectRoot $ModDir;   Type = "Python" }
    @{ Name = "Findingway";  Path = Join-Path $ProjectRoot $FindDir;  Type = "Go" }
    @{ Name = "Clearingway"; Path = Join-Path $ProjectRoot $ClearDir; Type = "Go" }
)

$failedChecks = @()
foreach ($s in $services) {
    if (-not (Test-Path $s.Path)) {
        Write-Log -Level Error -Message "$($s.Name): Directory not found"
        $failedChecks += $s.Name; continue
    }
    $isInstalled = switch ($s.Type) {
        "Node"   { Test-Path (Join-Path $s.Path "node_modules") }
        "Python" { Test-Path (Join-Path $s.Path "venv") }
        "Go"     { Test-Path (Join-Path $s.Path "go.mod") }
        "DotNet" {
            $projName = [System.IO.Path]::GetFileNameWithoutExtension($s.Path)
            $artifactsObjPath = Join-Path $ProjectRoot "artifacts/obj/$projName"
            (Test-Path $artifactsObjPath) -and (Get-ChildItem $artifactsObjPath -ErrorAction SilentlyContinue)
        }
        Default { $false }
    }
    if ($isInstalled) { Write-Log -Level Ok -Message "$($s.Name) ($($s.Type)) ready" } else { Write-Log -Level Warn -Message "$($s.Name): Verification failed"; $failedChecks += $s.Name }
}

# Pre-commit hooks
try {
    & npx prek --version > $null 2>&1
    Write-Log -Level Ok -Message "prek (pre-commit hooks) installed"
} catch {
    Write-Log -Level Error -Message "prek: Validation failed. Run 'npm install' manually."
    $failedChecks += "prek"
}

if ($failedChecks.Count -gt 0) {
    Write-Log -Level Error -Message "Monorepo setup is incomplete."
    exit 1
}

Write-Host "`n  [SUCCESS] Monorepo is fully installed and ready!" -ForegroundColor Cyan
