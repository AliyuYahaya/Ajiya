# Ajiya session-start hook for Windows (PowerShell 5.1 or later). Same behaviour
# as session-start.sh: Claude Code adds this script's stdout to the session's
# context.
#   no ajiya.toml in the project    -> print nothing
#   ajiya.toml but no ajiya binary  -> print the install line, nothing else
#   otherwise                       -> print 'ajiya status --brief'
# Every other failure is silent and the exit code is always 0.
$ErrorActionPreference = 'SilentlyContinue'
try {
    if ($env:OS -ne 'Windows_NT') { exit 0 }
    $dir = $env:CLAUDE_PROJECT_DIR
    if (-not $dir) { $dir = '.' }
    Set-Location -LiteralPath $dir
    if (-not (Test-Path -LiteralPath 'ajiya.toml' -PathType Leaf)) { exit 0 }

    if (-not (Get-Command ajiya -CommandType Application)) {
        Write-Output 'Ajiya is not installed. Install it with: go install github.com/AliyuYahaya/Ajiya/cmd/ajiya@latest (npm and Homebrew installs arrive with v0.2)'
        exit 0
    }
    & ajiya status --brief 2>$null
} catch {
}
exit 0
