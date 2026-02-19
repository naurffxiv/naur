<#
.SYNOPSIS
  Centralized configuration for all monorepo services and tools.

.DESCRIPTION
  Single source of truth for service metadata, paths, and package manager settings.
  This eliminates duplication across build-all.ps1, install.ps1, check.ps1, and troubleshoot.ps1.
#>

$ErrorActionPreference = 'Stop'

$script:ServiceRegistry = @{
    AppHost     = @{
        Type        = "DotNet"
        ProjectPath = "services/apphost/Naur.AppHost.csproj"
        DisplayName = "AppHost"
    }
    Authingway  = @{
        Type        = "DotNet"
        ProjectPath = "services/authingway/Naur.Authingway.csproj"
        DisplayName = "Authingway"
    }
    Naurffxiv   = @{
        Type        = "Node"
        DirPath     = "services/naurffxiv"
        DisplayName = "Naurffxiv"
    }
    Moddingway  = @{
        Type        = "Python"
        DirPath     = "services/moddingway"
        DisplayName = "Moddingway"
    }
    Findingway  = @{
        Type        = "Go"
        DirPath     = "services/findingway"
        DisplayName = "Findingway"
    }
    Clearingway = @{
        Type        = "Go"
        DirPath     = "services/clearingway"
        DisplayName = "Clearingway"
    }
    E2ETests    = @{
        Type        = "Playwright"
        DirPath     = "tests/e2e"
        DisplayName = "E2E-Tests"
    }
}

$script:PackageManager = @{
    Name         = "pnpm"
    InstallCmd   = "pnpm install --silent"
    BuildCmd     = "pnpm build"
    ExecCmd      = "pnpm exec"
    FilterCmd    = "pnpm --filter"
    LintCmd      = "pnpm lint:naur"
    TypecheckCmd = "pnpm typecheck"
}

function Get-ServiceConfig {
    <#
    .SYNOPSIS
      Retrieve service configuration by name.
    #>
    param([Parameter(Mandatory)][string]$ServiceName)

    $config = $script:ServiceRegistry[$ServiceName]
    if (-not $config) {
        throw "Unknown service: $ServiceName. Available: $($script:ServiceRegistry.Keys -join ', ')"
    }
    return $config
}

function Get-ServicePath {
    <#
    .SYNOPSIS
      Get absolute path for a service.
    .NOTES
      Requires Resolve-ProjectRoot to be defined if ProjectRoot parameter is not provided.
    #>
    param(
        [Parameter(Mandatory)][string]$ServiceName,
        [string]$ProjectRoot
    )

    $config = Get-ServiceConfig -ServiceName $ServiceName
    if (-not $ProjectRoot) {
        if (-not (Get-Command Resolve-ProjectRoot -ErrorAction SilentlyContinue)) {
            throw "ProjectRoot parameter required or Resolve-ProjectRoot function must be defined"
        }
        $root = Resolve-ProjectRoot
    } else {
        $root = $ProjectRoot
    }

    # Validate that service config has a path defined
    $relativePath = if ($config.ProjectPath) { $config.ProjectPath } else { $config.DirPath }

    if ([string]::IsNullOrWhiteSpace($relativePath)) {
        throw "Service '$ServiceName' is missing both ProjectPath and DirPath in configuration. Please update the ServiceRegistry."
    }

    return Join-Path $root $relativePath
}
function Get-AllServices {
    <#
    .SYNOPSIS
      Get list of all registered service names.
    #>
    param()
    return $script:ServiceRegistry.Keys
}

function Get-ServiceRegistry {
    <#
    .SYNOPSIS
      Returns the full service registry map.
    #>
    param()
    return $script:ServiceRegistry
}

function Get-PackageManagerCmd {
    <#
    .SYNOPSIS
      Get a specific package manager command.
    #>
    param(
        [Parameter(Mandatory)]
        [ValidateSet("Name", "InstallCmd", "BuildCmd", "ExecCmd", "FilterCmd", "LintCmd", "TypecheckCmd")]
        [string]$CommandType
    )

    return $script:PackageManager[$CommandType]
}
function Get-PMName {
    <#
    .SYNOPSIS
      Get the configured package manager name.
    #>
    return Get-PackageManagerCmd -CommandType "Name"
}

$script:ToolRegistry = @(
    @{ Name = ".NET";          Cmd = "dotnet";        Required = $true }
    @{ Name = "Node.js";       Cmd = "node";          Required = $true }
    @{ Name = "Python";        Cmd = "python";        Required = $true }
    @{ Name = "Go";            Cmd = "go";            Required = $true }
    @{ Name = "Ruff";          Cmd = "ruff";          Required = $true }
    @{ Name = "golangci-lint"; Cmd = "golangci-lint"; Required = $true }
    @{ Name = "Docker";        Cmd = "docker";        Required = $true }
    @{ Name = "Playwright";    Cmd = "playwright";    Required = $true; Type = "Playwright" }
)

$script:PrerequisiteRegistry = @(
    @{ Name = ".NET 10";      Cmd = "dotnet"; Id = "Microsoft.DotNet.SDK.10" }
    @{ Name = "Node.js";      Cmd = "node";   Id = "OpenJS.NodeJS" }
    @{ Name = "Python 3.14";  Cmd = "python"; Id = "Python.Python.3.14"; Check = { (py -3.14 --version) -match "3.14" } }
    @{ Name = "Go";           Cmd = "go";     Id = "GoLang.Go" }
    @{ Name = "PowerShell 7"; Cmd = "pwsh";   Id = "Microsoft.PowerShell" }
    @{ Name = "Docker";       Cmd = "docker"; Id = "Docker.DockerDesktop" }
)

$script:CleanupConfig = @{
    Targets = @(
        'node_modules', 'venv', '.venv', '.next', 'bin', 'obj', 'dist',
        '__pycache__', '.ruff_cache', 'artifacts', 'build',
        '.pytest_cache', '.mypy_cache', 'test-results', 'playwright-report',
        'coverage', 'TestResults', '.swc'
    )
    ExcludeDirs = @('.git', '.aspire', '.vs', '.vscode', '.idea', '.github', '.docker')
}

function Get-ToolRegistry {
    <#
    .SYNOPSIS
      Get the list of development tools with their metadata.
    #>
    param()
    return $script:ToolRegistry
}

function Get-PrerequisiteRegistry {
    <#
    .SYNOPSIS
      Get the list of system prerequisites for setup.
    #>
    param()
    return $script:PrerequisiteRegistry
}

function Get-CleanupConfig {
    <#
    .SYNOPSIS
      Get cleanup configuration (targets and exclude dirs).
    #>
    param()
    return $script:CleanupConfig
}
