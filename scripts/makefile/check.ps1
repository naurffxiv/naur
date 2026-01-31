<#
.SYNOPSIS
  Check for required development tools.
#>

[CmdletBinding()]
param()

. "$PSScriptRoot/_lib.ps1"

$tools = @(
    @{ Name = ".NET";          Cmd = "dotnet";        Required = $true }
    @{ Name = "Node.js";       Cmd = "node";          Required = $true }
    @{ Name = "Python";        Cmd = "python";        Required = $true }
    @{ Name = "Go";            Cmd = "go";            Required = $true }
    @{ Name = "Ruff";          Cmd = "ruff";          Required = $true }
    @{ Name = "golangci-lint"; Cmd = "golangci-lint"; Required = $true }
    @{ Name = "Docker";        Cmd = "docker";        Required = $true }
)

$missing = [System.Collections.Generic.List[string]]::new()

Write-Log -Level Info -Message "Checking Development Prerequisites..."

foreach ($tool in $tools) {
    if (Get-Command $tool.Cmd -ErrorAction SilentlyContinue) {
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
