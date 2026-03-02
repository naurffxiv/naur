<#
.SYNOPSIS
  Monorepo health and diagnostics.
#>

[CmdletBinding()]
param()

. "$PSScriptRoot/_lib.ps1"
. "$PSScriptRoot/_config.ps1"

$ProjectRoot = Resolve-ProjectRoot
$pmName = Get-PMName

Write-Log -Level Info -Message "--- Monorepo Health Check ---"

Write-Log -Level Info -Message "[1/2] Verifying Global Tooling..."
$tools = @(
    @{ Name = ".NET SDK"; Cmd = "dotnet"; Args = "--version" }
    @{ Name = "Node.js";  Cmd = "node";   Args = "--version" }
    @{ Name = "npm";      Cmd = "npm";    Args = "--version" }
    @{ Name = $pmName;    Cmd = $pmName;  Args = "--version" }
    @{ Name = "uv";       Cmd = "uv";     Args = "--version" }
    @{ Name = "Python";   Cmd = "py";     Args = @("-3.14", "--version") }
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
$registry = Get-ServiceRegistry
$services = foreach ($key in $registry.Keys) {
    $svc = $registry[$key]
    @{
        Name = $svc.DisplayName
        Path = Get-ServicePath -ServiceName $key -ProjectRoot $ProjectRoot
        Type = $svc.Type
        Proj = if ($svc.ProjectPath) { [System.IO.Path]::GetFileNameWithoutExtension($svc.ProjectPath) } else { $null }
    }
}

foreach ($s in @($services)) {
    if ($null -eq $s -or [string]::IsNullOrWhiteSpace($s.Path)) { continue }

    $checkPath = if ((Test-Path $s.Path -PathType Leaf)) { Split-Path $s.Path -Parent } else { $s.Path }

    if ($null -eq $checkPath -or -not (Test-Path $checkPath)) {
        Write-Log -Level Error -Message "$($s.Name): Directory missing!"
        continue
    }

    $isInstalled = Test-ServiceHealth -Type $s.Type -Path $checkPath -ProjectRoot $ProjectRoot -DotNetProjName $s.Proj

    if ($isInstalled) {
        Write-Log -Level Ok -Message "$($s.Name) ($($s.Type))"
    } else {
        Write-Log -Level Warn -Message "$($s.Name): Missing local artifacts (Run 'make install')"
    }
}

Write-Log -Level Info -Message "[AppHost Project]"
$appHostPath = Get-ServicePath -ServiceName "AppHost" -ProjectRoot $ProjectRoot
if ($null -eq $appHostPath) {
    Write-Log -Level Error -Message "AppHost: Path could not be resolved!"
} elseif (Test-Path $appHostPath) {
    Write-Log -Level Ok -Message "AppHost found"
} else {
    Write-Log -Level Error -Message "AppHost project file missing!"
}

Write-Log -Level Info -Message "--- Diagnostics Complete ---"
