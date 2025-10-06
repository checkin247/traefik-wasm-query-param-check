# Cross-platform helper to build plugin.wasm using TinyGo or Docker
# Usage (PowerShell):
#   ./scripts/build-wasm.ps1

param()
$ErrorActionPreference = 'Stop'

Write-Host "Building plugin.wasm using TinyGo (local) or Docker (fallback)"

# repo root assumed current directory
$repoRoot = Get-Location
$srcDir = Join-Path $repoRoot 'src'
$outputPath = Join-Path $repoRoot 'plugin.wasm'

# Helper: run tinygo locally
function Build-With-TinyGo {
    Write-Host "Using local tinygo..."
    Push-Location $srcDir
    try {
        tinygo build -o $outputPath -scheduler=none --no-debug -target=wasi .
    } finally {
        Pop-Location
    }
}

# Helper: run tinygo via Docker
function Build-With-Docker {
    Write-Host "Using Docker tinygo image..."
    # Use absolute path for the host mount
    $hostPath = $repoRoot.ProviderPath

    # Note: Docker on Windows may require path adjustments depending on your setup
    # e.g. using WSL: $(wslpath -a "$hostPath") or prefixing with /host_mnt/<drive>
    $dockerCmd = "docker run --rm -v `"$hostPath`":/work -w /work/src tinygo/tinygo:0.34.0 tinygo build -o /work/plugin.wasm -scheduler=none --no-debug -target=wasi ."
    Write-Host "Running: $dockerCmd"
    iex $dockerCmd
}

# Try local tinygo first
if (Get-Command tinygo -ErrorAction SilentlyContinue) {
    try {
        Build-With-TinyGo
        Write-Host "Built plugin.wasm -> $outputPath"
        exit 0
    } catch {
        Write-Warning "Local tinygo build failed: $_"
        Write-Host "Falling back to Docker..."
    }
}

# Fallback to docker
if (Get-Command docker -ErrorAction SilentlyContinue) {
    try {
        Build-With-Docker
        Write-Host "Built plugin.wasm -> $outputPath"
        exit 0
    } catch {
        Write-Error "Docker tinygo build failed: $_"
        exit 1
    }
} else {
    Write-Error "Neither tinygo nor docker found in PATH. Please install tinygo or docker and retry."
    exit 1
}
