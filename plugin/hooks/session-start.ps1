# Ajiya session-start hook for Windows (PowerShell 5.1 or later). Same behaviour
# as session-start.sh: Claude Code adds this script's stdout to the session's
# context.
#   no ajiya.toml in the project    -> print nothing
#   ajiya.toml but no ajiya binary  -> print the install line, nothing else
#   otherwise                       -> print 'ajiya status --brief', then one line
#                                      saying to run 'ajiya mcp install' when Ajiya
#                                      is not registered with Claude Code (asked of
#                                      the binary: 'ajiya mcp status --claude --quiet'
#                                      exits 1 for "not registered"; any other
#                                      answer, including an older ajiya, adds nothing)
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
    & ajiya mcp status --claude --quiet 2>&1 | Out-Null
    if ($LASTEXITCODE -eq 1) {
        Write-Output 'Ajiya is not registered with Claude Code, so a session cannot use its tools. Register it with: ajiya mcp install --claude'
    }
} catch {
}
exit 0
