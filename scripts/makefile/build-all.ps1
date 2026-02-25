<#
.SYNOPSIS
  Parallel build of all monorepo services.
#>

[CmdletBinding()]
param()

. "$PSScriptRoot/_lib.ps1"
. "$PSScriptRoot/_config.ps1"

$ProjectRoot = Resolve-ProjectRoot
$pm = Get-PackageManagerCmd -CommandType "BuildCmd"

$Workloads = @(
    @{ Name = "AppHost"; Action = { param($p) dotnet build $p --verbosity quiet --nologo }; Arg = Get-ServicePath -ServiceName "AppHost" -ProjectRoot $ProjectRoot }
    @{ Name = "Authingway"; Action = { param($p) dotnet build $p --verbosity quiet --nologo }; Arg = Get-ServicePath -ServiceName "Authingway" -ProjectRoot $ProjectRoot }
    @{ Name = "Naurffxiv"; Action = { param($p, $pm) Set-Location $p; Invoke-Expression $pm }; Arg = @(Get-ServicePath -ServiceName "Naurffxiv" -ProjectRoot $ProjectRoot), $pm }
    @{ Name = "Findingway"; Action = { param($p) Set-Location $p; go build ./... }; Arg = Get-ServicePath -ServiceName "Findingway" -ProjectRoot $ProjectRoot }
    @{ Name = "Clearingway"; Action = { param($p) Set-Location $p; go build ./... }; Arg = Get-ServicePath -ServiceName "Clearingway" -ProjectRoot $ProjectRoot }
    @{ Name = "Moddingway"; Action = {
            param($p)
            Push-Location $p
            try {
                $python = Join-Path $p "venv\Scripts\python.exe"
                & $python -m ty check ; if ($LASTEXITCODE -ne 0) { throw "Ty Type-Check failed" }
            }
            finally { Pop-Location }
        }; Arg = Get-ServicePath -ServiceName "Moddingway" -ProjectRoot $ProjectRoot
    }
)

$jobs = foreach ($w in $Workloads) {
    $workloadArgs = if ($w.Arg -is [array]) { $w.Arg } else { @($w.Arg) }
    Start-TaskJob -Name $w.Name -ScriptBlock $w.Action -ArgumentList $workloadArgs
}
$results = Wait-TaskJobs -Jobs $jobs

$failed = $results | Where-Object { -not $_.Success }
if ($failed) {
    Write-Log -Level Error -Message "Build failures detected."
    foreach ($f in $failed) {
        Write-Host "`n--- Details for $($f.Name) ---" -ForegroundColor $Theme.Warn
        $f.Output | ForEach-Object { Write-Host $_ }
    }
    exit 1
}

Write-Log -Level Ok -Message "Monorepo Build Complete!"
