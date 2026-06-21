# 微信小账本构建脚本
$ErrorActionPreference = "Stop"
Set-Location $PSScriptRoot

Write-Host "Building wxtrans.exe ..."
go build -ldflags="-s -w -H windowsgui" -o wxtrans.exe .
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

$size = (Get-Item wxtrans.exe).Length / 1MB
Write-Host ("Done: wxtrans.exe ({0:N1} MB)" -f $size)
