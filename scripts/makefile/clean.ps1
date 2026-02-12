<#
.SYNOPSIS
  Recursively remove common build artifacts.
#>

[CmdletBinding()]
param()

. "$PSScriptRoot/_lib.ps1"
. "$PSScriptRoot/_config.ps1"

$ProjectRoot = Resolve-ProjectRoot
$cleanupConfig = Get-CleanupConfig
$targets = $cleanupConfig.Targets
$excludeDirs = $cleanupConfig.ExcludeDirs

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
    } catch {
        Write-Verbose "Could not access '$currentPath': $_"
    }}

Find-Targets -currentPath $ProjectRoot

if ($foldersToDelete.Count -eq 0) {
    Write-Log -Level Warn -Message 'No artifacts found. Repository is already clean.'
    exit 0
}

# Display what will be deleted
Write-Log -Level Info -Message "Found $($foldersToDelete.Count) artifact folder(s):"
Write-Host ""
foreach ($folder in $foldersToDelete) {
    $displayPath = $folder.Replace($ProjectRoot, ".")
    Write-Host "  - $displayPath" -ForegroundColor $Theme.Debug
}
Write-Host ""

# Calculate total size
$totalSize = 0
foreach ($folder in $foldersToDelete) {
    if (Test-Path $folder) {
        try {
            $size = (Get-ChildItem -Path $folder -Recurse -Force -ErrorAction SilentlyContinue |
                     Measure-Object -Property Length -Sum -ErrorAction SilentlyContinue).Sum
            if ($size) { $totalSize += $size }
        } catch {
            Write-Verbose "Could not calculate size for '$folder': $_"
        }    }
}
$sizeInMB = [math]::Round($totalSize / 1MB, 2)
Write-Host "  Total size: ~$sizeInMB MB" -ForegroundColor $Theme.Info
Write-Host ""

# Ask for confirmation
Write-Host "Are you sure you want to delete these folders? [Y/n]: " -NoNewline -ForegroundColor $Theme.Warn
$response = Read-Host
if ($response -and $response -notmatch '^[Yy]') {
    Write-Log -Level Warn -Message "Cleanup cancelled by user."
    exit 0
}

Write-Log -Level Info -Message "Removing artifact folders..."

$foldersToDelete | ForEach-Object -Parallel {
    $path = $_
    $root = $using:ProjectRoot
    $theme = $using:Theme
    $displayPath = $path.Replace($root, ".")
    Write-Host "Removing: $displayPath" -ForegroundColor $theme.Debug

    . "$using:PSScriptRoot/_lib.ps1"
    Remove-DirectoryRobust -Path $path
} -ThrottleLimit ([Environment]::ProcessorCount)

Write-Log -Level Ok -Message "Cleanup Complete."
