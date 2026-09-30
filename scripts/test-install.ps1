# Tests scripts/install.ps1 against a local fixture release. Run on Windows.
#
#   pwsh -File scripts/test-install.ps1        (or powershell -File ...)
#
# Environment:
#   AJIYA_DIST  folder holding ajiya_<version>_windows_<arch>.zip and checksums.txt
#               from a GoReleaser snapshot (default: dist\, built here with goreleaser if missing)
#   INSTALL_PS1 script under test (default: install.ps1 next to this file)
#   CI          when set, also checks the user PATH change (it edits the real user PATH
#               and puts it back, so leave it unset on a personal machine)
#
# Needs python to serve the fixture over HTTP. Exit status 0 when every check passes.

$ErrorActionPreference = 'Stop'
$here = Split-Path -Parent $MyInvocation.MyCommand.Path
$root = Split-Path -Parent $here
$installPs1 = if ($env:INSTALL_PS1) { $env:INSTALL_PS1 } else { Join-Path $here 'install.ps1' }
$dist = if ($env:AJIYA_DIST) { $env:AJIYA_DIST } else { Join-Path $root 'dist' }
$script:fails = 0

function Check([string]$Name, [bool]$Ok) {
    if ($Ok) { Write-Host "ok   $Name" } else { Write-Host "FAIL $Name"; $script:fails++ }
}

$arch = if ($env:PROCESSOR_ARCHITECTURE -eq 'ARM64') { 'arm64' } else { 'amd64' }
if (-not (Get-ChildItem -Path $dist -Filter "ajiya_*_windows_$arch.zip" -ErrorAction SilentlyContinue)) {
    $env:HOMEBREW_TAP_TOKEN = 'unused'
    Push-Location $root
    try { goreleaser release --snapshot --clean --skip=publish } finally { Pop-Location }
}
$zip = Get-ChildItem -Path $dist -Filter "ajiya_*_windows_$arch.zip" | Select-Object -First 1
$version = $zip.Name -replace '^ajiya_(.+)_windows_.+\.zip$', '$1'
$tag = "v$version"
Write-Host "fixture: $($zip.Name) (tag $tag)"

$work = Join-Path ([IO.Path]::GetTempPath()) ('ajiya-test-install-' + [Guid]::NewGuid().ToString('N'))
$site = Join-Path $work 'site'
$good = Join-Path $site "releases\download\$tag"
$bad = Join-Path $site "tampered\$tag"
New-Item -ItemType Directory -Path $good, $bad | Out-Null
Copy-Item (Join-Path $dist 'ajiya_*.zip'), (Join-Path $dist 'checksums.txt') -Destination $good
Copy-Item (Join-Path $good '*') -Destination $bad
[IO.File]::AppendAllText((Join-Path $bad $zip.Name), 'tampered')

$port = Get-Random -Minimum 20000 -Maximum 40000
$server = Start-Process -FilePath python -ArgumentList '-m', 'http.server', $port, '--bind', '127.0.0.1', '--directory', $site -PassThru -WindowStyle Hidden
$originalUserPath = [Environment]::GetEnvironmentVariable('Path', 'User')
$originalSessionPath = $env:Path
try {
    Start-Sleep -Seconds 2
    $env:AJIYA_NO_PROMPT = '1'
    $env:AJIYA_BASE_URL = "http://127.0.0.1:$port/releases/download"
    $env:AJIYA_VERSION = $tag
    $installDir = Join-Path $work 'bin'
    $env:AJIYA_INSTALL_DIR = $installDir
    if (-not $env:CI) { $env:AJIYA_NO_PATH = '1' }

    # Run in a child process so a thrown error or exit cannot end this test.
    function Run-Installer([string[]]$Args1) {
        $shell = if (Get-Command pwsh -ErrorAction SilentlyContinue) { 'pwsh' } else { 'powershell' }
        & $shell -NoProfile -ExecutionPolicy Bypass -File $installPs1 @Args1 2>&1 | Out-Host
        return $LASTEXITCODE
    }

    Check 'install exits 0' ((Run-Installer @()) -eq 0)
    $exe = Join-Path $installDir 'ajiya.exe'
    Check 'ajiya.exe exists' (Test-Path $exe)
    Check "ajiya version reports $version" ((& $exe version 2>&1 | Out-String) -like "*$version*")
    if ($env:CI) {
        $userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
        Check 'install dir is on the user PATH' (($userPath.Split(';') | ForEach-Object { $_.TrimEnd('\') }) -contains $installDir.TrimEnd('\'))
    }

    Check 'uninstall exits 0' ((Run-Installer @('-Uninstall')) -eq 0)
    Check 'ajiya.exe removed' (-not (Test-Path $exe))
    if ($env:CI) {
        $userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
        Check 'install dir removed from the user PATH' (-not (($userPath.Split(';') | ForEach-Object { $_.TrimEnd('\') }) -contains $installDir.TrimEnd('\')))
    }

    # A tampered archive is refused and nothing is installed.
    $env:AJIYA_BASE_URL = "http://127.0.0.1:$port/tampered"
    Check 'tampered archive is refused' ((Run-Installer @()) -ne 0)
    Check 'nothing installed after a mismatch' (-not (Test-Path $exe))

    # An unknown version fails.
    $env:AJIYA_BASE_URL = "http://127.0.0.1:$port/releases/download"
    $env:AJIYA_VERSION = 'v9.9.9'
    Check 'unknown version fails' ((Run-Installer @()) -ne 0)
} finally {
    if ($server) { Stop-Process -Id $server.Id -Force -ErrorAction SilentlyContinue }
    if ($env:CI) { [Environment]::SetEnvironmentVariable('Path', $originalUserPath, 'User') }
    $env:Path = $originalSessionPath
    Remove-Item -LiteralPath $work -Recurse -Force -ErrorAction SilentlyContinue
}

if ($script:fails -eq 0) { Write-Host 'all install.ps1 checks passed' } else { Write-Host "$($script:fails) check(s) failed"; exit 1 }
