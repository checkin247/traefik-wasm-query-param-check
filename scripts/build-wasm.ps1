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

# Destination for local plugin tree (always create and copy into this path so
# docker-compose mounts pick up the latest build). Honor MODULE_PATH env var.
$modulePath = $env:MODULE_PATH
if ([string]::IsNullOrWhiteSpace($modulePath)) {
    $modulePath = 'github.com/checkin247/traefik-wasm-query-param-check'
}
$pluginsLocal = Join-Path $repoRoot 'plugins-local'
$destDir = Join-Path $pluginsLocal (Join-Path 'src' $modulePath)

# Ensure destination exists so the build can always copy into it
New-Item -ItemType Directory -Force -Path $destDir | Out-Null

# Remove any stale artifact at the repo root so build always produces a fresh file
if (Test-Path $outputPath) {
    Remove-Item -Force $outputPath
}

# Helper: run tinygo locally
function Build-With-TinyGo {
    Write-Host "Using local tinygo..."
    Push-Location $srcDir
    try {
        # Avoid VCS stamping failures inside containers by setting GOFLAGS
        $env:GOFLAGS = '-buildvcs=false'
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
    $args = @(
        'run','--rm',
        '-e','GOFLAGS=-buildvcs=false',
        '-v', "$hostPath`:/work",
        '-w', '/work/src',
        'tinygo/tinygo:0.34.0',
        'tinygo','build', '-o', '/work/plugin.wasm', '-scheduler=none', '--no-debug', '-target=wasi', '.'
    )
    Write-Host "Running: docker $($args -join ' ')"
    & docker @args
}

# Try to locate and run wasm-opt (Binaryen) to optimize the produced wasm. Respects
# the WASMOPT environment variable if set, otherwise looks for 'wasm-opt' in PATH.
function Check-And-Run-WasmOpt {
    param([string]$wasmPath)

    $wasmOptPath = $null
    if ($env:WASMOPT) {
        $wasmOptPath = $env:WASMOPT
    } elseif (Get-Command wasm-opt -ErrorAction SilentlyContinue) {
        $wasmOptPath = (Get-Command wasm-opt).Source
    }

    if ($wasmOptPath) {
        Write-Host "Optimizing wasm with: $wasmOptPath"
        try {
            # run optimizer in-place (write to a temp file then replace)
            $tmp = [System.IO.Path]::GetTempFileName()
            Remove-Item $tmp
            $tmpWasm = "$tmp.wasm"
            & $wasmOptPath -O2 -o $tmpWasm $wasmPath
            if (Test-Path $tmpWasm) {
                Move-Item -Force $tmpWasm $wasmPath
                Write-Host "Optimized plugin.wasm -> $wasmPath"
            } else {
                Write-Warning "wasm-opt completed but optimized file was not produced."
            }
        } catch {
            Write-Warning "wasm-opt run failed: $_"
        }
    } else {
        Write-Warning "wasm-opt not found. Install Binaryen (provides wasm-opt) or set WASMOPT to its path if you want optimized artifacts."
    }
}

# Try local tinygo first
if (Get-Command tinygo -ErrorAction SilentlyContinue) {
    try {
        Build-With-TinyGo
        # try to optimize the output if an optimizer is available
        Check-And-Run-WasmOpt -wasmPath $outputPath
        # Copy artifact and optional Traefik config into the expected local
        # plugin tree so docker-compose mounts can pick up the new build.
        Write-Host "Copying artifacts to: $destDir"
        Copy-Item -Force $outputPath -Destination $destDir
        $traefikSrc = Join-Path $repoRoot '.traefik.yml'
        if (Test-Path $traefikSrc) {
            Copy-Item -Force $traefikSrc -Destination $destDir
        }
        Write-Host "Copied plugin.wasm and any detected traefik config to plugins-local tree"
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
        # try to optimize the output if an optimizer is available on host
        Check-And-Run-WasmOpt -wasmPath $outputPath
        if (-not (Test-Path $outputPath)) {
            Write-Error "Build completed but $outputPath not found"
            exit 1
        }

        Write-Host "Copying artifacts to: $destDir"
        Copy-Item -Force $outputPath -Destination $destDir
        $traefikSrc = Join-Path $repoRoot '.traefik.yml'
        if (Test-Path $traefikSrc) {
            Copy-Item -Force $traefikSrc -Destination $destDir
        }
        Write-Host "Copied plugin.wasm and any detected traefik config to plugins-local tree"
        exit 0
    } catch {
        Write-Error "Docker tinygo build failed: $_"
        exit 1
    }
} else {
    Write-Error "Neither tinygo nor docker found in PATH. Please install tinygo or docker and retry."
    exit 1
}
