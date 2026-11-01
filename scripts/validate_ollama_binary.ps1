# Copyright 2026 The Ollama Authors.
# Licensed under the Apache License, Version 2.0.
# See the LICENSE file in the repository root for license information.

<#
.SYNOPSIS
  Read-only verification of an Ollama Go binary's SHA-256 and embedded build metadata.

.DESCRIPTION
  This script does NOT execute the Ollama binary, install software, download files,
  contact a network service, or modify the inspected file. It calculates SHA-256
  and asks the locally installed Go tool to read the binary's embedded build info.

  A successful result only confirms that the inspected file advertises the
  expected Go toolchain and selected module versions in its build metadata.
  It does not prove provenance, runtime safety, reachability of a vulnerable
  code path, or the contents of a container image.

.EXAMPLE
  .\scripts\validate_ollama_binary.ps1 -BinaryPath 'C:\Program Files\Ollama\ollama.exe'
#>

[CmdletBinding()]
param(
    [Parameter(Mandatory = $false)]
    [string]$BinaryPath
)

$ErrorActionPreference = 'Stop'
$MinimumGoVersion = [version]'1.26.10'

if ([string]::IsNullOrWhiteSpace($BinaryPath)) {
    $ollamaCommand = Get-Command 'ollama.exe' -ErrorAction SilentlyContinue
    if (-not $ollamaCommand) {
        $ollamaCommand = Get-Command 'ollama' -ErrorAction SilentlyContinue
    }
    if (-not $ollamaCommand -or -not $ollamaCommand.Source) {
        throw 'Ollama no aparece en PATH. Indica -BinaryPath con la ruta completa del ejecutable.'
    }
    $BinaryPath = $ollamaCommand.Source
}

$resolvedPath = (Resolve-Path -LiteralPath $BinaryPath -ErrorAction Stop).Path
if (-not (Test-Path -LiteralPath $resolvedPath -PathType Leaf)) {
    throw "La ruta no es un fichero: $resolvedPath"
}

$hash = Get-FileHash -LiteralPath $resolvedPath -Algorithm SHA256
Write-Output "Binary: $resolvedPath"
Write-Output "SHA256: $($hash.Hash.ToLowerInvariant())"

$goCommand = Get-Command 'go.exe' -ErrorAction SilentlyContinue
if (-not $goCommand) {
    $goCommand = Get-Command 'go' -ErrorAction SilentlyContinue
}
if (-not $goCommand -or -not $goCommand.Source) {
    throw 'No encuentro Go en PATH. Instala Go 1.26.10 o posterior desde la web oficial y vuelve a ejecutar.'
}

Write-Output ''
Write-Output '--- Embedded Go build information (read-only) ---'
$buildInfo = & $goCommand.Source version -m $resolvedPath 2>&1
$buildInfoExitCode = $LASTEXITCODE
$buildInfo | ForEach-Object { Write-Output $_ }
if ($buildInfoExitCode -ne 0) {
    throw "go version -m terminó con código $buildInfoExitCode"
}

$infoText = ($buildInfo | Out-String)
$versionMatch = [regex]::Match(
    $infoText,
    '(?m)^.*:\s*go([0-9]+\.[0-9]+(?:\.[0-9]+)?)\s*$'
)
if (-not $versionMatch.Success) {
    throw 'No se pudo identificar la versión Go embebida en el binario.'
}

$embeddedGoVersion = [version]$versionMatch.Groups[1].Value
$toolchainPass = $embeddedGoVersion -ge $MinimumGoVersion

$requirements = [ordered]@{
    'github.com/buger/jsonparser' = 'v1.1.2'
    'golang.org/x/crypto' = 'v0.58.0'
    'golang.org/x/net' = 'v0.61.0'
    'golang.org/x/text' = 'v0.43.0'
}

$results = foreach ($module in $requirements.Keys) {
    $escapedModule = [regex]::Escape($module)
    $moduleMatch = [regex]::Match(
        $infoText,
        "(?m)^\s*dep\s+$escapedModule\s+(\S+)"
    )
    $actualVersion = if ($moduleMatch.Success) {
        $moduleMatch.Groups[1].Value
    } else {
        'no incluido/no visible'
    }

    [pscustomobject]@{
        Module   = $module
        Expected = $requirements[$module]
        Actual   = $actualVersion
        Pass     = ($actualVersion -eq $requirements[$module])
    }
}

Write-Output ''
Write-Output "--- Dependency checks (embedded Go: $embeddedGoVersion) ---"
$results | Format-Table -AutoSize | Out-String | Write-Output

$failedModules = @($results | Where-Object { -not $_.Pass })
if (-not $toolchainPass -or $failedModules.Count -gt 0) {
    Write-Output 'RESULT: NOT VERIFIED - toolchain or dependency metadata does not meet the expected baseline.'
    exit 2
}

Write-Output 'RESULT: selected build metadata matches the expected baseline.'
Write-Output 'This is not a provenance attestation, SBOM scan, vulnerability reachability analysis, or container-image verification.'
exit 0
