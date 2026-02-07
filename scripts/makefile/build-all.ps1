<#
.SYNOPSIS
  Parallel build of all monorepo services.
#>

[CmdletBinding()]
param(
    [string]$AppHostPrj   = "services/apphost/Naur.AppHost.csproj",
    [string]$AuthPrj      = "services/authingway/Naur.Authingway.csproj",
    [string]$NaurDir      = "services/naurffxiv",
    [string]$ModDir       = "services/moddingway",
    [string]$FindDir      = "services/findingway",
    [string]$ClearDir     = "services/clearingway"
)

. "$PSScriptRoot/_lib.ps1"

$ProjectRoot = Resolve-ProjectRoot

$Workloads = @(
    @{ Name = "AppHost"; Action = { param($p) dotnet build $p --verbosity quiet --nologo }; Arg = Join-Path $ProjectRoot $AppHostPrj }
    @{ Name = "Authingway"; Action = { param($p) dotnet build $p --verbosity quiet --nologo }; Arg = Join-Path $ProjectRoot $AuthPrj }
    @{ Name = "Naurffxiv"; Action = { param($p) Set-Location $p; npm run build }; Arg = Join-Path $ProjectRoot $NaurDir }
    @{ Name = "Findingway"; Action = { param($p) Set-Location $p; go build ./... }; Arg = Join-Path $ProjectRoot $FindDir }
    @{ Name = "Clearingway"; Action = { param($p) Set-Location $p; go build ./... }; Arg = Join-Path $ProjectRoot $ClearDir }
    @{ Name = "Moddingway"; Action = {
            param($p)
            Push-Location $p
            try {
                $python = Join-Path $p "venv\Scripts\python.exe"
                & $python -m ty check ; if ($LASTEXITCODE -ne 0) { throw "Ty Type-Check failed" }
            }
            finally { Pop-Location }
        }; Arg = Join-Path $ProjectRoot $ModDir
    }
)

$jobs = foreach ($w in $Workloads) { Start-TaskJob -Name $w.Name -ScriptBlock $w.Action -ArgumentList @($w.Arg) }
$results = Wait-TaskJobs -Jobs $jobs

$failed = $results | Where-Object { -not $_.Success }
if ($failed) {
    Write-Log -Level Error -Message "Build failures detected."
    foreach ($f in $failed) {
        Write-Host "`n--- Details for $($f.Name) ---" -ForegroundColor Yellow
        $f.Output | ForEach-Object { Write-Host $_ }
    }
    exit 1
}

Write-Log -Level Ok -Message "Monorepo Build Complete!"
