$ErrorActionPreference = 'Stop';
$toolsDir               = "$(Split-Path -parent $MyInvocation.MyCommand.Definition)"


$packageArgs = @{
    PackageName    = $env:ChocolateyPackageName
    Destination    = $toolsDir
    FileFullPath   = Join-Path $toolsDir 'ffsclient_32.zip'
    FileFullPath64 = Join-Path $toolsDir 'ffsclient_64.zip'
}

Get-ChocolateyUnzip @packageArgs

Write-Host "ffsclient installed to $toolsDir"

Remove-Item -Force -Path $toolsDir\*.zip