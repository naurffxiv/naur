<#
.SYNOPSIS
  Run linters across services. Supports --Fix switch.
#>

[CmdletBinding()]
param(
    [switch]$Fix
)

. "$PSScriptRoot/_lib.ps1"
. "$PSScriptRoot/_config.ps1"

$ProjectRoot = Resolve-ProjectRoot
Push-Location $ProjectRoot
try {
    $Failures = @()
    $pmLint = Get-PackageManagerCmd -CommandType "LintCmd"
    $pmTypecheck = Get-PackageManagerCmd -CommandType "TypecheckCmd"

    # Get service paths from registry
    $moddingwayPath = Get-ServicePath -ServiceName "Moddingway" -ProjectRoot $ProjectRoot
    $findingwayPath = Get-ServicePath -ServiceName "Findingway" -ProjectRoot $ProjectRoot
    $clearingwayPath = Get-ServicePath -ServiceName "Clearingway" -ProjectRoot $ProjectRoot

    # .NET services to lint
    $dotnetServices = @("AppHost", "Authingway")
    function Run-Lint {
        param([string]$Name, [string]$Path, [scriptblock]$Command)
        Write-Log -Level Info -Message "Linting $Name..."
        Push-Location $Path
        try {
            & $Command
            if ($LASTEXITCODE -eq 0) {
                Write-Log -Level Ok -Message $Name
            }
            else {
                Write-Log -Level Error -Message "$Name (exit $LASTEXITCODE)"
                $global:Failures += $Name
            }
        }
        catch {
            Write-Log -Level Error -Message ("{0}: {1}" -f $Name, $_)
            $global:Failures += $Name
        }
        finally {
            Pop-Location
        }
    }

    Run-Lint -Name "Naurffxiv (Next.js)" -Path "." -Command {
        if ($Fix) { Invoke-Expression "$pmLint --fix" } else { Invoke-Expression $pmLint }
        if ($LASTEXITCODE -eq 0) { Invoke-Expression $pmTypecheck }
    }

    Run-Lint -Name "Moddingway (Python)" -Path $moddingwayPath -Command {
        if ($Fix) {
            ruff check . --fix
            if ($LASTEXITCODE -ne 0) { throw "ruff check failed (fix)" }
            ruff format .
            if ($LASTEXITCODE -ne 0) { throw "ruff format failed" }
        }
        else {
            ruff check .
            if ($LASTEXITCODE -ne 0) { throw "ruff check failed" }
            ruff format . --check
            if ($LASTEXITCODE -ne 0) { throw "ruff format --check failed" }
        }
        if ($LASTEXITCODE -eq 0) {
            $pythonExe = Get-PythonPath -Path "."
            if ($pythonExe) { & $pythonExe -m ty check . } else { ty check . }
        }
    }

    Run-Lint -Name "Findingway (Go)" -Path $findingwayPath -Command {
        if ($Fix) { golangci-lint run --fix } else { golangci-lint run }
    }

    Run-Lint -Name "Clearingway (Go)" -Path $clearingwayPath -Command {
        if ($Fix) { golangci-lint run --fix } else { golangci-lint run }
    }

    Run-Lint -Name ".NET Services" -Path "." -Command {
        foreach ($serviceName in $dotnetServices) {
            $proj = Get-ServicePath -ServiceName $serviceName -ProjectRoot $ProjectRoot
            if ($Fix) { dotnet format $proj } else { dotnet format $proj --verify-no-changes }
            if ($LASTEXITCODE -ne 0) { throw "dotnet format failed for $serviceName" }
        }
    }

    Write-Log -Level Info -Message "Lint summary:"
    if ($Failures.Count -gt 0) {
        Write-Log -Level Error -Message ("Linting failed for: " + ($Failures -join ', '))
        exit 1
    }
    else {
        Write-Log -Level Ok -Message "All linters passed!"
        exit 0
    }
}
finally {
    Pop-Location
}
