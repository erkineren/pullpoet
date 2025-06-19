param (
    [switch]$Update,
    [switch]$Force,
    [switch]$Uninstall
)

Write-Host "🚀 PullPoet Windows Install Script"
Write-Host "====================================="

$installDir = "$env:LOCALAPPDATA\Programs\pullpoet"
Write-Host "[INFO] Target installation directory: $installDir"

$arch = if ([System.Environment]::Is64BitOperatingSystem) { "x86_64" } else { "x86" }
Write-Host "[INFO] Detected architecture: $arch"

$repo = "erkineren/pullpoet"
$latestReleaseUrl = "https://api.github.com/repos/$repo/releases/latest"
$installedVersionFile = "$installDir\VERSION"

function Get-LatestVersion {
    try {
        $release = Invoke-RestMethod -Uri $latestReleaseUrl -UseBasicParsing
        return $release.tag_name
    } catch {
        Write-Host "[ERROR] Failed to fetch latest version"
        exit 1
    }
}

function Get-InstalledVersion {
    if (Test-Path $installedVersionFile) {
        return Get-Content $installedVersionFile -Raw
    }
    return $null
}

function Uninstall-PullPoet {
    if (Test-Path $installDir) {
        Remove-Item -Recurse -Force $installDir
        Write-Host "[SUCCESS] PullPoet has been uninstalled."
    } else {
        Write-Host "[INFO] PullPoet is not installed."
    }
    exit 0
}

if ($Uninstall) {
    Uninstall-PullPoet
}

$latestVersion = Get-LatestVersion
$installedVersion = Get-InstalledVersion

Write-Host "[INFO] Latest version available: $latestVersion"
Write-Host "[INFO] Currently installed version: $installedVersion"

if ($installedVersion -eq $latestVersion -and -not $Force -and -not $Update) {
    Write-Host "[SUCCESS] PullPoet $latestVersion is already installed"
    Write-Host "[INFO] Use -Force to reinstall or -Update to check for updates"
    exit 0
}

Write-Host "[INFO] Installing PullPoet $latestVersion..."

$downloadUrl = "https://github.com/$repo/releases/download/$latestVersion/pullpoet-windows-$arch.exe"
$exePath = "$installDir\pullpoet.exe"

if (-not (Test-Path $installDir)) {
    New-Item -ItemType Directory -Path $installDir -Force | Out-Null
}

try {
    Invoke-WebRequest -Uri $downloadUrl -OutFile $exePath -UseBasicParsing -ErrorAction Stop
} catch {
    Write-Host "[ERROR] Failed to download executable from $downloadUrl"
    exit 1
}

Set-Content -Path $installedVersionFile -Value $latestVersion

Write-Host "[SUCCESS] PullPoet $latestVersion installed to $exePath"
exit 0
