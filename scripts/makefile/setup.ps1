<#
.SYNOPSIS
  Install system prerequisites via winget (Windows).
#>

[CmdletBinding()]
param()

. "$PSScriptRoot/_lib.ps1"

$coreStack = @(
    @{ Name = ".NET 10";      Cmd = "dotnet"; Id = "Microsoft.DotNet.SDK.10" }
    @{ Name = "Node.js";      Cmd = "node";   Id = "OpenJS.NodeJS" }
    @{ Name = "Python 3.11";  Cmd = "python"; Id = "Python.Python.3.11" }
    @{ Name = "Go";           Cmd = "go";     Id = "GoLang.Go" }
    @{ Name = "PowerShell 7"; Cmd = "pwsh";   Id = "Microsoft.PowerShell" }
    @{ Name = "Docker";       Cmd = "docker"; Id = "Docker.DockerDesktop" }
)

if (-not (Get-Command winget -ErrorAction SilentlyContinue)) {
    Write-Log -Level Error -Message "winget is not installed. Please install App Installer from the Microsoft Store."
    exit 1
}

$missing = @()
foreach ($tool in $coreStack) {
    if (-not (Get-Command $tool.Cmd -ErrorAction SilentlyContinue)) {
        Write-Log -Level Warn -Message "$($tool.Name) missing"
        $missing += $tool
    } else {
        Write-Log -Level Ok -Message "$($tool.Name) already installed"
    }
}

if ($missing.Count -gt 0) {
    foreach ($tool in $missing) {
        Write-Log -Level Info -Message "Installing $($tool.Name)..."
        winget install -e --id $tool.Id --silent --accept-source-agreements --accept-package-agreements
        if ($LASTEXITCODE -ne 0) {
            Write-Log -Level Error -Message "Failed to install $($tool.Name)"
        }
    }

    Write-Log -Level Info -Message "Refreshing PATH..."
    $env:Path = [System.Environment]::GetEnvironmentVariable("Path", "Machine") + ";" + [System.Environment]::GetEnvironmentVariable("Path", "User")
    Write-Log -Level Info -Message "You may need to restart your terminal to use newly installed tools."

    Write-Log -Level Info -Message "Verifying installations..."
    foreach ($tool in $missing) {
        if (Get-Command $tool.Cmd -ErrorAction SilentlyContinue) {
            Write-Log -Level Ok -Message "$($tool.Name)"
        } else {
            Write-Log -Level Error -Message "$($tool.Name) - restart may be required"
        }
    }
    Write-Log -Level Ok -Message "Prerequisites installation process completed."
} else {
    Write-Log -Level Info -Message "No system installations required."
}
