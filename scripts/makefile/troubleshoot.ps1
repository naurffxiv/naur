<#
.SYNOPSIS
  Monorepo health and diagnostics.
#>

[CmdletBinding()]
param(
    [string]$AuthDir    = "services/authingway",
    [string]$NaurDir    = "services/naurffxiv",
    [string]$ModDir     = "services/moddingway",
    [string]$FindDir    = "services/findingway",
    [string]$ClearDir   = "services/clearingway",
    [string]$AppHostPrj = "services/apphost/Naur.AppHost.csproj"
)

. "$PSScriptRoot/_lib.ps1"

$ProjectRoot = Resolve-ProjectRoot
Write-Log -Level Info -Message "--- Monorepo Health Check ---"

Write-Log -Level Info -Message "[1/2] Verifying Global Tooling..."
$tools = @(
    @{ Name = ".NET SDK"; Cmd = "dotnet"; Args = "--version" }
    @{ Name = "Node.js";  Cmd = "node";   Args = "--version" }
    @{ Name = "npm";      Cmd = "npm";    Args = "--version" }
    @{ Name = "Python";   Cmd = "python"; Args = "--version" }
    @{ Name = "Go";       Cmd = "go";     Args = "version" }
    @{ Name = "Docker";   Cmd = "docker"; Args = "--version" }
    @{ Name = "Git";      Cmd = "git";    Args = "--version" }
)

foreach ($t in $tools) {
    try {
        if (Get-Command $t.Cmd -ErrorAction SilentlyContinue) {
            $version = & $t.Cmd $t.Args 2>$null
            if ($t.Cmd -eq 'go' -and $version) { $version = ($version -split ' ')[2] }
            Write-Log -Level Ok -Message "$($t.Name): $version"
        } else {
            Write-Log -Level Error -Message "$($t.Name) NOT FOUND"
        }
    } catch {
        Write-Log -Level Warn -Message "$($t.Name) check failed: $_"
    }
}

Write-Log -Level Info -Message "[2/2] Checking Service Health..."
$services = @(
    @{ Name = "Authingway";  Path = $AuthDir;  Type = "DotNet"; Proj = "Authingway" }
    @{ Name = "Naurffxiv";   Path = $NaurDir;  Type = "Node" }
    @{ Name = "Moddingway";  Path = $ModDir;   Type = "Python" }
    @{ Name = "Findingway";  Path = $FindDir;  Type = "Go" }
    @{ Name = "Clearingway"; Path = $ClearDir; Type = "Go" }
)

foreach ($s in $services) {
    if ($null -eq $s.Path -or -not (Test-Path $s.Path)) {
        Write-Log -Level Error -Message "$($s.Name): Directory missing!"
        continue
    }

    $isInstalled = switch ($s.Type) {
        "Node"   { Test-Path (Join-Path $s.Path "node_modules") }
        "Python" {
            $venvPython = Join-Path $s.Path "venv\Scripts\python.exe"
            if (-not (Test-Path $venvPython)) { $venvPython = Join-Path $s.Path "venv/bin/python" }
            if (Test-Path $venvPython) {
                $requirementsTxt = Join-Path $s.Path "requirements.txt"
                if (Test-Path $requirementsTxt) {
                    $installed = & $venvPython -m pip list --format=freeze 2>$null
                    ($null -ne $installed) -and ($installed.Count -gt 0)
                } else {
                    $true
                }
            } else {
                $false
            }
        }
        "Go"     { Test-Path (Join-Path $s.Path "go.mod") }
        "DotNet" {
            $artifactsObjPath = Join-Path $ProjectRoot "artifacts/obj/$($s.Proj)"
            (Test-Path $artifactsObjPath) -and (Get-ChildItem $artifactsObjPath -ErrorAction SilentlyContinue)
        }
        Default { $false }
    }

    if ($isInstalled) {
        Write-Log -Level Ok -Message "$($s.Name) ($($s.Type))"
    } else {
        Write-Log -Level Warn -Message "$($s.Name): Missing local artifacts (Run 'make install')"
    }
}

Write-Log -Level Info -Message "[AppHost Project]"
if ($null -ne $AppHostPrj -and (Test-Path $AppHostPrj)) {
    Write-Log -Level Ok -Message "AppHost found"
} else {
    Write-Log -Level Error -Message "AppHost project file missing!"
}

Write-Log -Level Info -Message "--- Diagnostics Complete ---"
