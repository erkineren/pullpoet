# PullPoet Windows Install Script
# Automatically detects architecture and installs latest release from GitHub
# 
# Usage Examples:
#   # Install latest version
#   Invoke-Expression (Invoke-WebRequest -Uri "https://raw.githubusercontent.com/erkineren/pullpoet/main/scripts/install.ps1" -UseBasicParsing).Content
#   
#   # Update to latest version (correct parameter passing)
#   Invoke-Expression "$((Invoke-WebRequest -Uri 'https://raw.githubusercontent.com/erkineren/pullpoet/main/scripts/install.ps1' -UseBasicParsing).Content) -Update"
#   
#   # Install to custom directory
#   Invoke-Expression "$((Invoke-WebRequest -Uri 'https://raw.githubusercontent.com/erkineren/pullpoet/main/scripts/install.ps1' -UseBasicParsing).Content) -InstallDir 'C:\Tools\pullpoet'"
#   
#   # Uninstall
#   Invoke-Expression "$((Invoke-WebRequest -Uri 'https://raw.githubusercontent.com/erkineren/pullpoet/main/scripts/install.ps1' -UseBasicParsing).Content) -Uninstall"
#   
#   # Alternative: Download and run locally
#   Invoke-WebRequest -Uri "https://raw.githubusercontent.com/erkineren/pullpoet/main/scripts/install.ps1" -OutFile "install.ps1"
#   .\install.ps1 -Uninstall

# Parse parameters first for Invoke-Expression compatibility
$Update = $false
$Force = $false
$Uninstall = $false
$Help = $false
$InstallDir = ""

# Configuration
$RepoOwner = "erkineren"
$RepoName = "pullpoet"
$RepoUrl = "https://github.com/$RepoOwner/$RepoName"
$DefaultInstallDir = "$env:LOCALAPPDATA\Programs\pullpoet"
$BinaryName = "pullpoet.exe"

# Parse arguments manually for Invoke-Expression compatibility
for ($i = 0; $i -lt $args.Count; $i++) {
    switch ($args[$i]) {
        "-Update" { $Update = $true }
        "-Force" { $Force = $true }
        "-Uninstall" { $Uninstall = $true }
        "-Help" { $Help = $true }
        "-InstallDir" { 
            if ($i + 1 -lt $args.Count) {
                $InstallDir = $args[$i + 1]
                $i++ # Skip next argument as it's the value
            }
        }
    }
}

# Set default installation directory if not provided
if (-not $InstallDir) {
    $InstallDir = $DefaultInstallDir
}

# Colors for output
$Colors = @{
    Red = "Red"
    Green = "Green"
    Yellow = "Yellow"
    Blue = "Blue"
    White = "White"
}

function Write-Info {
    param([string]$Message)
    Write-Host "[INFO] $Message" -ForegroundColor $Colors.Blue
}

function Write-Success {
    param([string]$Message)
    Write-Host "[SUCCESS] $Message" -ForegroundColor $Colors.Green
}

function Write-Warning {
    param([string]$Message)
    Write-Host "[WARNING] $Message" -ForegroundColor $Colors.Yellow
}

function Write-Error {
    param([string]$Message)
    Write-Host "[ERROR] $Message" -ForegroundColor $Colors.Red
}

# Function to detect Windows architecture
function Get-WindowsArchitecture {
    if ([Environment]::Is64BitOperatingSystem) {
        return "x86_64"
    } else {
        return "i386"
    }
}

# Function to get latest release version from GitHub
function Get-LatestVersion {
    try {
        $apiUrl = "https://api.github.com/repos/$RepoOwner/$RepoName/releases/latest"
        $response = Invoke-RestMethod -Uri $apiUrl -Method Get
        $version = $response.tag_name -replace '^v', ''
        return $version
    }
    catch {
        Write-Error "Failed to fetch latest version from GitHub: $($_.Exception.Message)"
        return $null
    }
}

# Function to get currently installed version
function Get-InstalledVersion {
    $binaryPath = Join-Path $InstallDir $BinaryName
    if (Test-Path $binaryPath) {
        try {
            $output = & $binaryPath --version 2>$null
            if ($output) {
                $version = $output -replace '[^0-9.]', ''
                return $version
            }
        }
        catch {
            # Ignore errors
        }
    }
    return "not_installed"
}

# Function to compare versions
function Compare-Versions {
    param([string]$Version1, [string]$Version2)
    
    $v1Parts = $Version1.Split('.') | ForEach-Object { [int]$_ }
    $v2Parts = $Version2.Split('.') | ForEach-Object { [int]$_ }
    
    for ($i = 0; $i -lt [Math]::Max($v1Parts.Length, $v2Parts.Length); $i++) {
        $v1Part = if ($i -lt $v1Parts.Length) { $v1Parts[$i] } else { 0 }
        $v2Part = if ($i -lt $v2Parts.Length) { $v2Parts[$i] } else { 0 }
        
        if ($v1Part -gt $v2Part) { return 1 }
        if ($v1Part -lt $v2Part) { return -1 }
    }
    return 0
}

# Function to download and install binary
function Install-Binary {
    param([string]$Version, [string]$Architecture)
    
    $downloadUrl = "https://github.com/$RepoOwner/$RepoName/releases/download/v$Version/pullpoet_Windows_$Architecture.zip"
    
    Write-Info "Downloading $BinaryName v$Version for Windows $Architecture..."
    Write-Info "Download URL: $downloadUrl"
    
    # Create temporary directory
    $tempDir = New-TemporaryFile | ForEach-Object { Remove-Item $_; New-Item -ItemType Directory -Path $_ }
    $archiveFile = Join-Path $tempDir "archive.zip"
    
    try {
        # Download the release archive
        Write-Info "Downloading archive..."
        Invoke-WebRequest -Uri $downloadUrl -OutFile $archiveFile -UseBasicParsing
        
        # Extract the archive
        Write-Info "Extracting archive..."
        $extractDir = Join-Path $tempDir "extract"
        New-Item -ItemType Directory -Path $extractDir -Force | Out-Null
        
        Expand-Archive -Path $archiveFile -DestinationPath $extractDir -Force
        
        # Find the binary
        $binaryPath = Get-ChildItem -Path $extractDir -Recurse -Name $BinaryName | Select-Object -First 1
        if (-not $binaryPath) {
            Write-Error "Binary not found in the downloaded archive"
            return
        }
        
        $fullBinaryPath = Join-Path $extractDir $binaryPath
        
        # Create installation directory
        if (-not (Test-Path $InstallDir)) {
            New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
        }
        
        # Install the binary
        Write-Info "Installing $BinaryName to $InstallDir..."
        Copy-Item -Path $fullBinaryPath -Destination (Join-Path $InstallDir $BinaryName) -Force
        
        Write-Success "$BinaryName v$Version installed successfully!"
    }
    catch {
        Write-Error "Failed to download or install: $($_.Exception.Message)"
    }
    finally {
        # Cleanup
        if (Test-Path $tempDir) {
            Remove-Item -Path $tempDir -Recurse -Force
        }
    }
}

# Function to uninstall
function Uninstall-PullPoet {
    Write-Info "Uninstalling $BinaryName..."
    
    $binaryPath = Join-Path $InstallDir $BinaryName
    
    if (Test-Path $binaryPath) {
        Remove-Item -Path $binaryPath -Force
        Write-Success "$BinaryName uninstalled successfully!"
        
        # Remove directory if empty
        if ((Get-ChildItem -Path $InstallDir -Force | Measure-Object).Count -eq 0) {
            Remove-Item -Path $InstallDir -Force
            Write-Info "Installation directory removed: $InstallDir"
        }
    } else {
        Write-Warning "$BinaryName is not installed in $InstallDir"
    }
}

# Function to add to PATH
function Add-ToPath {
    $currentPath = [Environment]::GetEnvironmentVariable("PATH", "User")
    if ($currentPath -notlike "*$InstallDir*") {
        $newPath = "$currentPath;$InstallDir"
        [Environment]::SetEnvironmentVariable("PATH", $newPath, "User")
        Write-Success "Added $InstallDir to PATH"
        
        # Refresh current session's PATH
        $env:PATH = "$env:PATH;$InstallDir"
        Write-Info "Updated current session PATH - pullpoet is now available!"
        
        # Verify the binary is accessible
        $binaryPath = Join-Path $InstallDir $BinaryName
        if (Test-Path $binaryPath) {
            try {
                $testOutput = & $binaryPath --version 2>$null
                if ($testOutput) {
                    Write-Success "✅ PullPoet is ready to use in this session!"
                }
            }
            catch {
                Write-Warning "Binary found but could not execute. You may need to restart your terminal."
            }
        }
    } else {
        Write-Info "$InstallDir is already in PATH"
        
        # Still refresh current session PATH to ensure it's available
        if ($env:PATH -notlike "*$InstallDir*") {
            $env:PATH = "$env:PATH;$InstallDir"
            Write-Info "Updated current session PATH"
        }
    }
}

# Function to show usage
function Show-Usage {
    Write-Host "PullPoet Windows Install Script" -ForegroundColor $Colors.White
    Write-Host ""
    Write-Host "Usage:" -ForegroundColor $Colors.White
    Write-Host "  .\install.ps1                     Install latest version"
    Write-Host "  .\install.ps1 -Update             Update to latest version"
    Write-Host "  .\install.ps1 -Force              Force reinstall current version"
    Write-Host "  .\install.ps1 -Uninstall          Uninstall pullpoet"
    Write-Host "  .\install.ps1 -Help               Show this help"
    Write-Host ""
    Write-Host "Parameters:" -ForegroundColor $Colors.White
    Write-Host "  -InstallDir <path>                Installation directory (default: $DefaultInstallDir)"
    Write-Host ""
    Write-Host "Examples:" -ForegroundColor $Colors.White
    Write-Host "  # Install latest version"
    Write-Host "  Invoke-Expression (Invoke-WebRequest -Uri 'https://raw.githubusercontent.com/erkineren/pullpoet/main/scripts/install.ps1' -UseBasicParsing).Content"
    Write-Host ""
    Write-Host "  # Update to latest version"
    Write-Host "  Invoke-Expression (Invoke-WebRequest -Uri 'https://raw.githubusercontent.com/erkineren/pullpoet/main/scripts/install.ps1' -UseBasicParsing).Content -Update"
    Write-Host ""
    Write-Host "  # Install to custom directory"
    Write-Host "  Invoke-Expression (Invoke-WebRequest -Uri 'https://raw.githubusercontent.com/erkineren/pullpoet/main/scripts/install.ps1' -UseBasicParsing).Content -InstallDir 'C:\Tools\pullpoet'"
}

# Main installation function
function Main {
    Write-Host "🚀 PullPoet Windows Install Script" -ForegroundColor $Colors.White
    Write-Host "=====================================" -ForegroundColor $Colors.White
    Write-Host ""
    
    Write-Info "Target installation directory: $InstallDir"
    
    # Detect architecture
    $architecture = Get-WindowsArchitecture
    Write-Info "Detected architecture: $architecture"
    
    # Get latest version
    Write-Info "Fetching latest release information..."
    $latestVersion = Get-LatestVersion
    
    if (-not $latestVersion) {
        Write-Error "Failed to fetch latest version from GitHub"
        return
    }
    
    Write-Info "Latest version available: v$latestVersion"
    
    # Check currently installed version
    $installedVersion = Get-InstalledVersion
    
    if ($installedVersion -ne "not_installed") {
        Write-Info "Currently installed version: v$installedVersion"
        
        if ($Update) {
            $comparison = Compare-Versions -Version1 $latestVersion -Version2 $installedVersion
            if ($comparison -eq 0) {
                Write-Success "Already running the latest version (v$latestVersion)"
                return
            } elseif ($comparison -gt 0) {
                Write-Info "Updating from v$installedVersion to v$latestVersion"
            } else {
                Write-Warning "Installed version (v$installedVersion) is newer than latest release (v$latestVersion)"
                if (-not $Force) {
                    $response = Read-Host "Do you want to downgrade? (y/N)"
                    if ($response -notmatch '^[Yy]$') {
                        Write-Info "Installation cancelled"
                        return
                    }
                }
            }
        } elseif (-not $Force) {
            $comparison = Compare-Versions -Version1 $latestVersion -Version2 $installedVersion
            if ($comparison -eq 0) {
                Write-Success "PullPoet v$latestVersion is already installed"
                Write-Info "Use -Force to reinstall or -Update to check for updates"
                return
            } else {
                Write-Warning "PullPoet is already installed (v$installedVersion)"
                $response = Read-Host "Do you want to install v$latestVersion? (y/N)"
                if ($response -notmatch '^[Yy]$') {
                    Write-Info "Installation cancelled"
                    return
                }
            }
        }
    }
    
    # Install the binary
    Install-Binary -Version $latestVersion -Architecture $architecture
    
    # Add to PATH
    Add-ToPath
    
    # Verify installation
    Write-Host ""
    Write-Info "Verifying installation..."
    
    $binaryPath = Join-Path $InstallDir $BinaryName
    if (Test-Path $binaryPath) {
        try {
            $finalVersion = & $binaryPath --version 2>$null
            if ($finalVersion -and $finalVersion.Contains($latestVersion)) {
                Write-Success "Installation verified successfully!"
                Write-Host ""
                Write-Host "🎉 PullPoet v$latestVersion is now installed!" -ForegroundColor $Colors.Green
                Write-Host ""
                Write-Host "Quick start:" -ForegroundColor $Colors.White
                Write-Host "  pullpoet --help                    # Show help"
                Write-Host "  pullpoet --version                 # Show version"
                Write-Host ""
                Write-Host "Example usage:" -ForegroundColor $Colors.White
                Write-Host "  `$env:PULLPOET_PROVIDER='openai'"
                Write-Host "  `$env:PULLPOET_MODEL='gpt-3.5-turbo'"
                Write-Host "  `$env:PULLPOET_API_KEY='your-api-key'"
                Write-Host "  pullpoet --target main"
                Write-Host ""
                Write-Host "Learn more: $RepoUrl" -ForegroundColor $Colors.Blue
            } else {
                Write-Warning "Installation completed but version verification failed"
                Write-Info "Expected: v$latestVersion, Got: $finalVersion"
            }
        }
        catch {
            Write-Warning "Installation completed but could not verify version"
        }
    } else {
        Write-Error "Installation completed but binary not found"
        Write-Info "You may need to add $InstallDir to your PATH manually"
    }
}

# Check if running as administrator (optional)
function Test-Administrator {
    $currentUser = [Security.Principal.WindowsIdentity]::GetCurrent()
    $principal = New-Object Security.Principal.WindowsPrincipal($currentUser)
    return $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
}

# Main execution
try {
    # Check PowerShell version
    if ($PSVersionTable.PSVersion.Major -lt 5) {
        Write-Error "PowerShell 5.0 or higher is required"
        return
    }
    
    # Handle parameters
    if ($Help) {
        Show-Usage
        return
    }
    
    if ($Uninstall) {
        Uninstall-PullPoet
        return
    }
    
    # Run main installation
    Main
}
catch {
    Write-Error "An error occurred: $($_.Exception.Message)"
    return
} 