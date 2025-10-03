<#
Simple local integration test for the Traefik qptoken plugin.

This script waits until the Traefik entrypoint is reachable, then sends HTTP
requests to the root path with "good" and "bad" Token query parameters and
verifies the response status codes.

Usage (PowerShell Core / pwsh):
    pwsh ./tests/local-integration-test.ps1 -TargetHost localhost -Port 80 -StartTimeoutSec 120

Exit codes:
  0 - all tests passed
  1 - one or more tests failed
  2 - timed out waiting for service
#>

param(
    [string]$TargetHost = 'localhost',
    [int]$Port = 80,
    [int]$StartTimeoutSec = 120
)

$base = "http://$TargetHost`:$Port/"

function Wait-ForHttp {
    param(
        [string]$Uri,
        [int]$TimeoutSec
    )
    $client = [System.Net.Http.HttpClient]::new()
    $deadline = [DateTime]::UtcNow.AddSeconds($TimeoutSec)
    while ([DateTime]::UtcNow -lt $deadline) {
        try {
            $resp = $client.GetAsync($Uri).Result
            # If we get any response (200 or 401) consider the proxy up
            if ($resp -ne $null) { return $true }
        } catch {
            # ignore and retry
        }
        Start-Sleep -Seconds 1
    }
    return $false
}

Write-Host "Waiting for $base to be reachable (timeout $StartTimeoutSec s)..."
if (-not (Wait-ForHttp -Uri $base -TimeoutSec $StartTimeoutSec)) {
    Write-Error "Timed out waiting for $base"
    exit 2
}
Write-Host "Host $TargetHost is reachable. Running tests..."

$client = [System.Net.Http.HttpClient]::new()

# These tokens mirror the allowedValues in traefik/dynamic.yml
$goodTokens = @('my-secret', 'another-secret')
# A few bad token scenarios: missing token is tested separately, plus invalid values
$badTokens = @('invalid-secret', ' ')

$failures = @()

foreach ($t in $goodTokens) {
    $uri = $base + "?Token=" + [System.Uri]::EscapeDataString($t)
    try {
        $resp = $client.GetAsync($uri).Result
        $code = [int]$resp.StatusCode
    } catch {
        $code = -1
    }
    Write-Host "GOOD token '$t' -> $code"
    if ($code -ne 200) {
        $failures += "GOOD token '$t' expected 200 got $code"
    }
}

# Missing token should be denied per dynamic.yml (denyStatus: 401)
try {
    $resp = $client.GetAsync($base).Result
    $code = [int]$resp.StatusCode
} catch {
    $code = -1
}
Write-Host "NO token -> $code"
if ($code -ne 401) { $failures += "NO token expected 401 got $code" }

foreach ($t in $badTokens) {
    $uri = $base + "?Token=" + [System.Uri]::EscapeDataString($t)
    try {
        $resp = $client.GetAsync($uri).Result
        $code = [int]$resp.StatusCode
    } catch {
        $code = -1
    }
    Write-Host "BAD token '$t' -> $code"
    if ($code -ne 401) {
        $failures += "BAD token '$t' expected 401 got $code"
    }
}

if ($failures.Count -gt 0) {
    Write-Host "`nFailures:`n"
    $failures | ForEach-Object { Write-Host "- $_" }
    exit 1
} else {
    Write-Host "`nAll tests passed."
    exit 0
}
