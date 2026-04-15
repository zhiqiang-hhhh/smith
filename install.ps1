# install.ps1 — Install smith for the current user on Windows.
# Usage: irm https://raw.githubusercontent.com/zhiqiang-hhhh/smith/main/install.ps1 | iex

$ErrorActionPreference = 'Stop'

$repo = 'zhiqiang-hhhh/smith'
$installDir = Join-Path $env:LOCALAPPDATA 'smith'

# Detect architecture
$arch = if ([Environment]::Is64BitOperatingSystem) {
    if ($env:PROCESSOR_ARCHITECTURE -eq 'ARM64') { 'arm64' } else { 'amd64' }
} else {
    Write-Error 'smith requires a 64-bit system'; exit 1
}

# Find latest release
$release = Invoke-RestMethod "https://api.github.com/repos/$repo/releases/latest"
$asset = $release.assets | Where-Object { $_.name -eq "smith-windows-$arch.zip" }
if (-not $asset) {
    Write-Error "No Windows $arch binary found in release $($release.tag_name)"
    exit 1
}

Write-Host "Installing smith $($release.tag_name) ($arch)..."

# Download and extract
$tmp = Join-Path $env:TEMP "smith-install-$([guid]::NewGuid().ToString('N').Substring(0,8)).zip"
Invoke-WebRequest -Uri $asset.browser_download_url -OutFile $tmp

# Stop any running smith processes before replacing the binary
$procs = Get-Process smith -ErrorAction SilentlyContinue
if ($procs) {
    $procs | Stop-Process -Force -ErrorAction SilentlyContinue
    $procs | Wait-Process -ErrorAction SilentlyContinue
}

if (Test-Path $installDir) {
    Remove-Item $installDir -Recurse -Force
}
New-Item -ItemType Directory -Path $installDir -Force | Out-Null
Expand-Archive -Path $tmp -DestinationPath $installDir -Force
Remove-Item $tmp

# Verify binary exists
$exe = Join-Path $installDir 'smith.exe'
if (-not (Test-Path $exe)) {
    Write-Error "smith.exe not found after extraction"
    exit 1
}

# Add to user PATH if not already present
$userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
if ($userPath -notlike "*$installDir*") {
    [Environment]::SetEnvironmentVariable('Path', "$installDir;$userPath", 'User')
    Write-Host "Added $installDir to user PATH (restart your terminal to take effect)"
}

Write-Host "smith installed to $exe"
Write-Host "Run 'smith' to get started."
