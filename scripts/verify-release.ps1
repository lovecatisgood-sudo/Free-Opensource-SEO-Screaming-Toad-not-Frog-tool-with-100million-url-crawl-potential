param(
  [string]$ReleaseDirectory = ".",
  [string]$Repository = "lovecatisgood-sudo/Free-Opensource-SEO-Screaming-Toad-not-Frog-tool-with-100million-url-crawl-potential"
)
$ErrorActionPreference = "Stop"
Set-Location $ReleaseDirectory
if (-not (Test-Path "SHA256SUMS")) { throw "SHA256SUMS is missing" }
Get-Content "SHA256SUMS" | ForEach-Object {
  if ($_ -match '^([0-9a-fA-F]{64})\s+(.+)$') {
    $actual = (Get-FileHash -Algorithm SHA256 $Matches[2]).Hash.ToLowerInvariant()
    if ($actual -ne $Matches[1].ToLowerInvariant()) { throw "Checksum mismatch: $($Matches[2])" }
  }
}
if (Get-Command gh -ErrorAction SilentlyContinue) {
  Get-ChildItem -File | Where-Object { $_.Name -match '\.(zip|tar\.gz)$' } | ForEach-Object { gh attestation verify $_.FullName -R $Repository }
}
Write-Host "release checksums verified"
