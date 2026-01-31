<#
.SYNOPSIS
  Recursively remove common build artifacts.
#>

[CmdletBinding()]
param()

. "$PSScriptRoot/_lib.ps1"

$ProjectRoot = Resolve-ProjectRoot

$targets = @('node_modules', 'venv', '.venv', '.next', 'bin', 'obj', 'dist', '__pycache__', '.ruff_cache', 'artifacts')
$excludeDirs = @('.git', '.aspire', '.vs', '.vscode')

$foldersToDelete = [System.Collections.Generic.List[string]]::new()

function Find-Targets {
    param([string]$currentPath)
    try {
        $subDirs = [System.IO.Directory]::GetDirectories($currentPath)
        foreach ($dir in $subDirs) {
            $dirName = [System.IO.Path]::GetFileName($dir)
            if ($targets -contains $dirName) {
                $foldersToDelete.Add($dir)
                continue
            }
            if ($excludeDirs -notcontains $dirName) {
                Find-Targets -currentPath $dir
            }
        }
    } catch { }
}

Find-Targets -currentPath $ProjectRoot

if ($foldersToDelete.Count -eq 0) {
    Write-Log -Level Warn -Message 'No artifacts found. Repository is already clean.'
    exit 0
}

Write-Log -Level Info -Message "Found $($foldersToDelete.Count) artifact folder(s). Removing..."

$foldersToDelete | ForEach-Object -Parallel {
    $path = $_
    $root = $using:ProjectRoot
    $displayPath = $path.Replace($root, ".")
    Write-Host "Removing: $displayPath" -ForegroundColor Gray

    . "$using:PSScriptRoot/_lib.ps1"
    Remove-DirectoryRobust -Path $path
} -ThrottleLimit ([Environment]::ProcessorCount)

Write-Log -Level Ok -Message "Cleanup Complete."
