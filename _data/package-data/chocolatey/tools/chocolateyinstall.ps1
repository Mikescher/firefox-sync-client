$ErrorActionPreference = 'Stop';
$toolsDir               = "$(Split-Path -parent $MyInvocation.MyCommand.Definition)"


$packageArgs = @{
    PackageName    = $env:ChocolateyPackageName
    Destination    = $toolsDir
    FileFullPath   = Join-Path $toolsDir 'ffsclient_32.zip'
    FileFullPath64 = Join-Path $toolsDir 'ffsclient_64.zip'
}

#Remove old versions of ripgrep in the tools directory
Get-ChildItem -Directory -Path $toolsDir | Remove-Item -Recurse -Ea 0

Get-ChocolateyUnzip @packageArgs

Write-Host "ripgrep installed to $toolsDir"

Remove-Item -Force -Path $toolsDir\*.zip