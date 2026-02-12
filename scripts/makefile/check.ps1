<#
.SYNOPSIS
  Check for required development tools.
#>

[CmdletBinding()]
param()

. "$PSScriptRoot/_lib.ps1"
. "$PSScriptRoot/_config.ps1"

$pmName = Get-PMName
$tools = Get-ToolRegistry

# Add package manager to tools list dynamically if not already present
if (-not ($tools | Where-Object { $_.Name -eq $pmName })) {
    $tools = @($tools) + @(@{ Name = $pmName; Cmd = $pmName; Required = $true })
}

$missing = [System.Collections.Generic.List[string]]::new()

Write-Log -Level Info -Message "Checking Development Prerequisites..."

foreach ($tool in $tools) {
    $exists = Test-ToolPresence -Tool $tool

    if ($exists) {
        Write-Log -Level Ok -Message "$($tool.Name) Found"
    } elseif ($tool.Required) {
        Write-Log -Level Error -Message "$($tool.Name) NOT Found"
        $missing.Add($tool.Name)
    } else {
        Write-Log -Level Warn -Message "$($tool.Name) not found (optional)"
    }
}

if ($missing.Count -gt 0) {
    Write-Log -Level Error -Message ("Missing mandatory tools: " + ($missing -join ', '))
    Write-Log -Level Info -Message "Run 'make setup' or 'make lint-tools' to install them."
    exit 1
}

exit 0
