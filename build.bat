@echo off
cd /d "%~dp0"
echo Building wxtrans.exe ...
go build -ldflags="-s -w -H windowsgui" -o wxtrans.exe .
if errorlevel 1 (
    echo Build failed.
    exit /b 1
)
for %%A in (wxtrans.exe) do echo Build OK: wxtrans.exe (%%~zA bytes)
