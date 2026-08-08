$ErrorActionPreference = 'Stop'
$toolsDir = Split-Path -Parent $MyInvocation.MyCommand.Definition
$version = '3.2.1'
$url = "https://github.com/bolens/millennium-helpers/releases/download/v$version/millennium-helpers-v$version-windows-amd64.zip"
$checksum = '0856c9befcef3c060805bbc2b0fa6cc250782dbc8ac333af757f73f5da0fca3e'

$packageArgs = @{
  packageName   = 'millennium-helpers'
  unzipLocation = Join-Path $toolsDir 'payload'
  url           = $url
  checksum      = $checksum
  checksumType  = 'sha256'
}

Install-ChocolateyZipPackage @packageArgs

$payload = $packageArgs.unzipLocation
$winScripts = Join-Path $payload 'scripts\windows'
if (!(Test-Path -LiteralPath $winScripts)) {
  throw "Unexpected archive layout: missing scripts\windows under $payload"
}

# Require Go .exe when the release zip embeds it.
$millenniumExe = Join-Path $winScripts 'millennium.exe'
if (Test-Path -LiteralPath $millenniumExe) {
  Install-ChocolateyPath -PathToInstall $winScripts -PathType 'User'
  Install-BinFile -Name 'millennium' -Path $millenniumExe
} else {
  throw 'millennium.exe (Go dispatcher) not found in release zip'
}
