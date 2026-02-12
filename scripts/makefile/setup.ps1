<#
.SYNOPSIS
  Install system prerequisites via winget (Windows).
#>

[CmdletBinding()]
param()

. "$PSScriptRoot/_lib.ps1"
. "$PSScriptRoot/_config.ps1"

$pmName = Get-PMName
$coreStack = Get-PrerequisiteRegistry

$coreStack = @($coreStack) + @(@{ Name = $pmName; Cmd = $pmName; Id = "$pmName.$pmName" })

if (-not (Get-Command winget -ErrorAction SilentlyContinue)) {
    Write-Log -Level Error -Message "winget is not installed. Please install App Installer from the Microsoft Store."
    exit 1
}

$missing = @()
foreach ($tool in $coreStack) {
    $present = Test-ToolPresence -Tool $tool

    if (-not $present) {
        Write-Log -Level Warn -Message "$($tool.Name) missing"
        $missing += $tool
    } else {
        Write-Log -Level Ok -Message "$($tool.Name) already installed"
    }
}

if ($missing.Count -gt 0) {
    $failed = @()
    foreach ($tool in $missing) {
        Write-Log -Level Info -Message "Installing $($tool.Name)..."
        winget install -e --id $tool.Id --silent --accept-source-agreements --accept-package-agreements
        if ($LASTEXITCODE -ne 0) {
            Write-Log -Level Error -Message "Failed to install $($tool.Name)"
            $failed += $tool
        }
    }

    Write-Log -Level Info -Message "Refreshing PATH..."
    $env:Path = [System.Environment]::GetEnvironmentVariable("Path", "Machine") + ";" + [System.Environment]::GetEnvironmentVariable("Path", "User")
    Write-Log -Level Info -Message "You may need to restart your terminal to use newly installed tools."

    Write-Log -Level Info -Message "Verifying installations..."
    foreach ($tool in $missing) {
        $verified = Test-ToolPresence -Tool $tool

        if ($verified) {
            Write-Log -Level Ok -Message "$($tool.Name)"
        } else {
            Write-Log -Level Error -Message "$($tool.Name) - restart may be required"
        }
    }
    if ($failed.Count -gt 0) {
        Write-Log -Level Warn -Message "Prerequisites installation completed with $($failed.Count) failure(s)."
    } else {
        Write-Log -Level Ok -Message "Prerequisites installation process completed."
    }
} else {
    Write-Log -Level Info -Message "No system installations required."
}
