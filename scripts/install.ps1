# Ajiya installer for Windows (PowerShell 5.1 or later).
#
#   irm https://github.com/AliyuYahaya/Ajiya/releases/latest/download/install.ps1 | iex
#
# Downloads the release zip for this machine, checks it against the release's
# checksums.txt, installs ajiya.exe into a folder you own and adds that folder
# to your user PATH. It never needs administrator rights.
#
# To uninstall (removes ajiya.exe and its PATH entry):
#   & ([scriptblock]::Create((irm https://github.com/AliyuYahaya/Ajiya/releases/latest/download/install.ps1))) -Uninstall
#
# Environment variables:
#   AJIYA_VERSION      version to install, for example v0.2.0 (default: latest release)
#   AJIYA_INSTALL_DIR  install folder (default: $env:LOCALAPPDATA\Programs\ajiya\bin)
#   AJIYA_NO_PROMPT=1  never ask questions
#   AJIYA_NO_PATH=1    do not change the user PATH
#   AJIYA_BASE_URL     release download base (default: the GitHub releases URL);
#                      files are read from $AJIYA_BASE_URL/<tag>/
#   AJIYA_LATEST_URL   page that redirects to the latest tag
#                      (default: https://github.com/AliyuYahaya/Ajiya/releases/latest)

param(
    [switch]$Uninstall,
    [switch]$Help
)

$RepoUrl = 'https://github.com/AliyuYahaya/Ajiya'

function Get-Setting([string]$Name, [string]$Default) {
    $value = [Environment]::GetEnvironmentVariable($Name)
    if ([string]::IsNullOrEmpty($value)) { return $Default }
    return $value
}

function Show-Usage {
    @'
Install Ajiya.

Usage: install.ps1 [-Uninstall] [-Help]

  -Uninstall  remove ajiya.exe and its PATH entry
  -Help       show this help

Environment variables:
  AJIYA_VERSION      version to install, for example v0.2.0 (default: latest)
  AJIYA_INSTALL_DIR  install folder (default: %LOCALAPPDATA%\Programs\ajiya\bin)
  AJIYA_NO_PROMPT=1  never ask questions
  AJIYA_NO_PATH=1    do not change the user PATH
  AJIYA_BASE_URL     release download base URL (for mirrors and tests)
'@
}

function Get-InstallDir {
    $dir = Get-Setting 'AJIYA_INSTALL_DIR' ''
    if ($dir -ne '') { return $dir }
    if ([string]::IsNullOrEmpty($env:LOCALAPPDATA)) { throw 'LOCALAPPDATA is not set; set AJIYA_INSTALL_DIR' }
    return (Join-Path $env:LOCALAPPDATA 'Programs\ajiya\bin')
}

function Get-Arch {
    # A 32-bit PowerShell on 64-bit Windows reports x86 here; the real one is in PROCESSOR_ARCHITEW6432.
    $arch = $env:PROCESSOR_ARCHITEW6432
    if ([string]::IsNullOrEmpty($arch)) { $arch = $env:PROCESSOR_ARCHITECTURE }
    switch ($arch) {
        'AMD64' { return 'amd64' }
        'ARM64' { return 'arm64' }
        default { throw "unsupported CPU architecture '$arch'" }
    }
}

function Get-LatestTag {
    $latest = Get-Setting 'AJIYA_LATEST_URL' "$RepoUrl/releases/latest"
    $response = Invoke-WebRequest -Uri $latest -UseBasicParsing
    # Windows PowerShell 5.1 and PowerShell 7 expose the final address differently.
    $final = $null
    if ($response.BaseResponse.PSObject.Properties['ResponseUri']) {
        $final = $response.BaseResponse.ResponseUri.AbsoluteUri
    } elseif ($response.BaseResponse.PSObject.Properties['RequestMessage']) {
        $final = $response.BaseResponse.RequestMessage.RequestUri.AbsoluteUri
    }
    if ($final -and $final -match '/tag/(.+)$') { return $Matches[1] }
    throw 'could not find the latest release; set AJIYA_VERSION, for example $env:AJIYA_VERSION = "v0.2.0"'
}

function Get-NormalizedDir([string]$Path) {
    return $Path.TrimEnd('\', '/')
}

function Test-OnUserPath([string]$Dir) {
    $userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
    if ([string]::IsNullOrEmpty($userPath)) { return $false }
    $want = Get-NormalizedDir $Dir
    foreach ($entry in $userPath.Split(';')) {
        if ($entry -ne '' -and (Get-NormalizedDir $entry) -ieq $want) { return $true }
    }
    return $false
}

function Add-ToUserPath([string]$Dir) {
    if ((Get-Setting 'AJIYA_NO_PATH' '') -eq '1') { return }
    if (-not (Test-OnUserPath $Dir)) {
        $userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
        if ([string]::IsNullOrEmpty($userPath)) { $new = $Dir } else { $new = $userPath.TrimEnd(';') + ';' + $Dir }
        [Environment]::SetEnvironmentVariable('Path', $new, 'User')
        Write-Host "Added $Dir to your user PATH. Open a new terminal to use it."
    }
    # Make ajiya usable in this session too.
    if (-not (($env:Path.Split(';') | ForEach-Object { Get-NormalizedDir $_ }) -contains (Get-NormalizedDir $Dir))) {
        $env:Path = $env:Path.TrimEnd(';') + ';' + $Dir
    }
}

function Remove-FromUserPath([string]$Dir) {
    if ((Get-Setting 'AJIYA_NO_PATH' '') -eq '1') { return }
    $userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
    if ([string]::IsNullOrEmpty($userPath)) { return }
    $want = Get-NormalizedDir $Dir
    $kept = @($userPath.Split(';') | Where-Object { $_ -ne '' -and (Get-NormalizedDir $_) -ine $want })
    if ($kept.Count -ne @($userPath.Split(';') | Where-Object { $_ -ne '' }).Count) {
        [Environment]::SetEnvironmentVariable('Path', ($kept -join ';'), 'User')
        Write-Host "Removed $Dir from your user PATH."
    }
}

function Test-CanPrompt {
    if ((Get-Setting 'AJIYA_NO_PROMPT' '') -eq '1') { return $false }
    if (-not [Environment]::UserInteractive) { return $false }
    try { return (-not [Console]::IsInputRedirected) } catch { return $false }
}

function Invoke-Uninstall {
    $dir = Get-InstallDir
    $exe = Join-Path $dir 'ajiya.exe'
    if (Test-Path -LiteralPath $exe) {
        Remove-Item -LiteralPath $exe -Force
        Write-Host "Removed $exe"
    } else {
        Write-Host "Nothing to remove: $exe does not exist"
    }
    Remove-FromUserPath $dir
    Write-Host "Your projects and agent settings were not touched. To unregister from agents, run 'ajiya mcp uninstall' before removing the binary."
}

function Invoke-Install {
    # Windows PowerShell 5.1 does not use TLS 1.2 by default, which GitHub requires.
    [Net.ServicePointManager]::SecurityProtocol = [Net.ServicePointManager]::SecurityProtocol -bor [Net.SecurityProtocolType]::Tls12

    $arch = Get-Arch
    $dir = Get-InstallDir

    $tag = Get-Setting 'AJIYA_VERSION' ''
    if ($tag -eq '') {
        Write-Host 'Looking up the latest release'
        $tag = Get-LatestTag
    }
    if (-not $tag.StartsWith('v')) { $tag = "v$tag" }
    $version = $tag.Substring(1)

    $base = (Get-Setting 'AJIYA_BASE_URL' "$RepoUrl/releases/download").TrimEnd('/')
    $archive = "ajiya_${version}_windows_${arch}.zip"

    $tmp = Join-Path ([IO.Path]::GetTempPath()) ('ajiya-install-' + [Guid]::NewGuid().ToString('N'))
    New-Item -ItemType Directory -Path $tmp | Out-Null
    try {
        Write-Host "Downloading Ajiya $tag for windows/$arch"
        Invoke-WebRequest -Uri "$base/$tag/$archive" -OutFile (Join-Path $tmp $archive) -UseBasicParsing
        Invoke-WebRequest -Uri "$base/$tag/checksums.txt" -OutFile (Join-Path $tmp 'checksums.txt') -UseBasicParsing

        # Each line of checksums.txt is "<sha256>  <file name>".
        $want = $null
        foreach ($line in Get-Content -LiteralPath (Join-Path $tmp 'checksums.txt')) {
            $parts = $line.Trim() -split '\s+', 2
            if ($parts.Count -eq 2 -and $parts[1].TrimStart('*') -eq $archive) { $want = $parts[0]; break }
        }
        if (-not $want) { throw "checksums.txt has no entry for $archive" }
        $got = (Get-FileHash -LiteralPath (Join-Path $tmp $archive) -Algorithm SHA256).Hash
        if ($got -ine $want) {
            throw "checksum mismatch for $archive`n  expected $want`n  got      $got`nrefusing to install; nothing was changed"
        }
        Write-Host 'Checksum verified'

        $extract = Join-Path $tmp 'x'
        Expand-Archive -LiteralPath (Join-Path $tmp $archive) -DestinationPath $extract -Force
        $exeSource = Join-Path $extract 'ajiya.exe'
        if (-not (Test-Path -LiteralPath $exeSource)) { throw 'the archive does not contain ajiya.exe' }

        New-Item -ItemType Directory -Path $dir -Force | Out-Null
        $exe = Join-Path $dir 'ajiya.exe'
        Copy-Item -LiteralPath $exeSource -Destination $exe -Force
    } finally {
        Remove-Item -LiteralPath $tmp -Recurse -Force -ErrorAction SilentlyContinue
    }

    Write-Host "Installed Ajiya $tag to $exe"
    Add-ToUserPath $dir

    if (Test-CanPrompt) {
        Write-Host ''
        $answer = Read-Host 'Register Ajiya with the coding agents found on this machine (ajiya mcp install)? [y/N]'
        if ($answer -match '^(y|yes)$') {
            & $exe mcp install
            if ($LASTEXITCODE -ne 0) { Write-Warning 'ajiya mcp install failed; run it yourself later' }
        } else {
            Write-Host "Skipped. Run 'ajiya mcp install' whenever you want to register."
        }
    } else {
        Write-Host ''
        Write-Host 'To register Ajiya with your coding agents, run: ajiya mcp install'
    }
}

function Main {
    if ($Help) { Show-Usage; return }
    # Run through iex this script shares the caller's session, so put back what it changed.
    $savedError = $ErrorActionPreference
    $savedProgress = $ProgressPreference
    $ErrorActionPreference = 'Stop'
    $ProgressPreference = 'SilentlyContinue'   # the progress bar makes Windows PowerShell 5.1 downloads very slow
    try {
        if ($Uninstall) { Invoke-Uninstall } else { Invoke-Install }
    } catch {
        # Make a failure unmistakable. Run as a file (-File, CI, scripts), exit 1
        # so the caller sees it; run through 'irm | iex', re-throw instead, since
        # exit would close the user's PowerShell window.
        [Console]::Error.WriteLine("ajiya install: $($_.Exception.Message)")
        if ($PSCommandPath) { $ErrorActionPreference = $savedError; $ProgressPreference = $savedProgress; exit 1 }
        throw
    } finally {
        $ErrorActionPreference = $savedError
        $ProgressPreference = $savedProgress
    }
}

Main
