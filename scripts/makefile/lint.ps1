<#
.SYNOPSIS
  Run linters across services. Supports --Fix switch.
#>

[CmdletBinding()]
param(
    [switch]$Fix
)

. "$PSScriptRoot/_lib.ps1"

$ProjectRoot = Resolve-ProjectRoot
Push-Location $ProjectRoot
try {
    $Failures = @()

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
        if ($Fix) { npm run lint:naur -- --fix } else { npm run lint:naur }
        if ($LASTEXITCODE -eq 0) { npm run typecheck -w naurffxiv }
    }

    Run-Lint -Name "Moddingway (Python)" -Path "services/moddingway" -Command {
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
            if (Test-Path "venv/Scripts/python.exe") { venv/Scripts/python.exe -m ty check . } else { ty check . }
        }
    }

    Run-Lint -Name "Findingway (Go)" -Path "services/findingway" -Command {
        if ($Fix) { golangci-lint run --fix } else { golangci-lint run }
    }

    Run-Lint -Name "Clearingway (Go)" -Path "services/clearingway" -Command {
        if ($Fix) { golangci-lint run --fix } else { golangci-lint run }
    }

    Run-Lint -Name ".NET Services" -Path "." -Command {
        $projects = @("services/authingway/Authingway.csproj", "services/AppHost/Naur.AppHost.csproj")
        foreach ($proj in $projects) {
            if ($Fix) { dotnet format $proj } else { dotnet format $proj --verify-no-changes }
            if ($LASTEXITCODE -ne 0) { throw "dotnet format failed for $proj" }
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
