$ErrorActionPreference = "Stop"

$repo = "ErikHellman/hemkop-cli"
$binary = "hemkop.exe"

# Detect architecture
$arch = if ([Environment]::Is64BitOperatingSystem) {
    if ($env:PROCESSOR_ARCHITECTURE -eq "ARM64" -or $env:PROCESSOR_IDENTIFIER -match "ARM") { "arm64" } else { "amd64" }
} else {
    Write-Error "32-bit Windows is not supported"
    exit 1
}

# Get latest release tag
$release = Invoke-RestMethod -Uri "https://api.github.com/repos/$repo/releases/latest"
$tag = $release.tag_name

$url = "https://github.com/$repo/releases/download/$tag/hemkop-windows-${arch}.exe"
$installDir = "$env:LOCALAPPDATA\hemkop"

if (-not (Test-Path $installDir)) {
    New-Item -ItemType Directory -Path $installDir | Out-Null
}

$dest = Join-Path $installDir $binary
Write-Host "Downloading hemkop $tag for windows/$arch..."
Invoke-WebRequest -Uri $url -OutFile $dest

# Add to PATH if not already there
$currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($currentPath -notlike "*$installDir*") {
    [Environment]::SetEnvironmentVariable("Path", "$currentPath;$installDir", "User")
    Write-Host "Added $installDir to PATH (restart your terminal to use)"
}

Write-Host "hemkop installed to $dest"
