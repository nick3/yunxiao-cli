param(
    [string]$Version = "",
    [string]$BinDir = "",
    [switch]$DryRun,
    [switch]$Uninstall
)

$ErrorActionPreference = "Stop"
$Repo = if ($env:YUNXIAO_INSTALL_REPO) { $env:YUNXIAO_INSTALL_REPO } else { "nick3/yunxiao-cli" }
$ApiUrl = if ($env:YUNXIAO_INSTALL_API_URL) { $env:YUNXIAO_INSTALL_API_URL } else { "https://api.github.com" }
$DownloadBaseUrl = if ($env:YUNXIAO_INSTALL_DOWNLOAD_BASE_URL) { $env:YUNXIAO_INSTALL_DOWNLOAD_BASE_URL } else { "https://github.com" }

if ($PSVersionTable.PSEdition -eq "Desktop") {
    [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12
}

function Write-InstallerInfo {
    param([string]$Message)
    [Console]::Error.WriteLine("yunxiao installer: $Message")
}

function Exit-Installer {
    param([int]$Code, [string]$Message)
    Write-InstallerInfo $Message
    exit $Code
}

function Join-Url {
    param([string]$Base, [string]$Path)
    if ($Base.EndsWith("/")) {
        return "$Base$Path"
    }
    return "$Base/$Path"
}

function Resolve-InstallOS {
    # The PowerShell installer only supports Windows; the override exists for tests.
    $detected = if ($env:YUNXIAO_INSTALL_OS) { $env:YUNXIAO_INSTALL_OS } else { "windows" }
    if ($detected -match "^(windows|Windows)$") {
        return "windows"
    }
    Exit-Installer 2 "unsupported OS: $detected"
}

function Resolve-InstallArch {
    if ($env:YUNXIAO_INSTALL_ARCH) {
        $detected = $env:YUNXIAO_INSTALL_ARCH
    } elseif ($env:PROCESSOR_ARCHITEW6432) {
        $detected = $env:PROCESSOR_ARCHITEW6432
    } elseif ($env:PROCESSOR_ARCHITECTURE) {
        $detected = $env:PROCESSOR_ARCHITECTURE
    } else {
        $detected = [System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture.ToString()
    }

    switch -Regex ($detected) {
        "^(amd64|x86_64|x64|AMD64)$" { return "amd64" }
        "^(arm64|aarch64|ARM64)$" { return "arm64" }
        default { Exit-Installer 2 "unsupported architecture: $detected" }
    }
}

function Resolve-BinDir {
    if ($BinDir) {
        return $BinDir
    }
    if ($env:YUNXIAO_INSTALL_DIR) {
        return $env:YUNXIAO_INSTALL_DIR
    }
    if ($env:USERPROFILE) {
        return (Join-Path $env:USERPROFILE "bin")
    }
    Exit-Installer 5 "USERPROFILE is not set and no install directory was provided"
}

function Get-ErrorMessage {
    param([object]$ErrorRecord)
    if ($ErrorRecord.Exception.Message) {
        return $ErrorRecord.Exception.Message
    }
    return [string]$ErrorRecord
}

function Download-File {
    param([string]$Url, [string]$Destination)
    try {
        Invoke-WebRequest -Uri $Url -OutFile $Destination -UseBasicParsing
    } catch {
        Exit-Installer 3 "failed to download ${Url}: $(Get-ErrorMessage $_)"
    }
}

function Resolve-LatestVersion {
    if ($env:YUNXIAO_INSTALL_LATEST_VERSION) {
        return $env:YUNXIAO_INSTALL_LATEST_VERSION
    }

    $latestUrl = Join-Url $ApiUrl "repos/$Repo/releases/latest"
    try {
        $release = Invoke-RestMethod -Uri $latestUrl -UseBasicParsing
    } catch {
        Exit-Installer 3 "failed to resolve latest stable release from ${latestUrl}: $(Get-ErrorMessage $_)"
    }

    if (-not $release.tag_name) {
        Exit-Installer 3 "failed to parse latest stable release from $latestUrl; pass -Version vX.Y.Z to install a known release"
    }
    return [string]$release.tag_name
}

function Get-Sha256 {
    param([string]$Path)
    return (Get-FileHash -Algorithm SHA256 -Path $Path).Hash.ToLowerInvariant()
}

function Get-ChecksumForAsset {
    param([string]$ChecksumsPath, [string]$AssetName)
    foreach ($line in Get-Content -Path $ChecksumsPath) {
        $parts = $line -split "\s+"
        if ($parts.Length -ge 2 -and $parts[1] -eq $AssetName) {
            return $parts[0].ToLowerInvariant()
        }
    }
    return ""
}

function Test-PathEntry {
    param([string]$Dir)
    return (($env:PATH -split [System.IO.Path]::PathSeparator) -contains $Dir)
}

function Get-PathDiagnostic {
    param([string]$Path)
    if ($Path.Length -le 500) {
        return $Path
    }
    return $Path.Substring(0, 500)
}

function Assert-SafeTarget {
    param([string]$Dir)
    if (-not $Dir) {
        Exit-Installer 7 "install directory is empty"
    }
    $fullDir = [System.IO.Path]::GetFullPath($Dir)
    $root = [System.IO.Path]::GetPathRoot($fullDir)
    $trim = [char[]]@([System.IO.Path]::DirectorySeparatorChar, [System.IO.Path]::AltDirectorySeparatorChar)
    if ($fullDir.TrimEnd($trim) -eq $root.TrimEnd($trim)) {
        Exit-Installer 7 "refusing to uninstall from root directory"
    }
    if (Test-Path -LiteralPath $Dir -PathType Container) {
        $item = Get-Item -LiteralPath $Dir -Force
        if (($item.Attributes -band [System.IO.FileAttributes]::ReparsePoint) -ne 0) {
            Exit-Installer 7 "refusing to uninstall from reparse point directory: $Dir"
        }
        $resolved = (Resolve-Path -LiteralPath $Dir).ProviderPath
        $resolvedRoot = [System.IO.Path]::GetPathRoot($resolved)
        if ($resolved.TrimEnd($trim) -eq $resolvedRoot.TrimEnd($trim)) {
            Exit-Installer 7 "refusing to uninstall from root directory"
        }
    }
}

$os = Resolve-InstallOS
$arch = Resolve-InstallArch
$resolvedBinDir = Resolve-BinDir
$binary = "yunxiao.exe"
$target = Join-Path $resolvedBinDir $binary

if ($Uninstall) {
    Assert-SafeTarget $resolvedBinDir
    if ($DryRun) {
        "action=uninstall"
        "os=$os"
        "arch=$arch"
        "target=$target"
        exit 0
    }

    if (Test-Path -LiteralPath $target -PathType Container) {
        Exit-Installer 7 "refusing to uninstall directory target: $target"
    }
    if (-not (Test-Path -LiteralPath $target)) {
        Write-InstallerInfo "$target is already not installed"
        exit 0
    }
    try {
        Remove-Item -LiteralPath $target -Force
    } catch {
        Exit-Installer 7 "failed to remove ${target}: $(Get-ErrorMessage $_)"
    }
    Write-InstallerInfo "removed $target"
    exit 0
}

if (-not $Version) {
    $Version = Resolve-LatestVersion
}

$asset = "yunxiao_${Version}_${os}_${arch}.zip"
$releasePath = "releases/download/$Version"
$archiveUrl = Join-Url $DownloadBaseUrl "$Repo/$releasePath/$asset"
$checksumsUrl = Join-Url $DownloadBaseUrl "$Repo/$releasePath/checksums.txt"

if ($DryRun) {
    "action=install"
    "os=$os"
    "arch=$arch"
    "version=$Version"
    "asset=$asset"
    "archive_url=$archiveUrl"
    "checksums_url=$checksumsUrl"
    "target=$target"
    exit 0
}

if ((Test-Path -LiteralPath $resolvedBinDir) -and -not (Test-Path -LiteralPath $resolvedBinDir -PathType Container)) {
    Exit-Installer 5 "$resolvedBinDir exists but is not a directory"
}
try {
    New-Item -ItemType Directory -Path $resolvedBinDir -Force | Out-Null
} catch {
    Exit-Installer 5 "failed to create ${resolvedBinDir}: $(Get-ErrorMessage $_)"
}
if (-not (Test-Path -LiteralPath $resolvedBinDir -PathType Container)) {
    Exit-Installer 5 "failed to create $resolvedBinDir"
}
if (Test-Path -LiteralPath $target -PathType Container) {
    Exit-Installer 5 "target is a directory: $target"
}

$tempRoot = Join-Path ([System.IO.Path]::GetTempPath()) ("yunxiao-installer-" + [Guid]::NewGuid().ToString("N"))
try {
    New-Item -ItemType Directory -Path $tempRoot -Force | Out-Null
} catch {
    Exit-Installer 5 "failed to create temporary directory ${tempRoot}: $(Get-ErrorMessage $_)"
}
try {
    $archive = Join-Path $tempRoot $asset
    $checksums = Join-Path $tempRoot "checksums.txt"
    $extractDir = Join-Path $tempRoot "extract"
    try {
        New-Item -ItemType Directory -Path $extractDir -Force | Out-Null
    } catch {
        Exit-Installer 5 "failed to create extract directory ${extractDir}: $(Get-ErrorMessage $_)"
    }

    Download-File $archiveUrl $archive
    Download-File $checksumsUrl $checksums

    $expected = Get-ChecksumForAsset $checksums $asset
    if (-not $expected) {
        Exit-Installer 4 "missing checksum entry for $asset in $checksumsUrl"
    }
    $actual = Get-Sha256 $archive
    if ($actual -ne $expected) {
        Exit-Installer 4 "checksum mismatch for $asset; expected $expected from $checksumsUrl, got $actual from $archiveUrl; existing install was not overwritten"
    }

    try {
        Expand-Archive -LiteralPath $archive -DestinationPath $extractDir -Force
    } catch {
        Exit-Installer 1 "failed to extract ${archive}: $(Get-ErrorMessage $_)"
    }
    $extracted = Join-Path $extractDir $binary
    if (-not (Test-Path -LiteralPath $extracted -PathType Leaf)) {
        Exit-Installer 1 "archive does not contain expected $binary binary"
    }

    $tempTarget = Join-Path $resolvedBinDir (".yunxiao.tmp." + [Guid]::NewGuid().ToString("N") + ".exe")
    try {
        Copy-Item -LiteralPath $extracted -Destination $tempTarget -Force
    } catch {
        Remove-Item -LiteralPath $tempTarget -Force -ErrorAction SilentlyContinue
        Exit-Installer 5 "failed to stage ${target}: $(Get-ErrorMessage $_)"
    }

    try {
        $verifyOutput = & $tempTarget commands --json 2>&1 | Out-String
    } catch {
        Remove-Item -LiteralPath $tempTarget -Force -ErrorAction SilentlyContinue
        Exit-Installer 6 "downloaded yunxiao could not be executed before install: $tempTarget commands --json; error: $(Get-ErrorMessage $_)"
    }
    if ($LASTEXITCODE -ne 0) {
        Remove-Item -LiteralPath $tempTarget -Force -ErrorAction SilentlyContinue
        Exit-Installer 6 "downloaded yunxiao failed verification before install: $tempTarget commands --json; output: $(Get-PathDiagnostic $verifyOutput)"
    }

    try {
        Move-Item -LiteralPath $tempTarget -Destination $target -Force
    } catch {
        Remove-Item -LiteralPath $tempTarget -Force -ErrorAction SilentlyContinue
        Exit-Installer 5 "failed to install ${target}: $(Get-ErrorMessage $_)"
    }

    if (-not (Test-PathEntry $resolvedBinDir)) {
        Write-InstallerInfo "$resolvedBinDir is not on PATH. Add this line to your PowerShell profile, then restart PowerShell: `$env:Path = '$resolvedBinDir;' + `$env:Path"
    }
    Write-InstallerInfo "installed $target"
} finally {
    Remove-Item -LiteralPath $tempRoot -Recurse -Force -ErrorAction SilentlyContinue
}
