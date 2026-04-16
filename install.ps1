#!/usr/bin/env pwsh
$ErrorActionPreference = 'Stop'

$Repo       = 'bensch98/arcane'
$Binary     = 'arcane.exe'
$InstallDir = Join-Path $env:LOCALAPPDATA 'Programs\arcane'

$arch = switch ($env:PROCESSOR_ARCHITECTURE) {
    'AMD64' { 'amd64' }
    'ARM64' { 'arm64' }
    default { throw "Unsupported architecture: $env:PROCESSOR_ARCHITECTURE" }
}

$asset = "arcane-windows-$arch.exe"
$url   = "https://github.com/$Repo/releases/latest/download/$asset"

New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
$target = Join-Path $InstallDir $Binary

Write-Host "Downloading $asset..."
Invoke-WebRequest -Uri $url -OutFile $target -UseBasicParsing

$userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
if (($userPath -split ';') -notcontains $InstallDir) {
    $newPath = if ([string]::IsNullOrEmpty($userPath)) { $InstallDir } else { "$userPath;$InstallDir" }
    [Environment]::SetEnvironmentVariable('Path', $newPath, 'User')
    Write-Host "Added $InstallDir to user PATH. Restart your shell to pick it up."
}

Write-Host "Installed arcane to $target"
try { & $target version } catch { }
